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
	Commands()
}

func ServeApp() {
	// Create a new engine
	engine := pug.New("./views", ".pug")
	engine.Reload(true)

	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		Prefork:           os.Getenv("APP_PREFORK") == "true",
		CaseSensitive:     true,
		StrictRouting:     false,
		ServerHeader:      "VGN",
		AppName:           os.Getenv("APP_NAME"),
		Views:             engine,
		ViewsLayout:       "base",
		PassLocalsToViews: true,
	})
	// Register routes
	serve_routes(app)
	// Serve static files from the public folder
	app.Static("/", "./public")
	// Start the job scheduler
	job.Start()
	// Log Errors
	log.Fatal(app.Listen(":3000"))
}

func Commands() {

}
