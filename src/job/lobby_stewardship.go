package job

import (
	"os"
	"fmt"
	"time"
	"strconv"
	// "viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/service"
)

func Job_Lobby_Stewardship() {
	fmt.Println("Job_Lobby_Stewardship")

	lobbies, err := repository.AllLobby()
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
			service.Service_Lobby_Notify(lobby.ID, msg, "")

			// Delete Server
			repository.DelServer(lobby.Lobby_Server.ID)

			// Delete Lobby
			repository.DelLobby(lobby.ID)
		}

		// Do we need a new server?
		fmt.Println("Job_Lobby_Stewardship: @",  lobby.Name, lobby.ID, date_diffr.Minutes())
	}
}