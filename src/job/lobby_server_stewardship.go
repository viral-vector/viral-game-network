package job

import (
	"fmt"
	"strings"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/src/service"

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
		node, lobpod, err := k8.LocateServerPod(pod.Name)

		if err != nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}

		if node == nil || lobpod == nil {
			fmt.Errorf("Job_Lobby_Server_Stewardship:  %s", err)
			continue
		}

		fmt.Println("Job_Lobby_Server_Stewardship: @ ", lobpod.Name)
		PrintPod(lobpod)

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

		switch pod.Status.Phase {
		// Check if the pod is running
		case v1.PodRunning:
			Job_Lobby_Server_Stewardship_Running(lobby, lobpod, node)
			break
		// Check if the pod is pending
		case v1.PodPending:
			break
		// Check if the pod is failed
		case v1.PodFailed:
			// Delete Pod
			k8.DeleteServerPod(pod.Name)

			// Delete the server from the lobby
			lobby.Lobby_Server = nil

			// Update the lobby
			_, err := repository.SetLobby(lobby.ID, lobby)
			if err != nil {
				fmt.Println("Job_Lobby_Server_Stewardship:", err)
				continue
			}
			break
		}

		if lobby.Lobby_Server == nil || lobby.Lobby_Server.Status == string(pod.Status.Phase) || lobby.Lobby_Server.Status == "Online" {
			continue
		}

		lobby.Lobby_Server.Status = string(pod.Status.Phase)
		// Update the server
		_, err = repository.SetServer(lobby.Lobby_Server.ID, lobby.Lobby_Server)
		if err != nil {
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

func PrintPod(pod *v1.Pod) {
	fmt.Println("Pod: ", pod.Name)
	fmt.Println("NS: ", pod.Namespace)
	fmt.Println("  Status: ", pod.Status.Phase, pod.Status.Message)
	fmt.Println("  Node: ", pod.Spec.NodeName)
	fmt.Println("  IP: ", pod.Status.PodIP)
	fmt.Println("  Containers:")
	for _, container := range pod.Spec.Containers {
		fmt.Println("    ", container.Name, container.Image)
	}
	// Iterate over the container statuses and print the restart count
	for _, containerStatus := range pod.Status.ContainerStatuses {
		fmt.Println(fmt.Printf("Container Name: %s, Restart Count: %d\n", containerStatus.Name, containerStatus.RestartCount))
	}
}
