package main

import (
	"log"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
)

func bootstrap(configs map[string]string) {
	log.Println("Bootstrapping application.")

	_, count, _ := repository.AllAdmin(1, 1, "")
	if count == 0 {
		log.Println("Bootstrapping: No admin user found, creating default admin user.")
		password, err := auth.HashGenerate(configs["adminPassword"])
		if err != nil {
			log.Fatalf("Bootstrapping: Failed to create admin user: %s", err)
		}
		_, err = repository.GenAdmin(configs["adminUsername"], password)
		if err != nil {
			log.Fatalf("Bootstrapping: Failed to create admin user: %s", err)
		}
	}

	keys, _ := repository.GetApiKeys()
	if len(keys) == 0 {
		log.Println("Bootstrapping: No api keys, creating key.")
		err := repository.AddApiKey(configs["appName"], configs["appKey"])
		if err != nil {
			log.Fatalf("Bootstrapping: Failed to create api key: %s", err)
		}
	}
}