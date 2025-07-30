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
	cmd := os.Getenv("VNET_JOB_NAME")
	// Start app in env
	log.Println("Starting VGN", cmd)
	if cmd != "" {
		ServeCMD(cmd)
	} else {
		ServeApp()
	}
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
		Prefork:           false,
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
	// Listen & Log
	log.Fatal(app.Listen(":" + os.Getenv("VNET_PORT")))
}

func ServeCMD(name string) {
	// Start the job
	entries := job.Stack()

	var selected *job.Entry
    for i := range entries {
        if entries[i].Name == name {
            selected = &entries[i]
            break
        }
    }
    if selected == nil {
        log.Fatalf("No job with name %q found", name)
    }
    selected.Func()
}