package job

import (
	"context"
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
	lobbies, _, err := repository.AllLobby(-1, 1, "")
	if err != nil {
		log.Printf("[Job_Lobby_Stewardship]ERROR: %v", err)
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
			log.Printf("[Job_Lobby_Stewardship]ERROR: Invalid lobby ID format: %v", lobby.ID)
			continue
		}
		label := parts[1]

		wg.Add(1)
		sem <- struct{}{} // acquire a semaphore slot

		go func(lobby dbtype.Lobby, label string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore when done

			_, lockErr := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+label, 60*time.Second, func(ctx context.Context) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				// Run the stewardship logic.
				fresh, err := repository.GetLobby(lobby.ModelID())
				if err != nil {
					return err
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}
				runLobbyStewardship(ctx, fresh)
				return nil
			})
			if lockErr != nil {
				log.Printf("lobby lifecycle: %v", lockErr)
			}
		}(lobby, label)
	}

	wg.Wait()
}

// RUN_Lobby_Stewardship processes a single lobby to check its persistence and potentially purge it.
func RUN_Lobby_Stewardship(lobby *dbtype.Lobby) {
	runLobbyStewardship(context.Background(), lobby)
}

func runLobbyStewardship(ctx context.Context, lobby *dbtype.Lobby) {
	if lobby == nil || lobby.ID == nil || (lobby.Lobby_Application == nil || lobby.Lobby_Application.ID == nil) {
		log.Println("[Job_Lobby_Stewardship]ERROR: missing lobby or application")
		return
	}
	log.Printf("[Job_Lobby_Stewardship]: Evaluating lobby %s", lobby.ID.String())

	// Parse the max persist from the application config.
	lobbyMaxPersist, err := strconv.ParseFloat(lobby.Lobby_Application.Lobby_Max_Persist, 64)
	if err != nil {
		log.Printf("[Job_Lobby_Stewardship]ERROR: parsing Lobby_Max_Persist %v", err)
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
		log.Printf("[Job_Lobby_Stewardship]ERROR: parsing date for lobby %s: %v", lobby.ID.String(), err)
		return
	}

	diffMinutes := now.Sub(dateLobby).Minutes()
	log.Printf("[Job_Lobby_Stewardship]: Lobby %s alive %.2f minutes", lobby.ID.String(), diffMinutes)

	// If the lobby has been active longer than allowed, purge it.
	if diffMinutes > lobbyMaxPersist {
		if ctx.Err() != nil {
			return
		}
		// Notify lobby users.
		msg := "Server:Lobby Closed"
		ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")
		log.Printf("[Job_Lobby_Stewardship]: Closing lobby %s", lobby.ID.String())
		if ctx.Err() != nil {
			return
		}
		// Delete the associated server, if it exists.
		if lobby.Lobby_Server != nil {
			if err := repository.DelServer(lobby.Lobby_Server.ID.String(), lobby.Lobby_Server); err != nil {
				log.Printf("[Job_Lobby_Stewardship]ERROR: deleting server for lobby %s: %v", lobby.ID.String(), err)
				return
			}
		}
		if ctx.Err() != nil {
			return
		}
		// Delete the lobby.
		if err := repository.DelLobby(lobby.ID.String(), lobby); err != nil {
			log.Printf("[Job_Lobby_Stewardship]ERROR: deleting lobby %s: %v", lobby.ID.String(), err)
		}
	}
}
