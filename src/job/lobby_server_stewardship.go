package job

import (
	"log"
	"sync"
	"time"

	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/src/ministration"

	v1 "k8s.io/api/core/v1"
)

/**
 * Manages Server/Pod Lifecycle for lobbies
 */
func Job_Lobby_Server_Stewardship() {
	// Retrieve all server pods.
	pods, err := k8.GetAllServerPods()
	if err != nil {
		log.Panicf("[Job_Lobby_Server_Stewardship]: %v", err)
		return
	}

	// Use a semaphore to limit concurrent goroutines.
	const maxConcurrent = 100 // Adjust as needed based on resources.
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, pod := range pods {
		// Capture the pod value to avoid closure issues.
		pod := pod

		// Check if there's already a lock for this pod.
		if lock, _ := cache.Get[string]("server-stewardship-lock-" + pod.Name); lock != "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // acquire a semaphore slot

		go func(pod *v1.Pod, podName string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore when done
			// Lock the pod for stewardship.
			if err := cache.Set[string]("server-stewardship-lock-"+podName, "true", 30*time.Second); err != nil {
				log.Panicf("[Job_Lobby_Server_Stewardship]: Error locking pod %s: %v", podName, err)
				return
			}
			defer cache.Del("server-stewardship-lock-" + podName)
			// Run the main stewardship process.
			RUN_Lobby_Server_Stewardship(pod)
		}(pod, pod.Name)
	}

	wg.Wait()
}

// RUN_Lobby_Server_Stewardship processes a single pod.
func RUN_Lobby_Server_Stewardship(pod *v1.Pod) {
	node, lobpod, err := k8.LocateServerPod(pod.Name)
	if err != nil {
		log.Panicf("[Job_Lobby_Server_Stewardship]: Error locating server pod for %s: %v", pod.Name, err)
		return
	}
	if node == nil || lobpod == nil {
		log.Printf("[Job_Lobby_Server_Stewardship]: Unable to locate node or pod for %s", pod.Name)
		return
	}

	// Get the corresponding lobby.
	lobby, err := repository.GetLobby("Lobby:" + lobpod.Name)
	if err != nil {
		log.Panicf("[Job_Lobby_Server_Stewardship]: Error fetching lobby for %s: %v", lobpod.Name, err)
		return
	}

	// If the pod is terminating, skip further processing.
	if pod.DeletionTimestamp != nil {
		log.Printf("[Job_Lobby_Server_Stewardship]: Pod %s is terminating", lobpod.Name)
		return
	}

	// If no lobby or no server recorded in the lobby, purge the stale pod/server.
	if lobby == nil || lobby.Lobby_Server == nil {
		RUN_Lobby_Server_Stewardship_Purger(lobby, lobpod, node)
		log.Printf("[Job_Lobby_Server_Stewardship]: Lobby/Server not found for pod %s", lobpod.Name)
		return
	}

	// Process based on pod phase.
	switch lobpod.Status.Phase {
	case v1.PodRunning:
		RUN_Lobby_Server_Stewardship_Running(lobby, lobpod, node)
	case v1.PodFailed:
		RUN_Lobby_Server_Stewardship_Purger(lobby, lobpod, node)
	}

	// Update the server status if needed.
	if lobby.Lobby_Server != nil &&
		(lobby.Lobby_Server.Status != string(lobpod.Status.Phase) && lobby.Lobby_Server.Status != "Online") {

		lobby.Lobby_Server.Status = string(lobpod.Status.Phase)

		if _, err := repository.SetServer(lobby.Lobby_Server.ID.String(), lobby.Lobby_Server); err != nil {
			log.Panicf("[Job_Lobby_Server_Stewardship]: Error updating server for lobby %s: %v", lobpod.Name, err)
			return
		}
	}

	// Notify lobby users about the current server status.
	msg := "Server:" + lobby.Lobby_Server.Status
	ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")
	log.Printf("[Job_Lobby_Server_Stewardship]: Server/Pod processed for lobby %s; phase: %s", lobpod.Name, lobpod.Status.Phase)
}

// RUN_Lobby_Server_Stewardship_Running updates server info when pod is running.
func RUN_Lobby_Server_Stewardship_Running(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) {
	lobby.Lobby_Server.Port = k8.GetPodHostPort(pod)
	lobby.Lobby_Server.Address = k8.GetNodeExternalIP(node)
}

// RUN_Lobby_Server_Stewardship_Purger cleans up a server pod and server record.
func RUN_Lobby_Server_Stewardship_Purger(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) error {
	// Delete the pod.
	k8.DeleteServerPod(pod.Name)

	// Determine the server ID to delete.
	var serverID string
	if lobby != nil && lobby.Lobby_Server != nil {
		serverID = lobby.Lobby_Server.ID.String()
		lobby.Lobby_Server = nil
	} else {
		server, _ := repository.GetServerByGuid(pod.Name)
		if server != nil {
			serverID = server.ID.String()
		}
	}
	if serverID != "" {
		repository.DelServer(serverID)
	}
	return nil
}
