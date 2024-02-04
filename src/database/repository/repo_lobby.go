package repository

import (
	"log"
	"viral-game-network/src/database"
	"viral-game-network/src/database/type"
)



func AllLobby() {
}

func GetLobby() {
}

func SetLobby() {
	// Define Lobby struct
	record := dbtype.Lobby{
		Name:    "viral-test",
	}

	// Insert Lobby
	_, err := database.DBS.Create("Lobby", record)
	if err != nil {
		panic(err)
	}

	log.Println("Database Execution!: " + record.Name)
}