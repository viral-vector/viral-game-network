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

// ##> 
func Handle_Dash(c *fiber.Ctx) error {
	// Get the cluster status
	cluster, err := k8.GetClusterStatus()
	cluster_error := "Cluster is not running"
	if err != nil {
		cluster_error = err.Error()
	}
	// Get cluster logs
	ctlines := int64(100)
	cevents, _ := k8.GetLogsCluster(&ctlines)

	// Get SystemEvent
	sevents, _ := repository.GetSystemEvents(ctlines)

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
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	servers, total, _ := repository.AllServer(perPage, curPage)

	return c.Render("admin/servers", fiber.Map{
		"servers": servers,
		"total":   total,
		"pages":   int(math.Ceil(float64(total) / float64(perPage))),
		"paged":   curPage,
	})
}

// ##> Pods
func Handle_Pods(c *fiber.Ctx) error {
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 15
	pods, _ := k8.GetAllServerPods()
	total := len(pods)

	return c.Render("admin/pods", fiber.Map{
		"pods":  pods,
		"total": total,
		"pages": int(math.Ceil(float64(total) / float64(perPage))),
		"paged": curPage,
	})
}

func Handle_Pods_Update_View(c *fiber.Ctx) error {
	node, pod, err := k8.LocateServerPod(
		strings.Replace(c.Params("id"), "server-", "", -1))

	if err != nil {
		return c.Redirect("/admin/pods")
	}

	tailLines := int64(100)
	logs, _ := k8.GetPodLogParts(pod, &tailLines)

	return c.Render("admin/pods", fiber.Map{
		"pod":   pod,
		"node":  node,
		"logs":  logs,
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

// ##> Configs
func Handle_Configs(c *fiber.Ctx) error {
	configs, _ := repository.GetSystemConfigs()

	return c.Render("admin/configs", fiber.Map{
		"configs": configs,
	})
}

func Handle_Configs_Update_Crud(c *fiber.Ctx) error {
	configs, _ := repository.GetSystemConfigs()
	for i := range configs {
		configs[i].Val = c.FormValue(configs[i].Key)
	}

	repository.PopSystemConfigs(&configs)

	systemEvent := dbtype.SystemEvent{
		Severity: "info",
		Message:  "Settings Save: Success",
		Ref_Source: "system",
	}
	repository.PutSystemEvent(&systemEvent)

	return c.JSON(fiber.Map{
		"message": "Settings Saved",
	})
}

// ##> Admin JSON
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