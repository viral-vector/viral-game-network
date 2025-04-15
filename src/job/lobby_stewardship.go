package job

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/ministration"
)

/**
 * Manages Lobby Lifecycle
 */
func Job_Lobby_Stewardship() {
	lobbies, _, err := repository.AllLobby(-1, 1)
	if err != nil {
		log.Printf("Job_Lobby_Stewardship Error: %v", err)
		return
	}

	// Limit concurrency (e.g. max 100 goroutines concurrently)
	const maxConcurrent = 100
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := range lobbies {
		lobby := lobbies[i] // capture by value
		parts := strings.Split(lobby.ID.String(), ":")
		if len(parts) < 2 {
			log.Printf("Invalid lobby ID format: %v", lobby.ID)
			continue
		}
		label := parts[1]

		// Skip if there's already a lock for this lobby.
		if lock, _ := cache.Get[string]("lobby-stewardship-lock-" + label); lock != "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // acquire a semaphore slot

		go func(lobby dbtype.Lobby, label string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore when done

			// Lock the lobby for stewardship.
			if err := cache.Set[string]("lobby-stewardship-lock-"+label, "true", 30*time.Second); err != nil {
				log.Printf("Error setting lock for lobby %s: %v", label, err)
				return
			}
			// Ensure the lock is removed regardless of errors.
			defer cache.Del("lobby-stewardship-lock-" + label)

			log.Printf("RUN_Lobby_Stewardship: Processing lobby %s", lobby.ID.String())

			// Run the stewardship logic.
			RUN_Lobby_Stewardship(&lobby)
		}(lobby, label)
	}

	wg.Wait()
}

// RUN_Lobby_Stewardship processes a single lobby to check its persistence and potentially purge it.
func RUN_Lobby_Stewardship(lobby *dbtype.Lobby) {
	log.Printf("[RUN_Lobby_Stewardship] Evaluating lobby %s", lobby.ID.String())

	// Parse the maximum persistence from the lobby configuration.
	lobbyMaxPersist, err := strconv.ParseFloat(lobby.Lobby_Application.Lobby_Max_Persist, 64)
	if err != nil {
		log.Printf("[RUN_Lobby_Stewardship] Error parsing Lobby_Max_Persist %v", err)
		return
	}

	now := time.Now().UTC()
	// Use Date_Updated if available; otherwise, fall back to Date_Created.
	dateStr := lobby.Date_Updated
	if dateStr == "" {
		dateStr = lobby.Date_Created
	}
	dateLobby, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		log.Printf("[RUN_Lobby_Stewardship] Error parsing date for lobby %s: %v", lobby.ID.String(), err)
		return
	}

	diffMinutes := now.Sub(dateLobby).Minutes()
	log.Printf("Lobby %s has persisted for %.2f minutes", lobby.ID.String(), diffMinutes)

	// If the lobby has been active longer than allowed, purge it.
	if diffMinutes > lobbyMaxPersist {
		// Notify lobby users.
		msg := "Server:Lobby Closed"
		ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")
		log.Printf("RUN_Lobby_Stewardship: Closing lobby %s", lobby.ID.String())

		// Delete the associated server, if it exists.
		if lobby.Lobby_Server != nil {
			if err := repository.DelServer(lobby.Lobby_Server.ID.String()); err != nil {
				log.Printf("Error deleting server for lobby %s: %v", lobby.ID.String(), err)
			}
		}

		// Delete the lobby.
		if err := repository.DelLobby(lobby.ID.String(), lobby); err != nil {
			log.Printf("Error deleting lobby %s: %v", lobby.ID.String(), err)
		}
	}
}