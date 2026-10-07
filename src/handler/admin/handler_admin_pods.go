package handler_admin

import (
	"github.com/gofiber/fiber/v3"
	"math"
	"strings"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/k8"
	"viral-game-network/src/utils/pagination"
)

func Handle_Pods(c fiber.Ctx) error {
	search := c.Query("search", "")
	perPage := int64(2)
	prvPage, err := pagination.Page(c.Query("prevPage", "1"), int(perPage))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid previous page"})
	}
	curPage, err := pagination.Page(c.Query("page", "1"), int(perPage))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid page"})
	}
	pageToken := c.Query("pageToken", "")
	pods, total, continueToken, err := k8.GetAllServerPodsPager(search, perPage, prvPage, curPage, pageToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not load pods"})
	}

	return c.Render("admin/pods", fiber.Map{
		"pods":      pods,
		"total":     total,
		"pages":     int(math.Ceil(float64(total) / float64(perPage))),
		"paged":     curPage,
		"pageToken": continueToken,
		"search":    search,
	})
}

func Handle_Pods_Update_View(c fiber.Ctx) error {
	node, pod, err := k8.LocateServerPod(c.Params("id"))

	if err != nil || pod == nil {
		return c.Redirect().Status(fiber.StatusFound).To("/admin/pods")
	}

	lobby, _ := repository.GetLobby("Lobby:"+pod.Name)

	tailLines := int64(100)
	logs, _ := k8.GetPodLogParts(pod, &tailLines)

	return c.Render("admin/pods", fiber.Map{
		"pod":   pod,
		"node":  node,
		"logs":  logs,
		"lobby": lobby,
	}) 
}

func Handle_Pods_Delete_Crud(c fiber.Ctx) error {
	err := k8.DeleteServerPod(
		strings.Replace(c.Params("id"), "server-", "", -1))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"redirect": "/admin/pods",
		"message":  "Deleted " + c.Params("id"),
	})
}
