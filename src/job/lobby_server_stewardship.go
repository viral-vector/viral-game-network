package job

import (
	"fmt"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/src/ministration"

	v1 "k8s.io/api/core/v1"
)

/**
 * Manages Server/Pod Lifecycle
 */
func Job_Lobby_Server_Stewardship() {
	pods, err := k8.GetAllServerPods()
	if err != nil {
		fmt.Println(fmt.Errorf("RUN_Lobby_Server_Stewardship: %s", err))
		return
	}
	for _, pod := range pods {		
		// Check if we have a lock on the pod
		if cache_lock, _ := cache.Get[string]("server-stewardship-lock-" + pod.Name); cache_lock != "" {
			continue
		}
		go func(pod *v1.Pod, podName string) {
			// Lock the pod
			cache.Set("server-stewardship-lock-" + podName, "true", time.Second*30)

			fmt.Println("RUN_Lobby_Server_Stewardship: @ Looking", pod.Name)

			// RUN
			RUN_Lobby_Server_Stewardship(pod)
			// Unlock the pod
			cache.Del("server-stewardship-lock-" + podName)
		}(pod, pod.Name)
	}
}

func RUN_Lobby_Server_Stewardship(pod *v1.Pod) {
	node, lobpod, err := k8.LocateServerPod(pod.Name)

	if err != nil {
		fmt.Println(fmt.Errorf("RUN_Lobby_Server_Stewardship: %s", err))
		return
	}

	if node == nil || lobpod == nil{
		fmt.Println(fmt.Errorf("RUN_Lobby_Server_Stewardship: %s", err))
		return
	}
	
	// Get the lobby
	lobby, err := repository.GetLobby("Lobby:" + lobpod.Name)
	if err != nil {
		fmt.Println(fmt.Errorf("RUN_Lobby_Server_Stewardship: @ Lobby Fetch Error %s", err))
		return
	}

	if pod.DeletionTimestamp != nil {
		fmt.Println(fmt.Errorf("RUN_Lobby_Server_Stewardship: Terminating %s", lobpod.Name))
		return
	}

	if (lobby == nil || lobby.Lobby_Server == nil) {
		// Purge Pod & Server
		RUN_Lobby_Server_Stewardship_Purger(lobby, lobpod, node)

		fmt.Println(fmt.Errorf("RUN_Lobby_Server_Stewardship: Lobby/Server not found %s", lobpod.Name))
		return
	}

	switch lobpod.Status.Phase {
	// Check if the pod is running
	case v1.PodRunning:
		RUN_Lobby_Server_Stewardship_Running(lobby, lobpod, node)
	// Check if the pod is pending
	case v1.PodPending:
	// Check if the pod is failed
	case v1.PodFailed:
		// Purge Pod & Server
		RUN_Lobby_Server_Stewardship_Purger(lobby, lobpod, node)
	}

	if lobby.Lobby_Server == nil || lobby.Lobby_Server.Status == string(lobpod.Status.Phase) || lobby.Lobby_Server.Status == "Online" {
		return
	}

	lobby.Lobby_Server.Status = string(lobpod.Status.Phase)
	// Update the server
	_, err = repository.SetServer(lobby.Lobby_Server.ID.String(), lobby.Lobby_Server)
	if err != nil {
		fmt.Println("RUN_Lobby_Server_Stewardship: @ Error Updating", err)
		return
	}

	// Notify Lobby Users
	msg := "Server:" + lobby.Lobby_Server.Status
	ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")
}

func RUN_Lobby_Server_Stewardship_Running(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) {
	// Update the server address
	lobby.Lobby_Server.Port = k8.GetPodHostPort(pod)
	lobby.Lobby_Server.Address = k8.GetNodeExternalIP(node)
}

// Purge a Server & Pod
func RUN_Lobby_Server_Stewardship_Purger(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) error {	
	// Delete Pod
	k8.DeleteServerPod(pod.Name)

	// Delete Server
	serverID := ""
	if lobby != nil && lobby.Lobby_Server != nil {
		serverID = lobby.Lobby_Server.ID.String()
		lobby.Lobby_Server = nil
	}else {
		server, _ := repository.GetServerByGuid(pod.Name)
		if server != nil{
			serverID = server.ID.String()
		}
	}
	if len(serverID) > 0 {
		repository.DelServer(serverID)
	}
	return nil
}
