package handler_admin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func Handle_Dash(c *fiber.Ctx) error {
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

func Handle_Users(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	users, total, _ := repository.AllUser(perPage, curPage)

	return c.Render("admin/users", fiber.Map{
		"users": users,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

func Handle_Lobbies(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	lobbies, total, _ := repository.AllLobby(perPage, curPage)

	return c.Render("admin/lobbies", fiber.Map{
		"lobbies": lobbies,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
	})
}

func Handle_Servers(c *fiber.Ctx) error {

	return c.Render("admin/servers", fiber.Map{})
}

func Handle_Pods(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	pods, _ := k8.GetAllServerPodsAndServices()
	total := len(pods)

	return c.Render("admin/pods", fiber.Map{
		"pods":  pods,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

func Handle_Pods_Create(c *fiber.Ctx) error {
	node, port, err := k8.FindOpenNodePort()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create the server pod
	cmd := []string{}
	pod, service, err := k8.CreateServerPod("nginx", node, int32(port), int32(80), "nginx", cmd)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"pod":     pod,
		"service": service,
	})
}

func Handle_Pods_Edit(c *fiber.Ctx) error {
	node, pod, service, err := k8.LocateServerPod(
		strings.Replace(c.Params("id"), "server-", "", -1))

	if err != nil {
		return c.Redirect("/admin/pods")
	}

	return c.Render("admin/pods", fiber.Map{
		"pod":     pod,
		"node":    node,
		"service": service,
	})
}

func Handle_Pods_Delete(c *fiber.Ctx) error {
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

// ##> Admin API
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

func Handle_Cluster_Start(c *fiber.Ctx) error {
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
