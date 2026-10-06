package handler_admin

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math"
	"viral-game-network/src/database/repository"
	form_builder "viral-game-network/src/utils/form_builder"
	"viral-game-network/src/utils/pagination"
)

func Handle_Servers(c *fiber.Ctx) error {
	search := c.Query("search", "")
	perPage := 15
	curPage, err := pagination.Page(c.Query("page", "1"), perPage)
	if err != nil {
		return fiber.ErrBadRequest
	}
	servers, total, err := repository.AllServer(perPage, curPage, search)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load servers")
	}

	return c.Render("admin/servers", fiber.Map{
		"servers": servers,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
		"search":  search,
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
		"form": form,
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

	if server == nil {
		return fiber.ErrNotFound
	}

	err = repository.DelServer(server.ID.String(), server)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Server Delete: %s", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":   "success",
		"redirect": "/admin/servers",
		"message":  fmt.Sprintf("Server Delete: Success %s", server.Guid),
	})
}
