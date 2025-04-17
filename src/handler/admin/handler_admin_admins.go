package handler_admin

import (
	"fmt"
	"math"
	"strconv"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"
	struct_merge "viral-game-network/src/utils/struct_merge"

	"github.com/gofiber/fiber/v2"
)

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
		"create",  
		dbtype.Admin{},
		"Create Admin",
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
		"redirect": "/admin/admins",
		"message": fmt.Sprintf("Admin Create: Success %s", admin.Name),
	})
}

func Handle_Admins_Update_View(c *fiber.Ctx) error {
	id := c.Params("id")

	admin, err := repository.GetAdmin(id)
	if err != nil || admin == nil {
		return c.Redirect("/admin/admins")
	}

	admin.Password = ""

	form, _ := form_builder.GenerateForm(
		"POST", 
		"/admin/admin/" + admin.ID.String(),
		"update", 
		admin,
		"Update Admin",
		"Update Admin",
	)
	form.Confirm = "Save Admin Config?"

	return c.Render("admin/admins", fiber.Map{
		"form" : form,
	})
}

func Handle_Admins_Update_Crud(c *fiber.Ctx) error {
	id := c.Params("id")

	dto := new(dbtype.Admin)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Update: Failed %s", err),
		})
	}

	// Fetch admin
	admin, err := repository.GetAdmin(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Update: Failed %s", err),
		})
	}

	// Password & Prop Merge
	dto.ID = admin.ID
	dto.Date_Created = admin.Date_Created
	dto.Date_LastLogin = admin.Date_LastLogin

	if dto.Password != "" {
		password, err := auth.HashGenerate(dto.Password)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Admin Update: Failed %s", err),
			})
		}
		dto.Password = password
	}else{
		dto.Password = admin.Password
	} 
	if err := struct_merge.Merge[dbtype.Admin](admin, dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Update: Failed %s", err),
		})
	}
	// Set
	_, err = repository.SetAdmin(admin.ID.String(), dto)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Update: Failed %s", err),
		})
	}

	repository.PutSystemEvent(&dbtype.SystemEvent{
		Severity: "info",
		Message:  fmt.Sprintf("Admin Update: Success %s", admin.Name) ,
		Ref_Source: "system",
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Admin Update: Success %s", admin.Name),
	})
}

func Handle_Admins_Delete_Crud(c *fiber.Ctx) error {
	id := c.Params("id")

	// Fetch admin
	admin, err := repository.GetAdmin(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Update: Failed %s", err),
		})
	}

	authAdmin := c.Locals("admin").(*dbtype.Admin)
	if authAdmin.ID.String() == admin.ID.String() {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Admin Update: Can't delete self",
		})
	}

	repository.DelAdmin(admin.ID.String(), admin)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Admin Update: Failed %s", err),
		})
	}
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"redirect": "/admin/admins",
		"message": fmt.Sprintf("Admin Delete: Success %s", admin.Name),
	})
}