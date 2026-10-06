package job

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	pods, _, err := k8.GetAllServerPodsList("", -1, "")
	if err != nil {
		log.Printf("[Job_Lobby_Server_Stewardship]ERROR: %v", err)
		return
	}

	// Use a semaphore to limit concurrent goroutines.
	const maxConcurrent = 100 // Adjust as needed based on resources.
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, pod := range pods {
		// Capture the pod value to avoid closure issues.
		pod := pod

		wg.Add(1)
		sem <- struct{}{} // acquire a semaphore slot

		go func(pod *v1.Pod, podName string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore when done
			_, lockErr := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+podName, 60*time.Second, func(ctx context.Context) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				// Run the main stewardship process.
				runServerStewardship(ctx, pod)
				return nil
			})
			if lockErr != nil {
				log.Printf("lobby lifecycle: %v", lockErr)
			}
		}(pod, pod.Name)
	}

	wg.Wait()
}

// RUN_Lobby_Server_Stewardship processes a single pod.
func RUN_Lobby_Server_Stewardship(pod *v1.Pod) {
	runServerStewardship(context.Background(), pod)
}

func runServerStewardship(ctx context.Context, pod *v1.Pod) {
	if pod == nil {
		return
	}
	lobpod, err := k8.GetServerPod(pod.Name)
	if err != nil {
		log.Printf("[Job_Lobby_Server_Stewardship]ERROR: reading pod %s: %v", pod.Name, err)
		return
	}
	if lobpod.DeletionTimestamp != nil {
		return
	}
	lobby, err := repository.GetLobby("Lobby:" + lobpod.Name)
	if err != nil {
		log.Printf("[Job_Lobby_Server_Stewardship]ERROR: fetching lobby for %s: %v", lobpod.Name, err)
		return
	}
	if ctx.Err() != nil {
		return
	}
	// Pending and terminal pods can be cleaned up before a node is assigned.
	if lobby == nil || lobby.Lobby_Server == nil {
		if err := purgeServer(ctx, lobby, lobpod); err != nil {
			log.Printf("[Job_Lobby_Server_Stewardship]ERROR: purging orphan %s: %v", lobpod.Name, err)
		}
		return
	}
	if lobpod.Status.Phase == v1.PodFailed || lobpod.Status.Phase == v1.PodSucceeded {
		if err := purgeServer(ctx, lobby, lobpod); err != nil {
			log.Printf("[Job_Lobby_Server_Stewardship]ERROR: purging pod %s: %v", lobpod.Name, err)
			return
		}
		ministration.Service_Lobby_Notify(lobby.ModelID(), "Server:"+string(lobpod.Status.Phase), "")
		return
	}
	server := lobby.Lobby_Server
	previousStatus, previousAddress, previousPort := server.Status, server.Address, server.Port
	status := string(lobpod.Status.Phase)
	if lobpod.Status.Phase == v1.PodRunning {
		node, err := k8.GetNode(lobpod.Spec.NodeName)
		if err != nil {
			log.Printf("[Job_Lobby_Server_Stewardship]ERROR: reading node for %s: %v", lobpod.Name, err)
			return
		}
		RUN_Lobby_Server_Stewardship_Running(lobby, lobpod, node)
		// A heartbeat confirms application readiness beyond the Kubernetes phase.
		if previousStatus == "Online" {
			status = "Online"
		}
	}
	server.Status = status
	if server.Status == previousStatus && server.Address == previousAddress && server.Port == previousPort {
		return
	}
	if ctx.Err() != nil {
		return
	}
	updated, err := repository.SyncServerPodState(server.ModelID(), string(lobpod.Status.Phase), server.Address, server.Port)
	if err != nil {
		log.Printf("[Job_Lobby_Server_Stewardship]ERROR: updating server for %s: %v", lobpod.Name, err)
		return
	}
	ministration.Service_Lobby_Notify(lobby.ModelID(), "Server:"+updated.Status, "")
}

// RUN_Lobby_Server_Stewardship_Running updates server info when pod is running.
func RUN_Lobby_Server_Stewardship_Running(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) {
	lobby.Lobby_Server.Port = k8.GetPodHostPort(pod)
	lobby.Lobby_Server.Address = k8.GetNodeExternalIP(node)
}

// RUN_Lobby_Server_Stewardship_Purger cleans up a server pod and server record.
func RUN_Lobby_Server_Stewardship_Purger(lobby *dbtype.Lobby, pod *v1.Pod, node *v1.Node) error {
	return purgeServer(context.Background(), lobby, pod)
}

func purgeServer(ctx context.Context, lobby *dbtype.Lobby, pod *v1.Pod) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if pod == nil {
		return fmt.Errorf("missing server pod")
	}
	if err := k8.DeleteServerPod(pod.Name); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	server, err := repository.GetServerByGuid(pod.Name)
	if err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if server != nil {
		if server.ID == nil {
			return fmt.Errorf("missing server record ID")
		}
		if err := repository.DelServer(server.ModelID(), server); err != nil {
			return err
		}
	}
	if lobby != nil {
		lobby.Lobby_Server = nil
	}
	return nil
}
