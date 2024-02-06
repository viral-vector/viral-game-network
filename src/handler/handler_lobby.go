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

	data := repository.AllLobby()

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
		panic(err)
	}
	var lobby *dbtype.Lobby = repository.SetLobby(c.Params("id"), record)

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": lobby,
	})
}

// Handle_GetLobby 
func Handle_GetLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/get")

	var lobby *dbtype.Lobby = repository.GetLobby(c.Params("id"))

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": lobby,
	})
}

// Handle_JoinLobby 
func Handle_JoinLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/join")

	var lobby *dbtype.Lobby = repository.GetLobby(c.Params("id"))

	max_players, err := strconv.ParseInt(os.Getenv("LOBBY_MAX_PLAYERS"), 0, 0)
	if err!= nil {
		panic(err)
	}
	if len(lobby.Users) >= int(max_players) {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "lobby is full",
		})
	}

	if lobby.Private == true && lobby.Code != c.Params("code") {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "invalid lobby code",
		})
	}

	req_user := dbtype.User{
		Name: "viral-vector-" + c.Get("Viral-Game-Network-Entity"),
	}

	for _, user := range lobby.Users {
		if user.Name == req_user.Name {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"status": "error",
				"message": "already joined",
			})
		}
	}

	lobby.Users = append(lobby.Users, req_user)

	return c.JSON(fiber.Map{
		"status": "ok",
		"data": repository.SetLobby(c.Params("id"), lobby),
	})
}
