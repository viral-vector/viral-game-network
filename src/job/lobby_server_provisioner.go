package job

import (
	"fmt"
	"time"
	"strings"
	"viral-game-network/src/k8"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
)

func Job_Lobby_Server_Provisioner() {
	lobbies, err := repository.AllLobby()
	if err!= nil {
		fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	for _, lobby := range lobbies {
		// Get the label for the lobby
		label := strings.Split(lobby.ID,":")[1]

		// Do we need a new server?
		if lobby.Lobby_Server == nil && len(lobby.Lobby_Users) >= 0 {
			// Check if we have a k8-Lock on the lobby
			if cache_lock, _ := cache.Get("k8-lock-" +label); cache_lock != nil {	
				continue
			}
			go func (lobby dbtype.Lobby, label string) {
				// k8-Lock the lobby
				cache.Set("k8-lock-" +label, "true", time.Minute * 10)
				// Create the server pod
				Lobby_Server_Provisioner_PUT(lobby, label)
				// k8-Unlock the lobby
				cache.Del("k8-lock-" +label)
			}(lobby, label) 
		}
	}
}

// Job_Lobby_Server_Provisioner_PUT
func Lobby_Server_Provisioner_PUT(lobby dbtype.Lobby, label string) error {
	// Get the server pod
	lobpod, err := k8.LocateServerPod(label)
	if lobpod != nil {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}
	// Create the server pod
	lobpod, err = k8.CreateServerPod(label)
	if err != nil {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	// Create A Server Db Entry 
	server, err := repository.PutServer(dbtype.Server{
		Name: lobby.Name,
		Guid: lobpod.Labels["app"],
		Address: "http://" + lobpod.Status.String(),
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

	fmt.Println("Server Provisioned for Lobby: ", label, " ", server.Name, " ", server.Address)

	return nil
}