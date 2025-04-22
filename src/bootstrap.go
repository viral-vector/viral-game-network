package main

import (
	"log"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
)

func bootstrap() {
	log.Println("Bootstrapping application.")

	// Check for default admin user
	_, count, _ := repository.AllAdmin(1, 1, "")
	if count == 0 {
		log.Println("Bootstrapping: No admin user found, creating default admin user.")
		password, err := auth.HashGenerate("password")
		if err != nil {
			log.Fatalf("Bootstrapping: Failed to create admin user: %s", err)
		}
		_, err = repository.GenAdmin(password)
		if err != nil {
			log.Fatalf("Bootstrapping: Failed to create admin user: %s", err)
		}
	}
}