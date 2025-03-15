package job

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/ministration"
)

/**
 * Manages Lobby Lifecycle
 */
func Job_Lobby_Stewardship() {
	lobbies, _, err := repository.AllLobby(-1, 1)
	if err != nil {
		fmt.Errorf("Job_Lobby_Stewardship Error:  %s", err)
		return
	}

	// Parse the string into a float64
	lobby_max_persist, err := strconv.ParseFloat(os.Getenv("LOBBY_MAX_PERSIST"), 64)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, lobby := range lobbies {
		// Get the label for the pod
		label := strings.Split(lobby.ID.String(), ":")[1]
		
		// Check if we have a lock on the pod
		if cache_lock, _ := cache.Get[string]("lobby-stewardship-lock-" + label); cache_lock != "" {
			continue
		}
		go func(lobby *dbtype.Lobby, label string, lobby_max_persist float64) {
			// Lock the Lobby
			cache.Set[string]("lobby-stewardship-lock-"+label, "true", time.Second*30)

			fmt.Println("RUN_Lobby_Stewardship: @ Looking", label)
			// RUN
			RUN_Lobby_Stewardship(lobby, lobby_max_persist)
			// Unlock the Lobby
			cache.Del("lobby-stewardship-lock-" + label)
		}(&lobby, label, lobby_max_persist)
	}
}

func RUN_Lobby_Stewardship(lobby *dbtype.Lobby, lobby_max_persist float64) {
	fmt.Println("Job_Lobby_Stewardship: @ Looking", lobby.ID.String())

	date_now := time.Now().UTC()

	// Use Date_Updated if available; otherwise, fall back to Date_Created.
	dateStr := lobby.Date_Updated
	if dateStr == "" {
		dateStr = lobby.Date_Created
	}

	date_lobby, _ := time.Parse(time.RFC3339, dateStr)
	date_diffr := date_now.Sub(date_lobby)

	if date_diffr.Minutes() > lobby_max_persist {
		// Notify Lobby Users
		msg := "Server:Lobby Closed"
		ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")

		fmt.Println("RUN_Lobby_Stewardship: @ CLosing", lobby.ID.String(), date_diffr.Minutes())
		// Delete Server
		if lobby.Lobby_Server != nil {
			repository.DelServer(lobby.Lobby_Server.ID.String())
		}
		// Delete Lobby
		repository.DelLobby(lobby.ID.String())
	}
}