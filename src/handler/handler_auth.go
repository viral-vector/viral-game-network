package handler

import (
	"time"
	"strings"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"
	"github.com/gofiber/fiber/v2"
)

type AuthUserRequestDTO struct {
	Name string `json:"name" xml:"name" form:"name"`
}

type AuthAdminRequestDTO struct {
	Username string `json:"username,omitempty" form:"username,label:Username,type:string"`
	Password string `json:"password,omitempty" form:"password,label:Password,type:password"`
}

var NVET_COOKIE_KEY string = "VNET_SESSION"

func Handle_ValidateAppKey(c *fiber.Ctx) error {
	app_key := c.Get("Viral-Game-Network-AppKey")

	pass, err := auth.ValidateAppKey(app_key)

	if err != nil || !pass {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Bad or missing App Key",
		})
	}
	return c.Next()
}

func Handle_ValidateTokenAdmin(c *fiber.Ctx) error {
	token := c.Cookies(NVET_COOKIE_KEY)

	claims, err := auth.ValidateToken(token)

	if err != nil {
		c.Status(fiber.StatusForbidden)
		return c.RedirectToRoute("auth/admin", fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Bad Token @ " + err.Error(),
		})
	}

	record := dbtype.Admin{
		ID: nil,
		Name: claims.Username,
		Email: claims.Username,
		Phone: claims.Username,
	}

	admin, err := repository.SelAdmin(&record)
	if err == nil {
		c.Locals("admin", admin)
	}

	return c.Next()
}

func Handle_RedirectAdmin(c *fiber.Ctx) error {
	token := c.Cookies(NVET_COOKIE_KEY)

	_, err := auth.ValidateToken(token)

	if err == nil {
		c.Status(fiber.StatusBadRequest)
		return c.RedirectToRoute("admin", fiber.Map{
			"status":  "success",
			"message": "Authorized",
		})
	}
	
	return c.Next()
}

func Handle_ValidateTokenUsers(c *fiber.Ctx) error {
	token := c.Get("Viral-Game-Network-Token")

	claims, err := auth.ValidateToken(token)

	if err != nil {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Bad Token @ " + err.Error(),
		})
	}

	record := dbtype.User{
		Guid: claims.Username,
		Name: claims.Username,
	}

	user, err := repository.SelUser(&record)
	if err == nil {
		c.Locals("user", user)
	}

	return c.Next()
}

func Handle_AuthLobby(c *fiber.Ctx) error {
	dto := new(AuthUserRequestDTO)

	if err := c.BodyParser(dto); err != nil {
		return err
	}

	// TODO: Get User from repo first  
	user := dbtype.User{
		Guid: dto.Name,
		Name: dto.Name,
	}

	access_token, err := auth.GenerateToken(user.Name)

	if err != nil {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Token generation failed: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"access_token": access_token,
	})
}

func Handle_AuthGuest(c *fiber.Ctx) error {
	dto := new(AuthUserRequestDTO)

	if err := c.BodyParser(dto); err != nil {
		return err
	}
	utmp := dbtype.User{
		Guid: dto.Name,
		Name: dto.Name,
	}

	user, err := repository.SelUser(&utmp)
	if err != nil {
		user, err = repository.PutUser(&utmp)
	}

	if err != nil || user == nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "User creation failed",
		})
	}

	access_token, err := auth.GenerateToken(user.Name)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Token generation failed: " + err.Error(),
		})
	}

	// Update last login
	user.Date_LastLogin = time.Now().UTC().Format(time.RFC3339)

	// Update user
	repository.SetUser(user.ID.String(), user)

	return c.JSON(fiber.Map{
		"status":       "success",
		"access_token": access_token,
	})
}

func Handle_AuthAdmin(c *fiber.Ctx) error {
	post := strings.ToLower(c.Method()) == "post"

	// Get - show form
	if post == false {
		form, _ := form_builder.GenerateForm(
			"POST", 
			"/auth/admin",
			"",
			AuthAdminRequestDTO{},
			"",
		)
		return c.Render("admin/login", fiber.Map{
			"form" : form,
		})
	}
	
	// Post - Gen token
	dto := new(AuthAdminRequestDTO)
	if err := c.BodyParser(dto); err != nil {
		return err
	}

	// Fetch admin
	admin, err := repository.SelAdmin(&dbtype.Admin {
		ID: nil,
		Name: dto.Username,
		Email: dto.Username,
		Phone: dto.Username,
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Admin fetch failed: " + err.Error(),
		})
	}

	// Validate Password 
	valid, err := auth.HashValidate(dto.Password, admin.Password)
	if valid == false {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Password check failed: " + err.Error(),
		})
	}

	// Gen Token
	access_token, err := auth.GenerateToken(admin.Name)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Token generation failed: " + err.Error(),
		})
	}

	// Update last login
	admin.Date_LastLogin = time.Now().UTC().Format(time.RFC3339)
	// Update admin
	repository.SetAdmin(admin.ID.String(), admin)

	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = NVET_COOKIE_KEY
	cookie.Value = access_token
	cookie.Expires = time.Now().Add(24 * time.Hour)

	// Set cookie
	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"message" 		: "Welcome back " + admin.Name,
		"redirect"		: "/admin",
		"status"		: "success",
		"access_token"	: access_token,
	})
}

func Handle_LogoutAdmin(c *fiber.Ctx) error {
	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = NVET_COOKIE_KEY
	cookie.Value = ""
	cookie.Expires = time.Now().Add(0 * time.Second)

	// Set cookie
	c.Cookie(cookie)

	return c.RedirectToRoute("admin", fiber.Map{
		"status":  "success",
		"message": "Authorized",
	})
}
