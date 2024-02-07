package handler

import (
	"time"
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
)

// Field names should start with an uppercase letter
type AuthRequestDTO struct {
    AppKey string `json:"Viral-Game-Network-AppKey" xml:"Viral-Game-Network-AppKey" form:"Viral-Game-Network-AppKey"`
	Username string `json:"username" xml:"username" form:"username"`
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

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Unauthorized: Bad Token @ " + err.Error(),
		})
	}

	record := dbtype.User{
		Name: user.Name,
	}

	user, err = repository.GetUser(record)

	c.Locals("user", user)

	return c.Next()
}

func Handle_AuthGuest(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "auth/guest")

	utmp := dbtype.User{
		Name:"guest_0",
		Guid:"guest_0",
	}
	user, err := repository.GetUser(utmp)
	if err != nil {
	user, err = repository.PutUser(utmp)
	}

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "User creation failed",
		})
	}

	access_token, err := auth.GenerateToken(user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status": "error",
			"message": "Token generation failed",
		})
	}

	// Update last login
	user.Date_LastLogin = time.Now().UTC().Format(time.RFC3339)

	// Update user
	user, err = repository.SetUser(user.ID, user)

	return c.JSON(fiber.Map{
		"status": "ok", 
		"access_token": access_token,
	})
}
