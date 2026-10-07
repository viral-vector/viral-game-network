package handler_admin

import (
	"math"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/src/utils/pagination"

	"github.com/gofiber/fiber/v3"
)

// ##> Dashboard
func Handle_Dash(c fiber.Ctx) error {
	// Get the cluster status
	cluster, err := k8.GetClusterStatus()
	cluster_error := "Cluster is not running"
	if err != nil {
		cluster_error = err.Error()
	}
	// Get cluster logs
	ctlines := int64(3)
	cevents, _ := k8.GetLogsCluster(&ctlines)

	// Get SystemEvent
	sevents, _ := repository.SelSystemEvents(ctlines)

	return c.Render("admin/index", fiber.Map{
		"cluster": map[string]interface{}{
			"version": cluster["version"],
			"nodes":   cluster["nodes"],
			"error":   cluster_error,
		},
		"cevents": cevents,
		"sevents": sevents,
	})
}

// ##> Cluster
func Handle_Cluster_Start(c fiber.Ctx) error {
	go func() {
		err := k8.CheckCreateCluster()
		systemEvent := dbtype.SystemEvent{
			Severity: "info",
			Message:  "Cluster Start: Success",
			Ref_Source: "system",
		}
		if err != nil {
			systemEvent.Message = "Cluster Start: Error @ " + err.Error()
			systemEvent.Severity = "error"
		}
		repository.PutSystemEvent(&systemEvent)
	}()
	return c.JSON(fiber.Map{
		"message": "Starting cluster",
	})
}

func Handle_Cluster_Stop(c fiber.Ctx) error {
	go func() {
		err := k8.CheckDeleteCluster()
		systemEvent := dbtype.SystemEvent{
			Severity: "info",
			Message:  "Cluster Stop: Success",
			Ref_Source: "system",
		}
		if err != nil {
			systemEvent.Message = "Cluster Stop: Error @ " + err.Error()
			systemEvent.Severity = "error"
		}
		repository.PutSystemEvent(&systemEvent)
	}()
	return c.JSON(fiber.Map{
		"message": "Stopping cluster",
	})
}

func Handle_Cluster_Pods_Stop(c fiber.Ctx) error {
	go func() {
		err := k8.KillAllServerPods()
		systemEvent := dbtype.SystemEvent{
			Severity: "info",
			Message:  "Pods Stop: Success",
			Ref_Source: "system",
		}
		if err != nil {
			systemEvent.Message = "Pods Stop: Error @ " + err.Error()
			systemEvent.Severity = "error"
		}
		repository.PutSystemEvent(&systemEvent)
	}()
	return c.JSON(fiber.Map{
		"message": "Stopping all pods",
	})
}

// ##> Events
func Handle_Events(c fiber.Ctx) error {
	perPage := 30
	curPage, err := pagination.Page(c.Query("page", "1"), perPage)
	if err != nil {
		return fiber.ErrBadRequest
	}
	sevents, total, err := repository.AllSystemEvents(perPage, curPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load events")
	}

	return c.Render("admin/events", fiber.Map{
		"sevents": sevents,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
	})
}

// ##> Metrics
func Handle_Metrics(c fiber.Ctx) error {
	return c.Render("admin/metrics", fiber.Map{
	
	})
}
