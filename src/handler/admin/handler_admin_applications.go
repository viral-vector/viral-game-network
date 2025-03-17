package handler_admin

import (
	"fmt"
	"math"
	"strconv"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"

	"github.com/gofiber/fiber/v2"
)

// ##> Application
func Handle_Applications(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	apps, total, _ := repository.AllApplication(perPage, curPage)

	return c.Render("admin/applications", fiber.Map{
		"apps" : apps,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

func Handle_Applications_Create_View(c *fiber.Ctx) error {
	form, _ := form_builder.GenerateForm(
		"POST", 
		"/admin/application",  
		dbtype.Application{},
		"Create Application",
	)

	return c.Render("admin/applications", fiber.Map{
		"form" : form,
	})
}

func Handle_Applications_Create_Crud(c *fiber.Ctx) error {
	dto := new(dbtype.Application)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Application Create: Failed %s", err),
		})
	}

	fmt.Println(dto)

	app, err := repository.PutApplication(dto)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Application Create: Failed %s", err),
		})
	}

	repository.PutSystemEvent(&dbtype.SystemEvent{
		Severity: "info",
		Message:  fmt.Sprintf("Application Create: Success %s", app.Guid) ,
		Ref_Source: "system",
	})

	return c.JSON(fiber.Map{
		"status":  "error",
		"message": fmt.Sprintf("Application Create: Success %s", app.Guid),
	})
}

func Handle_Applications_Update_View(c *fiber.Ctx) error {
	return nil
}

func Handle_Applications_Update_Crud(c *fiber.Ctx) error {
	return nil
}

func Handle_Applications_Delete_Crud(c *fiber.Ctx) error {
	return nil
}