package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/ministration"
	"viral-game-network/src/pubsub"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type HostLobbyRequestDTO struct {
	Name string `json:"name" xml:"name" form:"name"`
	App  string `json:"app" xml:"guid" form:"app"`
}

// Handle_AllLobby
func Handle_AllLobby(c *fiber.Ctx) error {
	search := c.Query("search", "")
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("show", "100"))

	data, total, err := repository.AllLobby(perPage, curPage, search)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching lobbies",
		})
	}

	user := c.Locals("user").(*dbtype.User)
	for i := range data {
		redactLobbyCode(&data[i], user)
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   data,
		"total":  total,
	})
}

// Handle_SetLobby
func Handle_SetLobby(c *fiber.Ctx) error {
	record := new(repository.LobbyPatch)
	err := json.Unmarshal(c.Body(), record)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Bad set lobby request",
		})
	}

	lobby, err := repository.PatchLobby(c.Params("id"), record)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error setting lobby",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   lobby,
	})
}

// Handle_GetLobby
func redactLobbyCode(lobby *dbtype.Lobby, user *dbtype.User) {
	if lobby != nil && (lobby.Lobby_Host == nil || user == nil || lobby.Lobby_Host.ModelID() != user.ModelID()) {
		lobby.Code = ""
	}
}

func Handle_GetLobby(c *fiber.Ctx) error {
	lobby, err := repository.GetLobby(c.Params("id"))
	if err != nil || lobby == nil {
		return fiber.ErrBadRequest
	}
	redactLobbyCode(lobby, c.Locals("user").(*dbtype.User))
	return c.JSON(fiber.Map{"status": "success", "data": lobby})
}

// Handle_JoinLobby accepts an optional JSON/form invitation code.
func Handle_JoinLobby(c *fiber.Ctx) error {
	var dto struct {
		Code string `json:"code" form:"code"`
	}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&dto); err != nil {
			return fiber.ErrBadRequest
		}
	}
	user := c.Locals("user").(*dbtype.User)
	lobby, err := repository.JoinLobby(c.Params("id"), user, dto.Code)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error joining lobby: " + err.Error()})
	}
	redactLobbyCode(lobby, user)
	return c.JSON(fiber.Map{"status": "success", "data": lobby})
}

// Handle_HostLobby
func Handle_HostLobby(c *fiber.Ctx) error {
	dto := new(HostLobbyRequestDTO)
	if err := c.BodyParser(dto); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Bad lobby host request",
		})
	}

	user := c.Locals("user").(*dbtype.User)

	// Fetch the app
	app, err := repository.SelApplication(&dbtype.Application{
		Guid: dto.App,
		Name: dto.App,
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error hosting lobby: " + err.Error(),
		})
	}

	// Insert Lobby
	lobby, err := repository.PutLobby(&dbtype.Lobby{
		Name: dto.Name,
	}, app, user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error hosting lobby: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   lobby,
	})
}

// Handle_SocketLobby
func Handle_SocketLobby(c *websocket.Conn) {
	id := c.Params("id")
	channel := "lobby:" + id + ":channel"
	user := c.Locals("user").(*dbtype.User)
	conn := c.Conn
	defer conn.Close()
	expires, ok := c.Locals("session_expires").(time.Time)
	if !ok || !expires.After(time.Now()) {
		return
	}
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(expires)

	subscription := pubsub.Sub(channel)
	defer pubsub.Close(subscription)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err := subscription.Receive(ctx)
	cancel()
	if err != nil {
		return
	}

	// Redis stores history as a list. Replay it before starting live writes.
	history, err := cache.List(channel)
	if err != nil {
		return
	}
	for _, message := range history {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			return
		}
	}

	// WebSocket connections allow only one writer at a time.
	var writer sync.Mutex
	write := func(message []byte) error {
		writer.Lock()
		defer writer.Unlock()
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		return conn.WriteMessage(websocket.TextMessage, message)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for message := range subscription.Channel() {
			if err := write([]byte(message.Payload)); err != nil {
				conn.Close()
				return
			}
		}
	}()
	defer func() {
		pubsub.Close(subscription)
		conn.Close()
		<-done
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if len(message) < 3 {
			if err := write([]byte(`{"error":"Message length should >= 3"}`)); err != nil {
				return
			}
			continue
		}
		if err := ministration.Service_Lobby_Notify(id, string(message), user.ID.String()); err != nil {
			return
		}
	}
}
