package handler_admin

import (
	"fmt"
	"math"
	"strconv"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"

	"github.com/gofiber/fiber/v2"
)

// ##> Admin
func Handle_Admins(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	admins, total, _ := repository.AllAdmin(perPage, curPage)

	return c.Render("admin/admins", fiber.Map{
		"admins" : admins,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

func Handle_Admins_Create_View(c *fiber.Ctx) error {
	form, _ := form_builder.GenerateForm(
		"POST", 
		"/admin/admin",  
		dbtype.Admin{},
		"Create Admin",
	)
	form.Confirm = "Save Admin Config?"

	return c.Render("admin/admins", fiber.Map{
		"form" : form,
	})
}

func Handle_Admins_Create_Crud(c *fiber.Ctx) error {
	dto := new(dbtype.Admin)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Create: Failed %s", err),
		})
	}

	// hash Password
	password, err := auth.HashGenerate(dto.Password)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Create: Failed %s", err),
		})
	}
	dto.Password = password
	
	admin, err := repository.PutAdmin(dto)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Create: Failed %s", err),
		})
	}

	repository.PutSystemEvent(&dbtype.SystemEvent{
		Severity: "info",
		Message:  fmt.Sprintf("Admin Create: Success %s", admin.Name) ,
		Ref_Source: "system",
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Admin Create: Success %s", admin.Name),
	})
}

func Handle_Admins_Update_View(c *fiber.Ctx) error {
	return nil
}

func Handle_Admins_Update_Crud(c *fiber.Ctx) error {
	return nil
}

func Handle_Admins_Delete_Crud(c *fiber.Ctx) error {
	return nil
}