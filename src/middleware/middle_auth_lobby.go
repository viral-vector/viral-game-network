package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/database/type"
)

func Lobby_Read(c fiber.Ctx) error {
	_, err := check_read(c)

	if err != nil {
		return fiber.ErrForbidden
	}
	return c.Next()
}

func Lobby_Write(c fiber.Ctx) error {
	lobby, err := check_read(c)

	if err != nil {
		return fiber.ErrForbidden
	}

	user := c.Locals("user").(*dbtype.User)

	if lobby.Lobby_Host == nil || lobby.Lobby_Host.ModelID() != user.ModelID() {
		return fiber.ErrForbidden
	}
	return c.Next()
}

func check_read(c fiber.Ctx) (*dbtype.Lobby, error) {
	id := c.Params("id")
	user, ok := c.Locals("user").(*dbtype.User)
	if !ok || user == nil {
		return nil, fiber.ErrForbidden
	}

	// Check Lobby
	lobby, err := repository.GetLobby(id)
	if err != nil {
		return nil, err
	}
	if lobby == nil {
		return nil, fmt.Errorf("Lobby nnot found.")
	}

	if hasLobbyAccess(lobby, user) {
		return lobby, nil
	}
	return lobby, fiber.ErrForbidden
}

func hasLobbyAccess(lobby *dbtype.Lobby, user *dbtype.User) bool {
	if lobby == nil || user == nil || user.ID == nil {
		return false
	}
	if lobby.Lobby_Host != nil && lobby.Lobby_Host.ModelID() == user.ModelID() {
		return true
	}
	for _, member := range lobby.Lobby_Users {
		if member != nil && member.ID != nil && member.ModelID() == user.ModelID() {
			return true
		}
	}
	return false
}
