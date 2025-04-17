package handler_admin

import (
	"fmt"
	"math"
	"strconv"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"
	struct_merge "viral-game-network/src/utils/struct_merge"

	"github.com/gofiber/fiber/v2"
)

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
		"create",
		dbtype.Application{},
		"Create Application",
		"Create Application",
	)
	form.Confirm = "Save Application Config?"

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
		"status":  "success",
		"message": fmt.Sprintf("Application Create: Success %s", app.Guid),
	})
}

func Handle_Applications_Update_View(c *fiber.Ctx) error {
	id := c.Params("id")

	application, err := repository.GetApplication(id)
	if err != nil || application == nil {
		return c.Redirect("/admin/applications")
	}

	form, _ := form_builder.GenerateForm(
		"POST", 
		"/admin/application/" + application.ID.String(),
		"update", 
		application,
		"Update Application",
		"Update Application",
	)
	form.Confirm = "Save Application Config?"

	return c.Render("admin/applications", fiber.Map{
		"form" : form,
	})
}

func Handle_Applications_Update_Crud(c *fiber.Ctx) error {
	id := c.Params("id")

	dto := new(dbtype.Application)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Application Update: Failed %s", err),
		})
	}

	// Fetch application
	application, err := repository.GetApplication(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Application Update: Failed %s", err),
		})
	}

	// Prop Merge
	dto.ID = application.ID
	dto.Date_Created = application.Date_Created
 
	if err := struct_merge.Merge[dbtype.Application](application, dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Application Update: Failed %s", err),
		})
	}
	// Set
	_, err = repository.SetApplication(application.ID.String(), dto)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Application Update: Failed %s", err),
		})
	}

	repository.PutSystemEvent(&dbtype.SystemEvent{
		Severity: "info",
		Message:  fmt.Sprintf("Application Update: Success %s", application.Name) ,
		Ref_Source: "system",
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Application Update: Success %s", application.Name),
	})
}

func Handle_Applications_Delete_Crud(c *fiber.Ctx) error {
	return nil
}