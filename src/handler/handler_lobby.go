package handler

import (
	"os"
	"strconv"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
)

// Handle_AllLobby 
func Handle_AllLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/all")

	data, err := repository.AllLobby()

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Error fetching lobbies",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": data,
	})	
}

// Handle_SetLobby
func Handle_SetLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/set")

	record := new(dbtype.Lobby)
	err := json.Unmarshal(c.Body(), record)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Bad set lobby request",
		})
	}

	lobby, err := repository.SetLobby(c.Params("id"), record)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Error setting lobby",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": lobby,
	})
}

// Handle_GetLobby 
func Handle_GetLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/get")

	lobby, err := repository.GetLobby(c.Params("id"))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Error fetching lobby",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": lobby,
	})
}

// Handle_JoinLobby 
func Handle_JoinLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/join")

	lobby, err := repository.GetLobby(c.Params("id"))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Error fetching lobby",
		})
	}

	max_players, err := strconv.ParseInt(os.Getenv("LOBBY_MAX_PLAYERS"), 0, 0)
	if err!= nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Server error fetching lobby max players", 
		})
	}

	if len(lobby.Lobby_Users) >= int(max_players) {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Lobby is full",
		})
	}

	if lobby.Private == true && lobby.Code != c.Params("code") {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Invalid lobby code",
		})
	}

	req_user := c.Locals("user").(*dbtype.User)

	for _, user := range lobby.Lobby_Users {
		if user.User.ID == req_user.ID {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"status": "error",
				"message": "Already joined lobby",
			})
		}
	}

	// Link User to lobby
	err = repository.LinkLobbyUser(lobby.ID, req_user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Error joining lobby: " + err.Error(),
		})
	}

	lobby, err = repository.GetLobby(lobby.ID)

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": lobby,
	})
}

// Handle_HostLobby 
func Handle_HostLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/host")

	record := new(dbtype.Lobby)
	if err := c.BodyParser(record); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Bad lobby host request",
		})
	}

	user := c.Locals("user").(*dbtype.User)

	lobby, err := repository.PutLobby(record, user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Error hosting lobby: " + err.Error() ,
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": lobby,
	})
}