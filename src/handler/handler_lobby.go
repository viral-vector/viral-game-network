package handler

import (
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
)

// HandleRoot 
func Handle_AllLobby(c *fiber.Ctx) error {
	repository.AllLobby()
	
	return c.SendString("viral-game-network: " + cache.Get("viral-game-network"))
}

// Handle_GetLobby 
func Handle_GetLobby(c *fiber.Ctx) error {
	repository.GetLobby()
	
	return c.SendString("viral-game-network: " + cache.Get("viral-game-network"))
}

// Handle_SetLobby
func Handle_SetLobby(c *fiber.Ctx) error {
	repository.SetLobby()
	
	return c.SendString("viral-game-network: " + cache.Get("viral-game-network"))
}
