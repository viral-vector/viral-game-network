package job

import (
	"fmt"
	"strings"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/src/ministration"

	v1 "k8s.io/api/core/v1"
)

/**
 * Manages Server Lifecycle
 */
func Job_Lobby_Server_Stewardship() {
	// fmt.Println("Job_Lobby_Server_Stewardship")

	pods, err := k8.GetAllServerPods()
	if err != nil {
		fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
		return
	}

	for _, pod := range pods {
		// Get the label for the lobby
		label := strings.Split(pod.Name, "-")[1]

		node, lobpod, err := k8.LocateServerPod(label)

		if err != nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}

		if node == nil || lobpod == nil{
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}

		fmt.Println("Job_Lobby_Server_Stewardship: @ ", label)
		
		// Get the lobby
		lobby, err := repository.GetLobby(
			strings.Replace(lobpod.Name, "server-", "Lobby:", -1),
		)
		if err != nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}
		if lobby == nil {
			// Delete Pod
			k8.DeleteServerPod(lobpod.Name)

			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", "Lobby not found")
			continue
		}

		switch lobpod.Status.Phase {
		// Check if the pod is running
		case v1.PodRunning:
			Job_Lobby_Server_Stewardship_Running(lobby, lobpod, node)
		// Check if the pod is pending
		case v1.PodPending:
		// Check if the pod is failed
		case v1.PodFailed:
			// Delete Pod
			k8.DeleteServerPod(lobpod.Name)

			// Delete the server from the lobby
			lobby.Lobby_Server = nil

			// Update the lobby
			_, err := repository.SetLobby(lobby.ID.String(), lobby)
			if err != nil {
				fmt.Println("Job_Lobby_Server_Stewardship:", err)
				continue
			}
		}

		if lobby.Lobby_Server == nil || lobby.Lobby_Server.Status == string(lobpod.Status.Phase) || lobby.Lobby_Server.Status == "Online" {
			continue
		}

		lobby.Lobby_Server.Status = string(lobpod.Status.Phase)
		// Update the server
		_, err = repository.SetServer(lobby.Lobby_Server.ID.String(), lobby.Lobby_Server)
		if err != nil {
			fmt.Println("Job_Lobby_Server_Stewardship:", err)
			continue
		}

		// Notify Lobby Users
		msg := "Server:" + lobby.Lobby_Server.Status
		ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")
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
