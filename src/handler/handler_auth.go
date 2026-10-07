package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"strings"
	"time"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"
)

type AuthUserRequestDTO struct {
	Name string `json:"name" xml:"name" form:"name"`
}

type AuthAdminRequestDTO struct {
	Username string `json:"username,omitempty" form:"username,label:Username,type:string"`
	Password string `json:"password,omitempty" form:"password,label:Password,type:password"`
}

var NVET_COOKIE_KEY string = "VNET_SESSION"

func Handle_ValidateAppKey(c fiber.Ctx) error {
	app_key := c.Get("Viral-Game-Network-AppKey")

	pass, err := auth.ValidateAppKey(app_key)

	if err != nil || !pass {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Unauthorized %s", err),
		})
	}
	return c.Next()
}

func Handle_ValidateTokenAdmin(c fiber.Ctx) error {
	token := c.Cookies(NVET_COOKIE_KEY)

	claims, err := auth.ValidateTokenFor(token, auth.AdminAudience)

	if err != nil {
		c.Status(fiber.StatusForbidden)
		return c.Redirect().Status(fiber.StatusFound).Route("auth/admin")
	}

	admin, err := repository.GetAdmin(claims.Subject)
	if err != nil || admin == nil {
		return fiber.ErrForbidden
	}
	c.Locals("admin", admin)
	c.Locals("session_expires", claims.ExpiresAt.Time)
	c.Locals("session_expires_unix", claims.ExpiresAt.Unix())

	return c.Next()
}

func Handle_RedirectAdmin(c fiber.Ctx) error {
	token := c.Cookies(NVET_COOKIE_KEY)

	claims, err := auth.ValidateTokenFor(token, auth.AdminAudience)

	if err == nil {
		admin, lookupErr := repository.GetAdmin(claims.Subject)
		if lookupErr != nil || admin == nil {
			return c.Next()
		}
		c.Status(fiber.StatusBadRequest)
		return c.Redirect().Status(fiber.StatusFound).Route("admin")
	}

	return c.Next()
}

func Handle_ValidateTokenUsers(c fiber.Ctx) error {
	token := c.Get("Viral-Game-Network-Token")

	claims, err := auth.ValidateTokenFor(token, auth.UserAudience)

	if err != nil {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Bad Token @ " + err.Error(),
		})
	}

	user, err := repository.GetUser(claims.Subject)
	if err != nil || user == nil {
		return fiber.ErrForbidden
	}
	c.Locals("user", user)
	c.Locals("session_expires", claims.ExpiresAt.Time)
	c.Locals("session_expires_unix", claims.ExpiresAt.Unix())

	return c.Next()
}

func Handle_AuthLobby(c fiber.Ctx) error {
	dto := new(AuthUserRequestDTO)

	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	user, ok := c.Locals("user").(*dbtype.User)
	if !ok || user == nil {
		return fiber.ErrForbidden
	}
	access_token, err := auth.GenerateToken(user.Name, user.ModelID(), auth.UserAudience)

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

func Handle_AuthGuest(c fiber.Ctx) error {
	dto := new(AuthUserRequestDTO)

	if err := c.Bind().Body(dto); err != nil {
		return err
	}
	name := strings.TrimSpace(dto.Name)
	if name == "" || len(name) > 128 {
		return fiber.ErrBadRequest
	}
	// A display name is not a credential. Each guest login creates a new identity.
	user, err := repository.PutUser(&dbtype.User{Name: name})

	if err != nil || user == nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "User creation failed",
		})
	}

	access_token, err := auth.GenerateToken(user.Name, user.ModelID(), auth.UserAudience)

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
	if _, err := repository.SetUser(user.ID.String(), user); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"access_token": access_token,
	})
}

func Handle_AuthAdmin(c fiber.Ctx) error {
	post := strings.ToLower(c.Method()) == "post"

	// Get - show form
	if post == false {
		form, _ := form_builder.GenerateForm(
			"POST",
			"/auth/admin",
			"",
			AuthAdminRequestDTO{},
			"",
			"Login",
		)
		return c.Render("admin/login", fiber.Map{
			"form": form,
		})
	}

	// Post - Gen token
	dto := new(AuthAdminRequestDTO)
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	// Fetch admin
	admin, err := repository.SelAdmin(&dbtype.Admin{
		ID:    nil,
		Name:  dto.Username,
		Email: dto.Username,
		Phone: dto.Username,
	})
	if err != nil || admin == nil {
		return fiber.ErrForbidden
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
	access_token, err := auth.GenerateToken(admin.Name, admin.ModelID(), auth.AdminAudience)
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
	if _, err := repository.SetAdmin(admin.ID.String(), admin); err != nil {
		return fiber.ErrInternalServerError
	}

	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = NVET_COOKIE_KEY
	cookie.Value = access_token
	claims, err := auth.ValidateTokenFor(access_token, auth.AdminAudience)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	cookie.Expires = claims.ExpiresAt.Time
	cookie.Path = "/"
	cookie.Secure = c.Secure()
	cookie.SameSite = "Strict"
	cookie.HTTPOnly = true

	// Set cookie
	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"message":    "Welcome back " + admin.Name,
		"redirect":   "/admin",
		"status":     "success",
		"expires_at": claims.ExpiresAt.Unix(),
	})
}

func Handle_AuthAdminRefresh(c fiber.Ctx) error {
	admin, ok := c.Locals("admin").(*dbtype.Admin)

	if !ok || admin == nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Admin not found",
		})
	}

	// Gen Token
	access_token, err := auth.GenerateToken(admin.Name, admin.ModelID(), auth.AdminAudience)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Token generation failed: " + err.Error(),
		})
	}

	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = NVET_COOKIE_KEY
	cookie.Value = access_token
	claims, err := auth.ValidateTokenFor(access_token, auth.AdminAudience)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	cookie.Expires = claims.ExpiresAt.Time
	cookie.Path = "/"
	cookie.Secure = c.Secure()
	cookie.SameSite = "Strict"
	cookie.HTTPOnly = true

	// Set cookie
	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"message":    "Authorized " + admin.Name,
		"status":     "success",
		"expires_at": claims.ExpiresAt.Unix(),
	})

}

func Handle_AuthAdminLogout(c fiber.Ctx) error {
	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = NVET_COOKIE_KEY
	cookie.Value = ""
	cookie.Expires = time.Unix(1, 0)
	cookie.MaxAge = -1
	cookie.Path = "/"
	cookie.Secure = c.Secure()
	cookie.SameSite = "Strict"
	cookie.HTTPOnly = true

	// Set cookie
	c.Cookie(cookie)

	return c.Redirect().Status(fiber.StatusFound).Route("admin")
}
