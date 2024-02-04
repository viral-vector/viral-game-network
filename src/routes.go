package main

import (
	"github.com/gofiber/fiber/v2"
	"viral-game-network/src/handler"
)

func serve_routes(app *fiber.App) {
	// Root Routes
	app.Get("/", handler.HandleRoot)

	// Lobby Routes
	app.Get("/lobby", handler.Handle_AllLobby)
	app.Get("/lobby/{ID}", handler.Handle_GetLobby)
	app.Post("/lobby", handler.Handle_SetLobby)
}