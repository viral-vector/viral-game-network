package main

import (
	"os"
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		Prefork:       os.Getenv("APP_PREFORK") == "true",
		CaseSensitive: true,
		StrictRouting: false,
		ServerHeader:  "Fiber",
		AppName: os.Getenv("APP_NAME"),
	})
	// Register routes
	serve_routes(app)
	// Serve static files from the public folder
	app.Static("/", "./../public")
	// Log Errors
    log.Fatal(app.Listen(":8080"))
}