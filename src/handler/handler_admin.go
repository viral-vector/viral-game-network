package handler

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

func Handle_Admin_Dash(c *fiber.Ctx) error {
	// Get the cluster status
	cluster, cluster_err := k8.GetClusterStatus()

	cluster_error_string := "Cluster is not running"
	if cluster_err != nil {
		cluster_error_string = cluster_err.Error()
	}

	// Get SystemEvent
	sevents, _ := repository.GetSystemEvents(10)

	return c.Render("admin/index", fiber.Map{
		"cluster": map[string]interface{}{
			"version": cluster["version"],
			"nodes":   cluster["nodes"],
			"error":   cluster_error_string,
		},
		"sevents": sevents,
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
func Handle_Admin_SSEvents(c *fiber.Ctx) error {
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
			messages, err := repository.GetSystemEventsInFrame(5)
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

func Handle_Admin_Cluster_Start(c *fiber.Ctx) error {
	go func() {
		err := k8.CheckCreateCluster()
		systemEvent := dbtype.SystemEvent{
			Severity: "info",
			Message:  "Cluster Start: Success",
		}
		if err != nil {
			systemEvent.Message = "Cluster Start: Error @ " + err.Error()
			systemEvent.Severity = "error"
		}
		repository.PutSystemEvent(systemEvent)
	}()
	return c.JSON(fiber.Map{
		"message": "Starting cluster",
	})
}

func Handle_Admin_Cluster_Stop(c *fiber.Ctx) error {
	go func() {
		err := k8.CheckDeleteCluster()
		systemEvent := dbtype.SystemEvent{
			Severity: "info",
			Message:  "Cluster Stop: Success",
		}
		if err != nil {
			systemEvent.Message = "Cluster Stop: Error @ " + err.Error()
			systemEvent.Severity = "error"
		}
		repository.PutSystemEvent(systemEvent)
	}()
	return c.JSON(fiber.Map{
		"message": "Stopping cluster",
	})
}

func Handle_Admin_Cluster_Pods_Stop(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{})
}
