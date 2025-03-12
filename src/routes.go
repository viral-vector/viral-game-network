package main

import (
	"os"
	"time"
	"viral-game-network/src/handler"
	handler_admin "viral-game-network/src/handler/admin"
	"viral-game-network/src/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/storage/redis/v3"
)

func serve_routes(app *fiber.App) {
	app.Use(helmet.New())
	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "America/New_York",
	}))
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Viral-Game-Network-Entity", c.Get("Viral-Game-Network-Entity"))
		c.Locals("route", c.Path())
		c.Locals("title", "VGN - "+c.Path())
		return c.Next()
	})

	// Root Routes
	app.Get("/", handler.Handle_Root)
	app.Get("/health", handler.Handle_Health)

	/**
	 * 	Auth Routes
	 */
	auth := app.Group("/auth")
	// App Key Validation
	auth.Use(handler.Handle_ValidateAppKey)
	// User Authentication
	auth.Post("/lobby", handler.Handle_AuthLobby)
	auth.Post("/guest", handler.Handle_AuthGuest)

	/**
	 * 	Admin Routes
	 */
	admin := app.Group("/admin")
	// App Key Validation
	// admin.Use(handler.Handle_ValidateAppKey)
	// User Authentication
	admin.Get("/", handler_admin.Handle_Dash)
	admin.Get("/users", handler_admin.Handle_Users)
	admin.Get("/lobbies", handler_admin.Handle_Lobbies)
	admin.Get("/servers", handler_admin.Handle_Servers)
	admin.Get("/pods", handler_admin.Handle_Pods)
	admin.Delete("/pods/:id", handler_admin.Handle_Pods_Delete)
	admin.Get("/pods/:id", handler_admin.Handle_Pods_Edit)
	admin.Get("/metrics", handler_admin.Handle_Dash)
	admin.Get("/configs", handler_admin.Handle_Dash)
	admin.Get("/ssevents", handler_admin.Handle_SSEvents)

	admin.Post("/cluster/start", handler_admin.Handle_Cluster_Start)
	admin.Post("/cluster/stop", handler_admin.Handle_Cluster_Stop)
	admin.Post("/cluster/pods/stop", handler_admin.Handle_Cluster_Pods_Stop)

	/**
	 * 	API Routes
	 */
	gapi := app.Group("/api")

	// Token Validation
	gapi.Use(handler.Handle_ValidateToken)

	// Rate Limit
	gapi.Use(limiter.New(limiter.Config{
		Max:               15,
		Expiration:        3 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		Storage: redis.New(redis.Config{
			URL: "redis://root:@" + os.Getenv("CACHE_ENDPOINT") + "/0",
		}),
	}))

	// Lobby Routes
	grp_lobby := gapi.Group("/lobby")
	grp_lobby.Get("", handler.Handle_AllLobby)
	grp_lobby.Post("/host", handler.Handle_HostLobby)
	grp_lobby.Post("/:id/join", handler.Handle_JoinLobby)
	grp_lobby.Get("/:id/socket", middleware.Lobby_Read, websocket.New(handler.Handle_SocketLobby))
	grp_lobby.Post("/:id", middleware.Lobby_Write, handler.Handle_SetLobby)
	grp_lobby.Get("/:id", handler.Handle_GetLobby)

	// Host/Server Routes
	grp_host := gapi.Group("/host")
	grp_host.Get("/:id/tick", handler.Handle_TickHost)
}
