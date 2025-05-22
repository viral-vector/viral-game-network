package main

import (
	"log"
	"os"
	"strings"
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
	engine.AddFunc("strip", func(haystack string, needle string) string {
		return strings.ReplaceAll(haystack, needle, "")
	})
	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		Prefork:           os.Getenv("development") == "production",
		CaseSensitive:     true,
		StrictRouting:     false,
		ServerHeader:      "VGN",
		AppName:           os.Getenv("VNET_NAME"),
		Views:             engine,
		ViewsLayout:       "base",
		PassLocalsToViews: true,
	})

	// bootstrap the application
	go bootstrap(map[string]string{
		"appKey": os.Getenv("VNET_KEY"),
		"appName": os.Getenv("VNET_NAME"),
		"adminUsername": os.Getenv("APP_ADMIN_USERNAME"),
		"adminPassword": os.Getenv("APP_ADMIN_PASSWORD"),
	})
	// Serve static files from the public folder
	app.Static("/", "./public")
	// Register routes 
	serve_routes(app)
	// Start the job scheduler
	job.Start()

	// Listen & Log
	log.Fatal(app.Listen(":" + os.Getenv("VNET_PORT")))
	// Listen & Log TLS
	// log.Fatal(app.Listen(":" + os.Getenv("VNET_PORT"),
		// "/app/bin/ssl/tls.crt",
		// "/app/bin/ssl/tls.key"))
}