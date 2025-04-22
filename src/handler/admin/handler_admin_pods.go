package handler_admin

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"viral-game-network/src/k8"
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/database/repository"
)

func Handle_Pods(c *fiber.Ctx) error {
	search := c.Query("search", "")
	prvPage, _ := strconv.Atoi(c.Query("prevPage", "1"))
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := int64(2)
	pageToken := c.Query("pageToken", "")
	pods, total, continueToken, _ := k8.GetAllServerPodsPager(search, perPage, prvPage, curPage, pageToken)

	fmt.Println("Pods: ", len(pods), continueToken)

	return c.Render("admin/pods", fiber.Map{
		"pods":  pods,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
		"pageToken": continueToken,
		"search": search,
	})
}

func Handle_Pods_Update_View(c *fiber.Ctx) error {
	node, pod, err := k8.LocateServerPod(c.Params("id"))

	if err != nil || pod == nil {
		return c.Redirect("/admin/pods")
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

func Handle_Pods_Delete_Crud(c *fiber.Ctx) error {
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
