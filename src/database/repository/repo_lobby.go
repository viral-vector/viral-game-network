package repository

import (
	"os"
	"time"
	"viral-game-network/src/database"
	"viral-game-network/src/database/type"
	"github.com/surrealdb/surrealdb.go"
)

func AllLobby() ([]dbtype.Lobby, error) {
	result, err := database.DBS.Query(`
	SELECT * 
	FROM type::table($tb) 
	WHERE COUNT(users) < $mx 
		-- AND (private != true) 
	ORDER BY date_created DESC;`, 
	map[string]string{
		"tb": "Lobby",
		"mx": os.Getenv("LOBBY_MAX_PLAYERS"),
	});
	if err != nil {
		return nil, err
	}

    var lobbies []dbtype.Lobby

    _, err = surrealdb.UnmarshalRaw(result, &lobbies)
    if err != nil {
        return nil, err
    }

	return lobbies, nil
}

func GetLobby(id string) (*dbtype.Lobby, error) {
	// Get user by ID
	data, err := database.DBS.Select(id)
	if err != nil {
		return nil, err
	}
  
	// Unmarshal data
	lobby := new(dbtype.Lobby)
	err = surrealdb.Unmarshal(data, &lobby)
	if err != nil {
		return nil, err
	}

	return lobby, nil
}

func SetLobby(id string, body *dbtype.Lobby) (*dbtype.Lobby, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Updated = now
	data, err = database.DBS.Update(id, body)

	if err != nil {
		return nil, err
	}

	lobby := new(dbtype.Lobby)
	err = surrealdb.Unmarshal(data, &lobby)
	if err != nil {
		return nil, err
	}
		
	return lobby, nil
}

func PutLobby(body *dbtype.Lobby) (*dbtype.Lobby, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Created = now
	body.Date_Updated = now
	data, err = database.DBS.Create("Lobby", body)

	if err != nil {
		return nil, err
	}

	// Unmarshal data
	lobby := make([]*dbtype.Lobby, 1)
	err = surrealdb.Unmarshal(data, &lobby)
	if err != nil {
		return nil, err
	}
	return lobby[0], nil
}