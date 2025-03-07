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
	lobby, err := check_read(c)

	if err != nil {
		return fmt.Errorf("Lobby Write Access Failed.")
	}

	user := c.Locals("user").(*dbtype.User)

	if lobby.Lobby_Host.ID != user.ID {
		return fmt.Errorf("Lobby Write Access Failed.")
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
	lobby_users := append(lobby.Lobby_Users, lobby.Lobby_Host)
	for _, lobby_user := range lobby_users {
		if user.ID == lobby_user.ID {
			return lobby, nil
		} 
	}

	return lobby, fmt.Errorf("Not authorized to access lobby.")
}