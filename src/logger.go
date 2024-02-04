package main

import (
	"os"
	"log"
)

func init() {
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("./tmp/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.SetOutput(os.Stdout)

	defer file.Close()
}