package job

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
)

/**
 * Creates Servers/Pods for lobbys
 */
func Job_Lobby_Server_Provisioner() {
	lobbies, _, err := repository.AllLobby(-1, 1)
	if err != nil {
		log.Printf("Job_Lobby_Server_Provisioner error: %v", err)
		return
	}

	// Limit maximum concurrent goroutines (adjust as needed).
	const maxConcurrent = 100
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := range lobbies {
		lobby := lobbies[i] // capture by value
		parts := strings.Split(lobby.ID.String(), ":")
		if len(parts) < 2 {
			log.Printf("Invalid lobby ID: %v", lobby.ID)
			continue
		}
		label := parts[1]

		// Check if a lock exists.
		if lock, _ := cache.Get[string]("provisioner-lock-" + label); lock != "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func(lobby dbtype.Lobby, label string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore

			// Lock the lobby (30-second duration).
			if err := cache.Set[string]("provisioner-lock-"+label, "true", 30*time.Second); err != nil {
				log.Printf("Error setting lock for lobby %s: %v", label, err)
				return
			}
			defer cache.Del("provisioner-lock-" + label)

			log.Printf("RUN_Lobby_Server_Provisioner: Processing lobby %s", label)

			// Locate existing server pod.
			_, lobpod, _ := k8.LocateServerPod(label)
	
			// If no server pod exists, but a Server record is present, delete it.
			if lobpod == nil && lobby.Lobby_Server != nil {
				log.Printf("RUN_Lobby_Server_Provisioner: Deleting stale server for lobby %s", label)
				if err := repository.DelServer(lobby.Lobby_Server.ID.String()); err != nil {
					log.Printf("Error deleting server for lobby %s: %v", label, err)
				}
				lobby.Lobby_Server = nil
			}

			// If no server is provisioned, and criteria are met, run provisioning.
			if lobby.Lobby_Server == nil && len(lobby.Lobby_Users) >= 0 {
				if err := RUN_Lobby_Server_Provisioner(&lobby, label); err != nil {
					log.Printf("Error provisioning server for lobby %s: %v", label, err)
				}
			}
		}(lobby, label)
	}

	wg.Wait()
}

func RUN_Lobby_Server_Provisioner(lobby *dbtype.Lobby, label string) error {
	log.Printf("RUN_Lobby_Server_Provisioner: Serving lobby %s", label)

	// Check if a server pod already exists.
	_, lobpod, err := k8.LocateServerPod(label)
	if lobpod != nil {
		return fmt.Errorf("pod already exists for lobby %s: phase %s", label, lobpod.Status.Phase)
	}

	// Select an open node and port.
	node, sPort, err := k8.FindOpenNodePort()
	if err != nil {
		return fmt.Errorf("error picking node port for lobby %s: %w", label, err)
	}

	if sPort <= 0 {
		return fmt.Errorf("no open node port for lobby %s: %w", label)
	}

	externalAddress := k8.GetNodeExternalIP(node)

	aPort, err := strconv.ParseInt(lobby.Lobby_Application.Port, 10, 32)
	if err != nil {
		return fmt.Errorf("error parsing application port for lobby %s: %w", label, err)
	}

	// Format the command.
	cmd := strings.Split(lobby.Lobby_Application.Command, " ")

	// Prepare environment variables.
	env := map[string]string{
		"GAME_PORT":         strconv.Itoa(int(aPort)),
		"NODE_HOST":         externalAddress,
		"NODE_PORT":         strconv.Itoa(int(sPort)),
		"VNET_HOST":         os.Getenv("VNET_HOST"),
		"VNET_PORT":         os.Getenv("VNET_PORT"),
		"VNET_KEY":          os.Getenv("VNET_KEY"),
		"LOBBY_MAX_PLAYERS": lobby.Lobby_Application.Lobby_Max_Players,
		"LOBBY_ID":          lobby.ID.String(),
		"LOBBY_NAME":        lobby.Name,
	}

	// Determine the image with version if provided.
	image := lobby.Lobby_Application.Image
	if lobby.Lobby_Application.Version != "" {
		image += ":" + lobby.Lobby_Application.Version
	}

	// Create the server pod.
	lobpod, err = k8.CreateServerPod(label, node, int32(sPort), int32(aPort), image, cmd, env)
	if err != nil {
		return fmt.Errorf("error creating server pod for lobby %s: %w", label, err)
	}

	// Create a server DB entry.
	server, err := repository.PutServer(&dbtype.Server{
		Name:    lobby.Name,
		Guid:    lobpod.Labels["app"],
		Status:  string(lobpod.Status.Phase),
		Address: externalAddress,
		Port:    int32(sPort),
	}, lobby)
	if err != nil {
		k8.DeleteServerPod(label)
		return fmt.Errorf("error creating server DB entry for lobby %s: %w", label, err)
	}

	// Link the lobby with the server.
	if err := repository.LinkLobbyServer(lobby, server); err != nil {
		k8.DeleteServerPod(label)
		repository.DelServer(server.ID.String())
		return fmt.Errorf("error linking lobby and server for lobby %s: %w", label, err)
	}

	// Update Pod Labels.
	if err := k8.AddServerPodLabel(label, map[string]string{
		"lobby":  strings.Split(lobby.ID.String(), ":")[1],
		"server": strings.Split(server.ID.String(), ":")[1],
	}); err != nil {
		log.Printf("Error updating pod labels for lobby %s: %v", label, err)
	}

	log.Printf("RUN_Lobby_Server_Provisioner: Server/Pod created for lobby %s, phase: %s", label, lobpod.Status.Phase)
	return nil
}