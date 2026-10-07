package handler_admin

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	form_builder "viral-game-network/src/utils/form_builder"

	"github.com/gofiber/fiber/v3"
)

type ApiCreateCrudDTO struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ##> Configs
func Handle_Configs(c fiber.Ctx) error {
	configs, _ := repository.GetSystemConfigs()

	// Clear cache
	repository.DelApiKeyCache()
	repository.DelSystemConfigsCache()

	fields := []form_builder.FormField{}
	for _, cnfg := range configs {
		fields = append(fields, form_builder.FormField{
			Name:      cnfg.Key,
			Label:     cnfg.Name,
			Type:      cnfg.Type,
			Value:     cnfg.Val,
			Options:   cnfg.Options,
			SortOrder: cnfg.SortOrder,
			ReadOnly:  cnfg.ReadOnly,
			Required:  true,
		})
	}

	form := form_builder.Form{
		Action:   "/admin/configs",
		Method:   "POST",
		Title:    "System Configurations",
		Fields:   fields,
		Confirm:  "Save Configurations?",
		Submit:   "Save",
		CanReset: true,
	}

	// Api Keys
	keys, _ := repository.GetApiKeys()

	return c.Render("admin/configs", fiber.Map{
		"form": form,
		"keys": keys,
	})
}

func Handle_Configs_Update_Crud(c fiber.Ctx) error {
	configs, err := repository.GetSystemConfigs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not load settings"})
	}
	form, err := c.MultipartForm()
	for i := range configs {
		present := c.Request().PostArgs().Has(configs[i].Key)
		if err == nil && form != nil {
			_, present = form.Value[configs[i].Key]
		}
		if !present || configs[i].ReadOnly {
			continue
		}
		value := c.FormValue(configs[i].Key)
		if configs[i].Key == "VNET_TOKEN_EXPIRE" && value != "" {
			minutes, err := strconv.ParseInt(value, 10, 64)
			if err != nil || minutes <= 0 || minutes > int64((1<<63-1)/time.Minute) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Token lifetime must be a positive number of minutes"})
			}
		}
		configs[i].Val = value
	}
	if _, err := repository.PopSystemConfigs(&configs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not save settings"})
	}
	repository.PutSystemEvent(&dbtype.SystemEvent{Severity: "info", Message: "Settings Save: Success", Ref_Source: "system"})
	return c.JSON(fiber.Map{"status": "success", "message": "Settings Saved"})
}

func Handle_Configs_ApiKey_Create_Crud(c fiber.Ctx) error {
	dto := new(ApiCreateCrudDTO)
	if err := c.Bind().Body(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Api Key Create: %s", err),
		})
	}

	if dto.Key == "" {
		dto.Key = auth.GenerateAppKey(20)
	}

	if err := repository.AddApiKey(dto.Name, dto.Key); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Api Key Create: %s", err),
		})
	}

	// Clear cache
	repository.DelApiKeyCache()
	repository.DelSystemConfigsCache()

	config, err := repository.SelConfig("API_KEY_" + dto.Name)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Api Key Create x: %s", err),
		})
	}

	var buf bytes.Buffer
	// Render the template into the buffer
	if err := c.App().Config().Views.Render(&buf, "admin/partials/configs/list_item_api_key", config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Api Key Saved",
		"render":  buf.String(),
	})
}

func Handle_Configs_Delete_Crud(c fiber.Ctx) error {
	id := c.Params("id")

	// Fetch Config
	config, err := repository.GetConfig(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Config Delete: %s", err),
		})
	}

	err = repository.DelConfig(config.ID.String(), config)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Config Delete: %s", err),
		})
	}

	// Clear cache
	repository.DelApiKeyCache()
	repository.DelSystemConfigsCache()

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Config Delete: Success %s", ""),
	})
}
