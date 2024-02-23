package main

import (
	"os"
	"log"
	"viral-game-network/src/job"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		Prefork:       os.Getenv("APP_PREFORK") == "true",
		CaseSensitive: true,
		StrictRouting: false,
		ServerHeader:  "VGN",
		AppName: os.Getenv("APP_NAME"),
	})
	// Register routes
	serve_routes(app)
	// Serve static files from the public folder
	app.Static("/", "./../public")

	// Start the job scheduler
	job.Start()

	// Log Errors
    log.Fatal(app.Listen(":3000"))
}