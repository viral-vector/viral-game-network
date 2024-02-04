package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create a new Fiber app
	app := fiber.New()
	// Register routes
	serve_routes(app)
	// Serve static files from the public folder
	app.Static("/", "./../public")
	// Log Errors
    log.Fatal(app.Listen(":8080"))
}