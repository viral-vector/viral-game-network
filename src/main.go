package main

import (
	"log"
	"os"
	"viral-game-network/src/job"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/pug/v2"
)

func main() {
	ServeApp()
}

func ServeApp() {
	// Create a new engine
	engine := pug.New("./views", ".pug")
	engine.Reload(true)
	engine.AddFunc("equals", func(a any, b any) bool {
		return a == b
	})
	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		Prefork:           os.Getenv("APP_PREFORK") == "true",
		CaseSensitive:     true,
		StrictRouting:     false,
		ServerHeader:      "VGN",
		AppName:           os.Getenv("VNET_NAME"),
		Views:             engine,
		ViewsLayout:       "base",
		PassLocalsToViews: true,
	})

	// bootstrap the application
	go bootstrap()
	// Serve static files from the public folder
	app.Static("/", "./public")
	// Register routes
	serve_routes(app)
	// Start the job scheduler
	job.Start()
	// Log Errors
	log.Fatal(app.Listen(":" + os.Getenv("VNET_PORT")))
}