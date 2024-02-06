package handler

import (
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/cache"
)

func Handle_Root(c *fiber.Ctx) error {
	return c.SendString("viral-game-network: " + cache.Get("viral-game-network"))
}

func Handle_Health(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "root/health")

	return c.JSON(fiber.Map{
		"status": "ok",
		"viral-game-network": cache.Get("viral-game-network"),
	})
}
