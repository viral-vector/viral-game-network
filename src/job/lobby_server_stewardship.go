package job

import (
	"fmt"
	"strings"
	"viral-game-network/src/k8"
	"viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/service"
	"k8s.io/api/core/v1"
)

func Job_Lobby_Server_Stewardship() {
	fmt.Println("Job_Lobby_Server_Stewardship")

	pods, err := k8.GetAllServerPods()
	if err != nil {
		fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
		return
	}

	for _, pod := range pods {
		node, lobpod, err := k8.LocateServerPod(pod.Name)

		if err!= nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}

		if node == nil || lobpod == nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}

		fmt.Println("Job_Lobby_Server_Stewardship: @",  lobpod.Name)

		// Get the lobby
		lobby, err := repository.GetLobby(
			strings.Replace(pod.Name, "server-", "Lobby:", -1),
		)
		if err != nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}
		if lobby == nil {
			// Delete Pod
			k8.DeleteServerPod(pod.Name)

			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", "Lobby not found")
			continue
		}

		if lobby.Lobby_Server.Status == string(pod.Status.Phase) {
			continue
		}

		switch pod.Status.Phase {
			// Check if the pod is running
			case v1.PodRunning: 
				Job_Lobby_Server_Stewardship_Running(lobby, lobpod, node)
				break;
			// Check if the pod is pending
			case v1.PodPending:
				break;
			// Check if the pod is failed
			case v1.PodFailed:
				// Delete Pod
				k8.DeleteServerPod(pod.Name)

				// Delete the server from the lobby
				lobby.Lobby_Server = nil

				// Update the lobby
				_, err := repository.SetLobby(lobby.ID, lobby)
				if err!= nil {
					fmt.Println("Job_Lobby_Server_Stewardship:", err)
					continue
				}
				break;
		}

		lobby.Lobby_Server.Status = string(pod.Status.Phase)
		// Update the server
		_, err = repository.SetServer(lobby.Lobby_Server.ID, lobby.Lobby_Server)
		if err!= nil {
			fmt.Println("Job_Lobby_Server_Stewardship:", err)
			continue
		}

		// Notify Lobby Users
		msg := "Server:" + lobby.Lobby_Server.Status
		service.Service_Lobby_Notify(lobby.ID, msg, "")
	}
}

// Job_Lobby_Server_Stewardship_Running
func Job_Lobby_Server_Stewardship_Running(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) {
	
	// Find the external IP address
	var external_address string = "https://localhost"

	for _, address := range node.Status.Addresses {
		if address.Type == "ExternalIP" {
			external_address = address.Address
		}
	}
	if external_address == "" {
		fmt.Println("Job_Lobby_Server_Stewardship: No ExternalIP", node.Name)
		return
	}
	
	// Update the server address
	lobby.Lobby_Server.Address = external_address
}

