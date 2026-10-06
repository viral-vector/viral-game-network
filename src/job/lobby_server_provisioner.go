package job

import (
	"context"
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"viral-game-network/src/auth"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
)

/**
 * Creates Servers/Pods for lobbys
 */
func Job_Lobby_Server_Provisioner() {
	lobbies, _, err := repository.AllLobbyNotRunning(-1, 1)
	if err != nil {
		log.Printf("[Job_Lobby_Server_Provisioner]: %v", err)
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
			log.Printf("[Job_Lobby_Server_Provisioner]ERROR: Invalid lobby ID: %v", lobby.ID)
			continue
		}
		label := parts[1]

		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func(lobby dbtype.Lobby, label string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore

			_, lockErr := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+label, 60*time.Second, func(ctx context.Context) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				fresh, err := repository.GetLobby(lobby.ModelID())
				if err != nil {
					return err
				}
				if fresh == nil {
					return nil
				}
				lobby = *fresh
				// Locate existing server pod.
				lobpod, podErr := k8.GetServerPod(label)
				if podErr != nil && !apierrors.IsNotFound(podErr) {
					log.Printf("[Job_Lobby_Server_Provisioner]ERROR: reading pod %s: %v", label, podErr)
					return nil
				}

				if ctx.Err() != nil {
					return ctx.Err()
				}
				// If no server pod exists, but a Server record is present, delete it.
				if lobpod == nil && lobby.Lobby_Server != nil {
					log.Printf("[Job_Lobby_Server_Provisioner]: Deleting lobby %s server", label)
					if err := repository.DelServer(lobby.Lobby_Server.ID.String(), lobby.Lobby_Server); err != nil {
						log.Printf("[Job_Lobby_Server_Provisioner]ERROR: deleting lobby %s server: %v", label, err)
						return nil
					}
					lobby.Lobby_Server = nil
				}

				// If no server is provisioned, and criteria are met, run provisioning.
				if lobby.Lobby_Server == nil {
					if err := provisionWithLock(ctx, &lobby, label); err == nil {
						log.Printf("[Job_Lobby_Server_Provisioner]: Server created for lobby %s", label)
					} else {
						return err
					}
				}
				return nil
			})
			if lockErr != nil {
				log.Printf("lobby lifecycle: %v", lockErr)
			}
		}(lobby, label)
	}

	wg.Wait()
}

func RUN_Lobby_Server_Provisioner(lobby *dbtype.Lobby, label string) error {
	return provisionWithLock(context.Background(), lobby, label)
}

func provisionWithLock(parent context.Context, lobby *dbtype.Lobby, label string) error {
	held, err := cache.WithLock(parent, "server-port-allocation-lock", 60*time.Second, func(ctx context.Context) error { return provisionLobbyServer(ctx, lobby, label) })
	if err != nil {
		return err
	}
	if !held {
		return errors.New("server port allocation busy; retry provisioning")
	}
	return nil
}

func provisionLobbyServer(ctx context.Context, lobby *dbtype.Lobby, label string) error {
	if lobby == nil || lobby.ID == nil || (lobby.Lobby_Application == nil || lobby.Lobby_Application.ID == nil) {
		return errors.New("missing lobby or application")
	}
	log.Printf("[RUN_Lobby_Server_Provisioner]: Serving lobby %s", label)

	// Check if a server pod already exists.
	lobpod, err := k8.GetServerPod(label)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if lobpod != nil {
		log.Printf("[Job_Lobby_Server_Provisioner]: pod already exists for lobby %s: phase %s", label, lobpod.Status.Phase)
		return nil
	}

	// Select an open node and port.
	node, sPort, err := k8.FindOpenNodePort()
	if err != nil {
		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: picking node port for lobby %s: %v", label, err)
		return err
	}

	if sPort <= 0 {
		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: no open node port for lobby %s", label)
		return errors.New("no open node port")
	}

	externalAddress := k8.GetNodeExternalIP(node)

	aPort, err := strconv.ParseInt(lobby.Lobby_Application.Port, 10, 32)
	if err != nil {
		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: parsing application port for lobby %s: %v", label, err)
		return err
	}

	// Format the command.
	cmd := strings.Split(lobby.Lobby_Application.Command, " ")

	serverToken, err := auth.GenerateToken("server:"+label, lobby.ModelID(), auth.ServerAudience)
	if err != nil {
		return err
	}
	// Prepare environment variables.
	env := map[string]string{
		"GAME_PORT":         strconv.Itoa(int(aPort)),
		"NODE_HOST":         externalAddress,
		"NODE_PORT":         strconv.Itoa(int(sPort)),
		"VNET_HOST":         os.Getenv("VNET_HOST"),
		"VNET_PORT":         os.Getenv("VNET_PORT"),
		"VNET_KEY":          os.Getenv("VNET_KEY"),
		"VNET_SERVER_TOKEN": serverToken,
		"LOBBY_MAX_PLAYERS": lobby.Lobby_Application.Lobby_Max_Players,
		"LOBBY_ID":          lobby.ID.String(),
		"LOBBY_NAME":        lobby.Name,
	}

	// Determine the image with version if provided.
	image := lobby.Lobby_Application.Image
	if lobby.Lobby_Application.Version != "" {
		image += ":" + lobby.Lobby_Application.Version
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Create the server pod.
	lobpod, err = k8.CreateServerPod(label, node, int32(sPort), int32(aPort), image, cmd, env)
	if err != nil {
		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: creating server pod for lobby %s: %v", label, err)
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
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

		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: creating server DB entry for lobby %s: %v", label, err)
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Link the lobby with the server.
	if err := repository.LinkLobbyServer(lobby, server); err != nil {
		k8.DeleteServerPod(label)
		repository.DelServer(server.ID.String(), server)

		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: linking lobby and server for lobby %s: %v", label, err)
		return err
	}

	// Update Pod Labels.
	if err := k8.AddServerPodLabel(label, map[string]string{
		"lobby":  strings.Split(lobby.ID.String(), ":")[1],
		"server": strings.Split(server.ID.String(), ":")[1],
	}); err != nil {
		log.Printf("[Job_Lobby_Server_Provisioner]ERROR: updating pod labels for lobby %s: %v", label, err)
	}
	return nil
}
