package job

import (
	"fmt"
	"time"
	"slices"
	"strings"
	"viral-game-network/src/k8"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
)

func Job_Lobby_Server_Provisioner() {
	fmt.Println("Job_Lobby_Server_Provisioner")

	
	lobbies, err := repository.AllLobby()
	if err != nil {
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
				cache.Set("k8-lock-" +label, "true", time.Minute * 1)
				// Create the server pod
				err := Lobby_Server_Provisioner_PUT(lobby, label)
				if err != nil {
					fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
				}
				// k8-Unlock the lobby
				cache.Del("k8-lock-" +label)
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

	// Pick Open Port
	port_null := repository.GetServerPorts()
	port_open := int32(-1)
	for i :=  30000; i <=  30500; i++ {
        if ok := slices.Contains(port_null, int32(i)); ok == false {
			port_open = int32(i)
			break
		}
    }
	if port_open == int32(-1) {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", "No open ports")
	}

	// Create the server pod
	lobpod, service, err := k8.CreateServerPod(label, port_open)
	if err != nil {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	// Create A Server Db Entry
	external_address := ""

	server, err := repository.PutServer(dbtype.Server{
		Name: lobby.Name,
		Guid: lobpod.Labels["app"],
		Status: string(lobpod.Status.Phase),
		Address: external_address,
		Port: int32(service.Spec.Ports[0].NodePort),
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