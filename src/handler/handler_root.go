package handler

import (
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/cache"
)

func HandleRoot(c *fiber.Ctx) error {
	return c.SendString("viral-game-network: " + cache.Get("viral-game-network"))
}
