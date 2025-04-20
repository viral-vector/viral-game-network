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
		c.Locals("title", "VGN - " + c.Path())
		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		route := c.Route()
		if route.Name != "" {
			c.Set("Viral-Game-Network-Action", route.Name)
		}
		return err
	})

	// Root Routes
	app.Get("/", handler.Handle_Root).Name("root")
	app.Get("/health", handler.Handle_Health).Name("health")

	/**
	 * 	Auth Routes
	 */
	auth := app.Group("/auth")
	// User Authentication
	auth.Use([]string{"/lobby", "/guest"}, handler.Handle_ValidateAppKey)
	auth.Post("/lobby", handler.Handle_AuthLobby).Name("auth/lobby")
	auth.Post("/guest", handler.Handle_AuthGuest).Name("auth/guest")
	// Admin Authentication
	auth.Get("/admin", handler.Handle_RedirectAdmin, handler.Handle_AuthAdmin).Name("auth/admin")
	auth.Post("/admin", handler.Handle_RedirectAdmin, handler.Handle_AuthAdmin).Name("auth/admin/submit")
	auth.Get("/admin/refresh", handler.Handle_ValidateTokenAdmin, handler.Handle_AuthAdminRefresh).Name("auth/admin/refresh")
	auth.Get("/admin/logout", handler.Handle_ValidateTokenAdmin, handler.Handle_AuthAdminLogout).Name("auth/admin/logout")

	/**
	 * 	Admin Routes -----------------------------------------------------------------  
	 */
	admin := app.Group("/admin")
	// Rate Limit
	admin.Use(limiter.New(limiter.Config{
		Max:               30,
		Expiration:        1 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		Storage: redis.New(redis.Config{
			URL: "redis://root:@" + os.Getenv("CACHE_ENDPOINT") + "/0",
		}),
	}))
	// Token Validation
	admin.Use(handler.Handle_ValidateTokenAdmin)
	
	admin.Get("/", handler_admin.Handle_Dash).Name("admin")
	// Cluster
	admin.Post("/cluster/start", handler_admin.Handle_Cluster_Start)
	admin.Post("/cluster/stop", handler_admin.Handle_Cluster_Stop)
	admin.Post("/cluster/pods/stop", handler_admin.Handle_Cluster_Pods_Stop)
	// Admins
	admin.Get("/admins", handler_admin.Handle_Admins).Name("admin/admins")
	admin.Get("/admin", handler_admin.Handle_Admins_Create_View)
	admin.Post("/admin", handler_admin.Handle_Admins_Create_Crud)
	admin.Get("/admin/:id", handler_admin.Handle_Admins_Update_View)
	admin.Post("/admin/:id", handler_admin.Handle_Admins_Update_Crud)
	admin.Delete("/admin/:id", handler_admin.Handle_Admins_Delete_Crud)
	// Aplications
	admin.Get("/applications", handler_admin.Handle_Applications).Name("admin/apllications")
	admin.Get("/application", handler_admin.Handle_Applications_Create_View)
	admin.Post("/application", handler_admin.Handle_Applications_Create_Crud)
	admin.Get("/application/:id", handler_admin.Handle_Applications_Update_View)
	admin.Post("/application/:id", handler_admin.Handle_Applications_Update_Crud)
	admin.Delete("/application/:id", handler_admin.Handle_Applications_Delete_Crud)
	// Users
	admin.Get("/users", handler_admin.Handle_Users).Name("admin/users")
	admin.Get("/user", handler_admin.Handle_Users_Create_View)
	admin.Post("/user", handler_admin.Handle_Users_Create_Crud)
	admin.Get("/user/:id", handler_admin.Handle_Users_Update_View)
	admin.Post("/user/:id", handler_admin.Handle_Users_Update_Crud)
	admin.Delete("/user/:id", handler_admin.Handle_Users_Delete_Crud)
	// Lobbies
	admin.Get("/lobbies", handler_admin.Handle_Lobbies).Name("admin/lobbies")
	admin.Get("/lobby", handler_admin.Handle_Lobbies_Create_View)
	admin.Post("/lobby", handler_admin.Handle_Lobbies_Create_Crud)
	admin.Get("/lobby/:id", handler_admin.Handle_Lobbies_Update_View)
	admin.Post("/lobby/:id", handler_admin.Handle_Lobbies_Update_Crud)
	admin.Delete("/lobby/:id", handler_admin.Handle_Lobbies_Delete_Crud)
	// Servers
	admin.Get("/servers", handler_admin.Handle_Servers).Name("admin/servers")
	admin.Get("/server/:id", handler_admin.Handle_Servers_Update_View)
	admin.Delete("/server/:id", handler_admin.Handle_Servers_Delete_Crud)
	// Pods
	admin.Get("/pods", handler_admin.Handle_Pods).Name("admin/pods")
	admin.Get("/pod/:id", handler_admin.Handle_Pods_Update_View)
	admin.Delete("/pod/:id", handler_admin.Handle_Pods_Delete_Crud)
	// Metrics Configs Etc
	admin.Get("/metrics", handler_admin.Handle_Metrics).Name("admin/metrics")
	admin.Get("/events", handler_admin.Handle_Events).Name("admin/events")
	admin.Get("/configs", handler_admin.Handle_Configs).Name("admin/configs")
	admin.Post("/configs", handler_admin.Handle_Configs_Update_Crud)
	admin.Get("/ssevents", handler_admin.Handle_SSEvents).Name("admin/ssevents")
	// -------------------------------------------------------------------------------

	/**
	 * 	API Routes
	 */
	gapi := app.Group("/api")

	// Token Validation
	gapi.Use(handler.Handle_ValidateTokenUsers)

	// Rate Limit
	gapi.Use(limiter.New(limiter.Config{
		Max:               100,
		Expiration:        1 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		Storage: redis.New(redis.Config{
			URL: "redis://root:@" + os.Getenv("CACHE_ENDPOINT") + "/0",
		}),
	}))

	// Lobby Routes
	grp_lobby := gapi.Group("/lobby")
	grp_lobby.Get("", handler.Handle_AllLobby).Name("lobby/all")
	grp_lobby.Post("/host", handler.Handle_HostLobby).Name("lobby/host")
	grp_lobby.Post("/:id/join", handler.Handle_JoinLobby).Name("lobby/join")
	grp_lobby.Get("/:id/socket", middleware.Lobby_Read, websocket.New(handler.Handle_SocketLobby)).Name("lobby/socket")
	grp_lobby.Post("/:id", middleware.Lobby_Write, handler.Handle_SetLobby).Name("lobby/put")
	grp_lobby.Get("/:id", handler.Handle_GetLobby).Name("lobby/get")

	// Host/Server Routes
	grp_host := gapi.Group("/host")
	grp_host.Get("/:id/tick", handler.Handle_TickHost).Name("host/tick")
}
