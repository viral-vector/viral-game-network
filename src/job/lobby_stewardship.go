package job

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"viral-game-network/src/database/repository"
	"viral-game-network/src/ministration"
)

/**
 * Manages Lobby Lifecycle
 */
func Job_Lobby_Stewardship() {
	// fmt.Println("Job_Lobby_Stewardship")

	lobbies, _, err := repository.AllLobby(-1, 1)
	if err != nil {
		fmt.Errorf("Job_Lobby_Stewardship:  %s", err)
		return
	}

	date_now := time.Now().UTC()

	// Parse the string into a float64
	lobby_max_persist, err := strconv.ParseFloat(os.Getenv("LOBBY_MAX_PERSIST"), 64)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, lobby := range lobbies {
		date_lobby, _ := time.Parse(time.RFC3339, lobby.Date_Updated)
		date_diffr := date_now.Sub(date_lobby)

		if date_diffr.Minutes() > lobby_max_persist {
			// Notify Lobby Users
			msg := "Server:Lobby Closed"
			ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")

			// Delete Server
			if lobby.Lobby_Server != nil {
				repository.DelServer(lobby.Lobby_Server.ID.String())
			}

			// Delete Lobby
			repository.DelLobby(lobby.ID.String())
		}
		fmt.Println("Job_Lobby_Stewardship: @", lobby.ID.String())
	}
}
