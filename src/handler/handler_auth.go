package handler

import (
	"fmt"
	"time"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"

	"github.com/gofiber/fiber/v2"
)

// Field names should start with an uppercase letter
type AuthRequestDTO struct {
	Name string `json:"name" xml:"name" form:"name"`
	Guid string `json:"guid" xml:"guid" form:"guid"`
}

func Handle_ValidateAppKey(c *fiber.Ctx) error {
	app_key := c.Get("Viral-Game-Network-AppKey")

	pass, err := auth.ValidateAppKey(app_key)

	if err != nil || pass == false {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Bad or missing App Key",
		})
	}
	return c.Next()
}

func Handle_ValidateToken(c *fiber.Ctx) error {
	token := c.Get("Viral-Game-Network-Token")

	user, err := auth.ValidateToken(token)

	if err != nil {
		fmt.Println(err, token)
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
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

func Handle_AuthLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "auth/lobby")

	dto := new(AuthRequestDTO)

	if err := c.BodyParser(dto); err != nil {
		return err
	}
	user := dbtype.User{
		Name: dto.Name,
		Guid: dto.Guid,
	}

	access_token, err := auth.GenerateToken(&user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Token generation failed",
		})
	}

	return c.JSON(fiber.Map{
		"status":       "ok",
		"access_token": access_token,
	})
}

func Handle_AuthGuest(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "auth/guest")

	dto := new(AuthRequestDTO)

	if err := c.BodyParser(dto); err != nil {
		return err
	}
	utmp := dbtype.User{
		Name: dto.Name,
		Guid: dto.Guid,
	}

	user, err := repository.GetUser(utmp)
	if err != nil {
		user, err = repository.PutUser(utmp)
	}

	if err != nil || user == nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "User creation failed",
		})
	}

	access_token, err := auth.GenerateToken(user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Token generation failed",
		})
	}

	// Update last login
	user.Date_LastLogin = time.Now().UTC().Format(time.RFC3339)

	// Update user
	user, err = repository.SetUser(user.ID, user)

	return c.JSON(fiber.Map{
		"status":       "ok",
		"access_token": access_token,
	})
}
