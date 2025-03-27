package handler

import (
	"github.com/gofiber/fiber/v2"
)

func Handle_Root(c *fiber.Ctx) error {
	return c.SendString("viral-game-network")
}

func Handle_Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "success",
	})
}
