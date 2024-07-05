package handler

import (
	"math"
	"strconv"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/k8"

	"github.com/gofiber/fiber/v2"
)

func Handle_Admin_Root(c *fiber.Ctx) error {
	// Get the cluster status
	cluster, cluster_err := k8.GetClusterStatus()

	error_string := "Cluster is not running"
	if cluster_err != nil {
		error_string = cluster_err.Error()
	}

	return c.Render("admin/index", fiber.Map{
		"cluster": map[string]interface{}{
			"version": cluster["version"],
			"nodes":   cluster["nodes"],
			"error":   error_string,
		},
	})
}

func Handle_Admin_Users(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 5
	users, total, error := repository.AllUser(perPage, curPage)
	if error != nil {
		return c.SendString(error.Error())
	}

	return c.Render("admin/users", fiber.Map{
		"users": users,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

func Handle_Admin_Lobbies(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 5
	lobbies, total, error := repository.AllLobby(perPage, curPage)
	if error != nil {
		return c.SendString(error.Error())
	}

	return c.Render("admin/lobbies", fiber.Map{
		"lobbies": lobbies,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
	})
}

func Handle_Admin_Servers(c *fiber.Ctx) error {

	return c.Render("admin/servers", fiber.Map{})
}

func Handle_Admin_Pods(c *fiber.Ctx) error {
	// pods, _ := k8.GetAllServerPods()
	node, port, err := k8.FindOpenNodePort()

	// Create a new pod

	return c.JSON(fiber.Map{
		"node": node,
		"port": port,
		"err":  err,
	})

	if err != nil {
		return c.SendString(err.Error())
	}

	return c.Render("admin/k8_pods", fiber.Map{
		// "iPods": pods,
	})
}

// ##> Admin API
func Handle_Admin_Cluster_Start(c *fiber.Ctx) error {
	err := k8.CheckCreateCluster()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{})
}

func Handle_Admin_Cluster_Stop(c *fiber.Ctx) error {
	err := k8.CheckDeleteCluster()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{})
}

func Handle_Admin_Cluster_Pods_Stop(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{})
}
