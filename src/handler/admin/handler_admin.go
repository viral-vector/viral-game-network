package handler_admin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// ##> Dashboard
func Handle_Dash(c *fiber.Ctx) error {
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
func Handle_Cluster_Start(c *fiber.Ctx) error {
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

func Handle_Cluster_Stop(c *fiber.Ctx) error {
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

func Handle_Cluster_Pods_Stop(c *fiber.Ctx) error {
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
func Handle_Events(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 30
	sevents, total, _ := repository.AllSystemEvents(perPage, curPage)

	return c.Render("admin/events", fiber.Map{
		"sevents" : sevents,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

// ##> Metrics
func Handle_Metrics(c *fiber.Ctx) error {
	return c.Render("admin/metrics", fiber.Map{
	
	})
}

// ##> SSEvents
func Handle_SSEvents(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// Create a channel to send messages
	messageChan := make(chan dbtype.SystemEvent)

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for msg := range messageChan {
			// Marshal to JSON
			msg, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			// Send the message
			fmt.Fprintf(w, "data: %s\n\n", msg)
			err = w.Flush()
			if err != nil {
				fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
				return
			}
		}
	}))

	go func() {
		for {
			messages, err := repository.SelSystemEventsInFrame(5)
			if err != nil {
				continue
			}
			for i := range messages {
				messageChan <- messages[i]
			}
			time.Sleep(5 * time.Second)
		}
	}()
	return nil
}