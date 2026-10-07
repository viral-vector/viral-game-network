package handler

import (
	"context"
	"encoding/json"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/ministration"
	"viral-game-network/src/pubsub"
	"viral-game-network/src/utils/pagination"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

type HostLobbyRequestDTO struct {
	Name string `json:"name" xml:"name" form:"name"`
	App  string `json:"app" xml:"guid" form:"app"`
}

// Handle_AllLobby
func Handle_AllLobby(c fiber.Ctx) error {
	search := c.Query("search", "")
	perPage, err := pagination.Size(c.Query("show", "100"), 100)
	if err != nil {
		return fiber.ErrBadRequest
	}
	curPage, err := pagination.Page(c.Query("page", "1"), perPage)
	if err != nil {
		return fiber.ErrBadRequest
	}

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
func Handle_SetLobby(c fiber.Ctx) error {
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

func Handle_GetLobby(c fiber.Ctx) error {
	lobby, err := repository.GetLobby(c.Params("id"))
	if err != nil || lobby == nil {
		return fiber.ErrBadRequest
	}
	redactLobbyCode(lobby, c.Locals("user").(*dbtype.User))
	return c.JSON(fiber.Map{"status": "success", "data": lobby})
}

// Handle_JoinLobby accepts an optional JSON/form invitation code.
func Handle_JoinLobby(c fiber.Ctx) error {
	var dto struct {
		Code string `json:"code" form:"code"`
	}
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&dto); err != nil {
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
func Handle_HostLobby(c fiber.Ctx) error {
	dto := new(HostLobbyRequestDTO)
	if err := c.Bind().Body(dto); err != nil {
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
	authorize := func() error {
		if !expires.After(time.Now()) {
			return fiber.ErrForbidden
		}
		member, err := repository.IsLobbyMember(id, user.ModelID())
		if err != nil {
			return err
		}
		if !member {
			return fiber.ErrForbidden
		}
		return nil
	}
	send := func(message []byte) error {
		if err := authorize(); err != nil {
			return err
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		return conn.WriteMessage(websocket.TextMessage, message)
	}

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
		if err := send([]byte(message)); err != nil {
			return
		}
	}

	// WebSocket connections allow only one writer at a time.
	write := lobbyLiveWriter(history, send)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			// fasthttp may defer closing a hijacked connection until this
			// handler returns. Wake its blocked reader so it can return.
			conn.SetReadDeadline(time.Now())
			conn.Close()
		}()
		membership := time.NewTicker(time.Second)
		defer membership.Stop()
		messages := subscription.Channel()
		for {
			select {
			case message, open := <-messages:
				if !open || write([]byte(message.Payload)) != nil {
					return
				}
			case <-membership.C:
				// Quiet sockets must also close when their member leaves.
				if authorize() != nil {
					return
				}
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
		if authorize() != nil {
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
