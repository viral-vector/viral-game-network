package job

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
)

/**
 * Creates Servers for lobbys
 */
func Job_Lobby_Server_Provisioner() {
	// fmt.Println("Job_Lobby_Server_Provisioner")

	lobbies, _, err := repository.AllLobby(-1, 1)
	if err != nil {
		fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	for _, lobby := range lobbies {
		// Get the label for the lobby
		label := strings.Split(lobby.ID, ":")[1]

		// If the server pod exists, we are done
		// Get the server pod
		fmt.Println("Job_Lobby_Server_Provisioner: @ Looking ", label)
		_, lobpod, _ := k8.LocateServerPod(label)
		if lobpod == nil && lobby.Lobby_Server != nil {
			repository.DelServer(lobby.Lobby_Server.ID)
			lobby.Lobby_Server = nil
		}

		// Do we need a new server?
		if lobby.Lobby_Server == nil && len(lobby.Lobby_Users) >= 0 {
			// Check if we have a k8-Lock on the lobby
			if cache_lock, _ := cache.Get("k8-lock-" + label); cache_lock != nil {
				continue
			}
			go func(lobby dbtype.Lobby, label string) {
				// k8-Lock the lobby
				cache.Set("k8-lock-"+label, "true", time.Minute*1)
				// Create the server pod
				err := Lobby_Server_Provisioner_PUT(lobby, label)
				if err != nil {
					fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
				}
				// k8-Unlock the lobby
				cache.Del("k8-lock-" + label)
			}(lobby, label)
		}
	}
}

// Job_Lobby_Server_Provisioner_PUT
func Lobby_Server_Provisioner_PUT(lobby dbtype.Lobby, label string) error {
	// Get the server pod
	_, lobpod, err := k8.LocateServerPod(label)
	if lobpod != nil {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	return nil

	// Pick Open Node/Port TODO: get port to pass to createserverpod
	node, _, err := k8.FindOpenNodePort()

	// Format the command
	cmd := []string{}
	for _, str := range strings.Split(os.Getenv("GAME_COMMANDS"), " ") {
		str = strings.Replace(str, "${GAME_NAME}", os.Getenv("GAME_NAME"), -1)
		str = strings.Replace(str, "${GAME_HOST}", os.Getenv("GAME_HOST"), -1)
		str = strings.Replace(str, "${GAME_PORT}", os.Getenv("GAME_PORT"), -1)

		str = strings.Replace(str, "${VNET_HOST}", os.Getenv("VNET_HOST"), -1)
		str = strings.Replace(str, "${VNET_PORT}", os.Getenv("VNET_PORT"), -1)
		str = strings.Replace(str, "${VNET_KEY}", os.Getenv("VNET_KEY"), -1)

		str = strings.Replace(str, "${LOBBY_MAX_PLAYERS}", os.Getenv("LOBBY_MAX_PLAYERS"), -1)
		str = strings.Replace(str, "${LOBBY_ID}", lobby.ID, -1)
		str = strings.Replace(str, "${LOBBY_NAME}", lobby.Name, -1)
		cmd = append(cmd, str)
	}

	aport, err := strconv.ParseInt(os.Getenv("GAME_PORT"), 10, 32)
	if err != nil {
		return fmt.Errorf("error converting string to int: %s", err)
	}

	// Create the server pod
	// TODO: use port from FindOpenNodePort
	lobpod, service, err := k8.CreateServerPod(label, node, int32(0), int32(aport), os.Getenv("GAME_DOKIMAGE"), cmd)
	if err != nil {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}
	fmt.Println("Job_Lobby_Server_Provisioner: Server Created : ", lobpod.Status.Phase)

	// Create A Server Db Entry
	external_address := ""

	server, err := repository.PutServer(dbtype.Server{
		Name:    lobby.Name,
		Guid:    lobpod.Labels["app"],
		Status:  string(lobpod.Status.Phase),
		Address: external_address,
		Port:    int32(service.Spec.Ports[0].NodePort),
	}, &lobby)
	if err != nil {
		// Delete the server pod if we failed to create the server entry
		k8.DeleteServerPod(label)

		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	// Link Lobby & Server
	err = repository.LinkLobbyServer(lobby.ID, server)
	if err != nil {
		// Delete the server pod if we failed to create the server entry
		k8.DeleteServerPod(label)
		// Delete the server entry if we failed to create the server entry
		repository.DelServer(server.ID)
	}

	return nil
}
