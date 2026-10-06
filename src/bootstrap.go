package main

import (
	"fmt"
	"log"
	"strings"
	"viral-game-network/src/auth"
	"viral-game-network/src/database/repository"
)

func bootstrap(configs map[string]string) error {
	log.Println("Bootstrapping application.")
	_, count, err := repository.AllAdmin(1, 1, "")
	if err != nil {
		return fmt.Errorf("bootstrap: read admins: %w", err)
	}
	keys, err := repository.GetApiKeys()
	if err != nil {
		return fmt.Errorf("bootstrap: read API keys: %w", err)
	}
	if count == 0 && (strings.TrimSpace(configs["adminUsername"]) == "" || configs["adminPassword"] == "") {
		return fmt.Errorf("bootstrap: admin username and password are required")
	}
	if len(keys) == 0 && (strings.TrimSpace(configs["appName"]) == "" || strings.TrimSpace(configs["appKey"]) == "") {
		return fmt.Errorf("bootstrap: app name and API key are required")
	}
	if count == 0 {
		password, err := auth.HashGenerate(configs["adminPassword"])
		if err != nil {
			return fmt.Errorf("bootstrap: hash admin password: %w", err)
		}
		if _, err := repository.GenAdmin(configs["adminUsername"], password); err != nil {
			return fmt.Errorf("bootstrap: create admin: %w", err)
		}
	}
	if len(keys) == 0 {
		if err := repository.AddApiKey(configs["appName"], configs["appKey"]); err != nil {
			return fmt.Errorf("bootstrap: create API key: %w", err)
		}
	}
	return nil
}
