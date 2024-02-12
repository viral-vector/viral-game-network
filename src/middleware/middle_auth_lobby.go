package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
)

func Lobby_Read(c *fiber.Ctx) error {
	_, err := check_read(c)

	if err != nil {
		return fmt.Errorf("Lobby Read Access Failed.")
	}
	return c.Next()
}

func Lobby_Write(c *fiber.Ctx) error {
	_, err := check_read(c)

	if err != nil {
		return fmt.Errorf("Lobby Read Access Failed.")
	}

	return c.Next()
}


func check_read(c *fiber.Ctx) (*dbtype.Lobby, error) {
	id := c.Params("id")
	user := c.Locals("user").(*dbtype.User)

	// Check Lobby 
	lobby, err := repository.GetLobby(id)
	if err != nil {
		return nil, err
	}
	if lobby == nil {
		return nil, fmt.Errorf("Lobby nnot found.")
	}

	// Check host/user
	host := new(dbtype.LobbyUser)
	host.User = *lobby.Lobby_Host
	lobby_users := append(lobby.Lobby_Users, host)
	for _, lobby_user := range lobby_users {
		if user.ID == lobby_user.User.ID {
			return lobby, nil
		} 
	}

	return nil, fmt.Errorf("Not authorized to access lobby.")
}