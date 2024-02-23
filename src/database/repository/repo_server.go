package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	"viral-game-network/src/database/type"
	"github.com/surrealdb/surrealdb.go"
)

func AllServer() ([]dbtype.Server, error) {
	result, err := database.DBS.Query("SELECT * FROM type::table($tb);", 
	map[string]string{
		"tb": "Server",
	});
	if err != nil {
		return nil, err
	}

    var servers []dbtype.Server

    _, err = surrealdb.UnmarshalRaw(result, &servers)
    if err != nil {
        return nil, err
    }

	return servers, nil
}

func GetServer(info dbtype.Server) (*dbtype.Server, error) {
	// Get server by ID
	data, err := database.DBS.Query(`
		SELECT * 
		FROM type::table($tb) 
		WHERE 
			(ID = $id) 
		LIMIT 1;`, 
	map[string]string{
		"tb": "Server",
		"id": info.ID,
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal data
	server := make([]*dbtype.Server, 1)
	_, err = surrealdb.UnmarshalRaw(data, &server)
	if err != nil {
		return nil, err
	}

	if server[0] == nil {
		return nil, fmt.Errorf("Server not found")
	}

	return server[0], err
}

func SetServer(id string, body *dbtype.Server) (*dbtype.Server, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Updated = now
	data, err = database.DBS.Update(id, body)

	if err != nil {
		return nil, err
	}

	server := new(dbtype.Server)
	err = surrealdb.Unmarshal(data, &server)
	if err != nil {
		return nil, err
	}
		
	return server, nil
}

func PutServer(body dbtype.Server, lobby *dbtype.Lobby) (*dbtype.Server, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Created = now
	body.Date_Updated = now
	data, err = database.DBS.Create("Server", body)

	if err != nil {
		return nil, err
	}

	// Unmarshal data
	server := make([]*dbtype.Server, 1)
	err = surrealdb.Unmarshal(data, &server)
	if err != nil {
		return nil, err
	}

	// Link Lobby
	_, err = database.DBS.Query(`UPDATE $id MERGE {
		lobby:$lobby
	}`, 
	map[string]interface{}{
		"id": server[0].ID,
		"lobby": lobby.ID,
	})
	
	if err != nil {
		return nil, err
	}
	
	server[0].Lobby = lobby

	return server[0], nil
}

func DelServer(id string) error {
	_, err := database.DBS.Delete(id)
	if err != nil {
		return err
	}
	return nil
}