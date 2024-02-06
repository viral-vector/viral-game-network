package handler

import (
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/type"
)

// Field names should start with an uppercase letter
type AuthRequestDTO struct {
    AppKey string `json:"Viral-Game-Network-AppKey" xml:"Viral-Game-Network-AppKey" form:"Viral-Game-Network-AppKey"`
}

func Handle_ValidateAppKey(c *fiber.Ctx) error {
	dto := new(AuthRequestDTO)

	if err := c.BodyParser(dto); err != nil {
		return err
	}

	pass, err := auth.ValidateAppKey(dto.AppKey)

	if err != nil || pass == false {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Unauthorized: Bad or missing App Key",
		})
	}
	return c.Next()
}

func Handle_ValidateToken(c *fiber.Ctx) error {
	token := c.Get("Viral-Game-Network-Token")

	user, err := auth.ValidateToken(token)

	c.Locals("user", user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Unauthorized: Bad or missin User Token",
		})
	}
	return c.Next()
}

func Handle_AuthGuest(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "auth/guest")

	user := new(dbtype.User)
	user.Name = "guest_0"

	access_token, err := auth.GenerateToken(user)

	if err!= nil {
		panic(err)
	}

	return c.JSON(fiber.Map{
		"status": "ok", 
		"access_token": access_token,
	})
}
