package handler_admin

import (
	"fmt"
	"math"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"
	"viral-game-network/src/utils/pagination"
	struct_merge "viral-game-network/src/utils/struct_merge"

	"github.com/gofiber/fiber/v3"
)

func Handle_Users(c fiber.Ctx) error {
	search := c.Query("search", "")
	perPage := 15
	curPage, err := pagination.Page(c.Query("page", "1"), perPage)
	if err != nil {
		return fiber.ErrBadRequest
	}
	users, total, err := repository.AllUser(perPage, curPage, search)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load users")
	}

	return c.Render("admin/users", fiber.Map{
		"users":  users,
		"total":  total,
		"pages":  int(math.Ceil(float64(total) / float64(perPage))),
		"paged":  curPage,
		"search": search,
	})
}

func Handle_Users_Create_View(c fiber.Ctx) error {
	form, _ := form_builder.GenerateForm(
		"POST",
		"/admin/user",
		"create",
		dbtype.User{},
		"Create User",
		"Create User",
	)
	form.Confirm = "Save User Config?"

	return c.Render("admin/users", fiber.Map{
		"form": form,
	})
}

func Handle_Users_Create_Crud(c fiber.Ctx) error {
	dto := new(dbtype.User)
	if err := c.Bind().Body(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Create: Failed %s", err),
		})
	}

	user, err := repository.PutUser(dto)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Create: Failed %s", err),
		})
	}

	repository.PutSystemEvent(&dbtype.SystemEvent{
		Severity:   "info",
		Message:    fmt.Sprintf("User Create: Success %s", user.Name),
		Ref_Source: "system",
	})

	return c.JSON(fiber.Map{
		"status":   "success",
		"redirect": "/admin/users",
		"message":  fmt.Sprintf("User Create: Success %s", user.Name),
	})
}

func Handle_Users_Update_View(c fiber.Ctx) error {
	id := c.Params("id")

	user, err := repository.GetUser(id)
	if err != nil || user == nil {
		return c.Redirect().Status(fiber.StatusFound).To("/admin/users")
	}

	form, _ := form_builder.GenerateForm(
		"POST",
		"/admin/user/"+user.ID.String(),
		"update",
		user,
		"Update User",
		"Update User",
	)
	form.Confirm = "Save User Config?"

	return c.Render("admin/users", fiber.Map{
		"form": form,
	})
}

func Handle_Users_Update_Crud(c fiber.Ctx) error {
	id := c.Params("id")

	dto := new(dbtype.User)
	if err := c.Bind().Body(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Update: Failed %s", err),
		})
	}

	// Fetch user
	user, err := repository.GetUser(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Update: Failed %s", err),
		})
	}

	// Prop Merge
	dto.ID = user.ID
	dto.Date_Created = user.Date_Created
	dto.Date_LastLogin = user.Date_LastLogin

	if err := struct_merge.Merge[dbtype.User](user, dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Update: Failed %s", err),
		})
	}
	// Set
	_, err = repository.SetUser(user.ID.String(), user)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Update: Failed %s", err),
		})
	}

	repository.PutSystemEvent(&dbtype.SystemEvent{
		Severity:   "info",
		Message:    fmt.Sprintf("User Update: Success %s", user.Name),
		Ref_Source: "system",
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("User Update: Success %s", user.Name),
	})
}

func Handle_Users_Delete_Crud(c fiber.Ctx) error {
	id := c.Params("id")

	// Fetch user
	user, err := repository.GetUser(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Update: Failed %s", err),
		})
	}

	err = repository.DelUser(user.ID.String(), user)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User Update: Failed %s", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":   "success",
		"redirect": "/admin/users",
		"message":  fmt.Sprintf("User Delete: Success %s", user.Name),
	})
}
