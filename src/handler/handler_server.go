package handler

import (
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/ministration"

	"github.com/gofiber/fiber/v2"
)

// Heartbeats accept either the owning player or a server session for this lobby.
// Server sessions have no access to the general player or admin APIs.
func Handle_ValidateHeartbeatToken(c *fiber.Ctx) error {
	claims, err := auth.ValidateToken(c.Get("Viral-Game-Network-Token"))
	if err != nil {
		return fiber.ErrForbidden
	}
	serverSession := claims.Audience[0] == string(auth.ServerAudience)
	if !serverSession && claims.Audience[0] != string(auth.UserAudience) {
		return fiber.ErrForbidden
	}
	if serverSession && claims.Subject != c.Params("id") {
		return fiber.ErrForbidden
	}
	lobby, err := repository.GetLobby(c.Params("id"))
	if err != nil {
		return fiber.ErrInternalServerError
	}
	if lobby == nil {
		return fiber.ErrBadRequest
	}
	if !serverSession {
		user, err := repository.GetUser(claims.Subject)
		if err != nil || user == nil || lobby.Lobby_Host == nil || lobby.Lobby_Host.ModelID() != user.ModelID() {
			return fiber.ErrForbidden
		}
	}
	c.Locals("server_session", serverSession)
	return c.Next()
}

func Handle_RefreshServerToken(c *fiber.Ctx) error {
	if server, _ := c.Locals("server_session").(bool); !server {
		return fiber.ErrForbidden
	}
	lobby, err := repository.GetLobby(c.Params("id"))
	if err != nil || lobby == nil || lobby.Lobby_Server == nil {
		return fiber.ErrBadRequest
	}
	token, err := auth.GenerateToken("server:"+lobby.Guid, lobby.ModelID(), auth.ServerAudience)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"status": "success", "access_token": token})
}

// Handle_HostTick
func Handle_TickHost(c *fiber.Ctx) error {
	lobby, err := repository.GetLobby(c.Params("id"))

	if err != nil || lobby == nil || lobby.Lobby_Server == nil {
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
