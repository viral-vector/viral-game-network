package handler_admin

import (
	"fmt"
	"math"
	"strconv"
	"viral-game-network/src/database/repository"
	form_builder "viral-game-network/src/utils/form_builder"
	"github.com/gofiber/fiber/v2"
)

func Handle_Servers(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	servers, total, _ := repository.AllServer(perPage, curPage)

	return c.Render("admin/servers", fiber.Map{
		"servers": servers,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
	})
}

func Handle_Servers_Update_View(c *fiber.Ctx) error {
	id := c.Params("id")

	server, err := repository.GetServer(id)
	if err != nil || server == nil {
		return c.Redirect("/admin/servers")
	}

	form, _ := form_builder.GenerateForm(
		"", 
		"",
		"update", 
		server,
		"View Server",
		"",
	)
	form.DisableSubmit = true

	return c.Render("admin/servers", fiber.Map{
		"form" : form,
	})
}

func Handle_Servers_Delete_Crud(c *fiber.Ctx) error {
	id := c.Params("id")

	// Fetch Server
	server, err := repository.GetServer(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Server Delete: %s", err),
		})
	}

	repository.DelServer(server.ID.String(), server)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Server Delete: %s", err),
		})
	}
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"redirect": "/admin/servers",
		"message": fmt.Sprintf("Server Delete: Success %s", server.Guid),
	})
}