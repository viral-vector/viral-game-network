package handler_admin

import (
	"fmt"
	"math"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"
	"viral-game-network/src/utils/pagination"

	"github.com/gofiber/fiber/v2"
)

func Handle_Lobbies(c *fiber.Ctx) error {
	search := c.Query("search", "")
	perPage := 15
	curPage, err := pagination.Page(c.Query("page", "1"), perPage)
	if err != nil {
		return fiber.ErrBadRequest
	}
	lobbies, total, err := repository.AllLobby(perPage, curPage, search)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load lobbies")
	}

	return c.Render("admin/lobbies", fiber.Map{
		"lobbies": lobbies,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
		"search":  search,
	})
}

func Handle_Lobbies_Create_View(c *fiber.Ctx) error {
	apps, _, err := repository.AllApplication(-1, 1, "")
	if err != nil {
	}

	form, _ := form_builder.GenerateForm(
		"POST",
		"/admin/lobby",
		"create",
		dbtype.Lobby{},
		"Create Lobby",
		"Create Lobby",
	)
	form.Confirm = "Save Lobby Config?"

	// Lobby Application Options
	for field := range form.Fields {
		if form.Fields[field].Name == "lobby_application" {
			form.Fields[field].Options = []map[string]string{}
			for _, app := range apps {
				form.Fields[field].Options = append(form.Fields[field].Options, map[string]string{
					"label": app.Name,
					"value": app.ID.String(),
				})
			}
		}
	}

	return c.Render("admin/lobbies", fiber.Map{
		"form": form,
	})
}

func Handle_Lobbies_Create_Crud(c *fiber.Ctx) error {
	dto := new(dbtype.Lobby)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Create: %s", err),
		})
	}

	if dto.Lobby_Application == nil || dto.Lobby_Application.ID == nil {
		return fiber.ErrBadRequest
	}

	// Fetch Application
	app, err := repository.GetApplication(dto.Lobby_Application.ID.String())
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Create: Failed %s", err),
		})
	}

	// Clear & Put
	dto.Lobby_Host = nil
	dto.Lobby_Application = nil
	lobby, err := repository.PutLobby(dto, app, nil)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Create: Failed %s", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Lobby Create: Success %s", lobby.Guid),
	})
}

func Handle_Lobbies_Update_View(c *fiber.Ctx) error {
	id := c.Params("id")

	lobby, err := repository.GetLobby(id)
	if err != nil || lobby == nil {
		return c.Redirect("/admin/lobbies")
	}

	apps, _, err := repository.AllApplication(-1, 1, "")
	if err != nil {
	}

	form, _ := form_builder.GenerateForm(
		"POST",
		"/admin/lobby/"+lobby.ID.String(),
		"update",
		lobby,
		"Update Lobby",
		"Update Lobby",
	)
	form.Confirm = "Save Lobby Config?"

	// Lobby Application Options
	for field := range form.Fields {
		if form.Fields[field].Name == "lobby_application" {
			form.Fields[field].Options = []map[string]string{}
			for _, app := range apps {
				form.Fields[field].Options = append(form.Fields[field].Options, map[string]string{
					"label": app.Name,
					"value": app.ID.String(),
				})
			}
		}
	}

	return c.Render("admin/lobbies", fiber.Map{
		"form": form,
	})
}

func Handle_Lobbies_Update_Crud(c *fiber.Ctx) error {
	id := c.Params("id")

	dto := new(repository.LobbyPatch)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Update: Failed %s", err),
		})
	}

	// Fetch lobby
	lobby, err := repository.GetLobby(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Update: Failed %s", err),
		})
	}

	if lobby == nil {
		return fiber.ErrNotFound
	}

	lobby, err = repository.PatchLobby(lobby.ModelID(), dto)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Update: Failed %s", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Lobby Update: Success %s", lobby.Name),
	})
}

func Handle_Lobbies_Delete_Crud(c *fiber.Ctx) error {
	id := c.Params("id")

	// Fetch Lobby
	lobby, err := repository.GetLobby(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Delete: %s", err),
		})
	}

	if lobby == nil {
		return fiber.ErrNotFound
	}

	err = repository.DelLobby(lobby.ID.String(), lobby)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Lobby Delete: %s", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":   "success",
		"redirect": "/admin/lobbies",
		"message":  fmt.Sprintf("Lobby Delete: Success %s", lobby.Guid),
	})
}
