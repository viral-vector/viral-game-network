package handler

import (
	"encoding/json"
	"strconv"
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
	Guid string `json:"guid" xml:"guid" form:"guid"`
	App string  `json:"app" xml:"guid" form:"guid"`
}

// Handle_AllLobby
func Handle_AllLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/all")
	curPage, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("show", "100"))

	data, total, err := repository.AllLobby(perPage, curPage)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching lobbies",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data":   data,
		"total":  total,
	})
}

// Handle_SetLobby
func Handle_SetLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/set")

	record := new(dbtype.Lobby)
	err := json.Unmarshal(c.Body(), record)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Bad set lobby request",
		})
	}

	lobby, err := repository.SetLobby(c.Params("id"), record)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error setting lobby",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data":   lobby,
	})
}

// Handle_GetLobby
func Handle_GetLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/get")

	lobby, err := repository.GetLobby(c.Params("id"))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching lobby",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data":   lobby,
	})
}

// Handle_JoinLobby
func Handle_JoinLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/join")

	lobby, err := repository.GetLobby(c.Params("id"))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching lobby",
		})
	}

	max_players, err := strconv.ParseInt(lobby.Lobby_Application.Lobby_Max_Players, 0, 0)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Server error fetching lobby max players",
		})
	}

	if len(lobby.Lobby_Users) >= int(max_players) {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Lobby is full",
		})
	}

	if lobby.Private && lobby.Code != c.Params("code") {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid lobby code",
		})
	}

	req_user := c.Locals("user").(*dbtype.User)

	// for _, user := range lobby.Lobby_Users {
	// 	if user.User.ID == req_user.ID {
	// 		c.Status(fiber.StatusBadRequest)
	// 		return c.JSON(fiber.Map{
	// 			"status": "error",
	// 			"message": "Already joined lobby",
	// 		})
	// 	}
	// }

	// Link User to lobby
	err = repository.LinkLobbyUser(lobby, req_user)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Error joining lobby: " + err.Error(),
		})
	}

	lobby, _ = repository.GetLobby(lobby.ID.String())

	return c.JSON(fiber.Map{
		"status": "ok",
		"data":   lobby,
	})
}

// Handle_HostLobby
func Handle_HostLobby(c *fiber.Ctx) error {
	c.Set("Viral-Game-Network-Action", "lobby/host")

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
	app, err := repository.GetApplication(&dbtype.Application{
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
		"status": "ok",
		"data":   lobby,
	})
}

// Handle_SocketLobby
func Handle_SocketLobby(c *websocket.Conn) {
	id := c.Params("id")
	guid := "lobby:" + id + ":channel"
	user := c.Locals("user").(*dbtype.User)
	var (
		msg []byte
		err error
	)

	// Sync
	channel_cache, _ := cache.Get[interface{}](guid)
	if channel_cache != nil {
		go func(c *websocket.Conn) {
			for _, element := range channel_cache.([]string) {
				time.Sleep(100 * time.Millisecond)
				if err = c.WriteMessage(1, []byte(element)); err != nil {
					return
				}
			}
		}(c)
	}

	// Sub
	sb := pubsub.Sub(guid)
	go func(c *websocket.Conn) {
		for msg := range sb.Channel() {
			if err = c.WriteMessage(1, []byte(msg.Payload)); err != nil {
				defer pubsub.Close(sb)
				return
			}
		}
	}(c)

	defer pubsub.Close(sb)

	// Pub
	for {
		if _, msg, err = c.ReadMessage(); err != nil {
			return
		}
		if msg == nil || len(string(msg)) < 3 {
			if err = c.WriteMessage(1, []byte("{\"error\":\"Message length should >= 3\"}")); err != nil {
				return
			}
			continue
		}

		ministration.Service_Lobby_Notify(id, string(msg), user.ID.String())
	}
}
