package repository

import (
	"os"
	// "log"
	"time"
	"viral-game-network/src/database"
	"viral-game-network/src/database/type"
	"github.com/surrealdb/surrealdb.go"

)


func AllLobby() []dbtype.Lobby {
	result, err := database.DBS.Query("SELECT * FROM type::table($tb) WHERE COUNT(users) < $mx AND (private != true);", 
	map[string]string{
		"tb": "Lobby",
		"mx": os.Getenv("LOBBY_MAX_PLAYERS"),
	});
	if err != nil {
		panic(err)
	}

    var lobbies []dbtype.Lobby

    _, err = surrealdb.UnmarshalRaw(result, &lobbies)
    if err != nil {
        panic(err)
    }

	return lobbies
}

func GetLobby(id string) *dbtype.Lobby {
	// Get user by ID
	data, err := database.DBS.Select(id)
	if err != nil {
		panic(err)
	}
  
	// Unmarshal data
	lobby := new(dbtype.Lobby)
	err = surrealdb.Unmarshal(data, &lobby)
	if err != nil {
		panic(err)
	}

	return lobby
}

func SetLobby(id string, body *dbtype.Lobby) *dbtype.Lobby {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Updated = now
	if id == "" {
		body.Date_Created = now
		data, err = database.DBS.Create(id, body)
	} else {
		data, err = database.DBS.Update(id, body)
	}

	if err != nil {
		panic(err)
	}

	lobby := new(dbtype.Lobby)
	err = surrealdb.Unmarshal(data, &lobby)
	if err != nil {
		panic(err)
	}
		
	return lobby
}