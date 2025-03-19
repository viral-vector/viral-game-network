package handler

import (
	"viral-game-network/src/database/repository"
	"viral-game-network/src/ministration"

	"github.com/gofiber/fiber/v2"
)

// Handle_HostTick
func Handle_TickHost(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "host/tick")

	lobby, err := repository.GetLobby(c.Params("id"))

	if err != nil || lobby == nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching lobby",
		})
	}

	server, err := repository.TickServer(lobby.Lobby_Server.ID.String())

	if err != nil || server == nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching server " + lobby.Lobby_Server.ID.String(),
		})
	}

	// Notify Lobby Users
	if lobby.Lobby_Server.Status != "Online" {
		msg := "Server:Online"
		ministration.Service_Lobby_Notify(lobby.ID.String(), msg, "")
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   server,
	})
}
