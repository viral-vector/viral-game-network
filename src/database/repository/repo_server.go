package repository

import (
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"

	"github.com/surrealdb/surrealdb.go"
)

func AllServer() ([]dbtype.Server, error) {
	result, err := database.DBS.Query("SELECT * FROM type::table($tb);",
		map[string]string{
			"tb": "Server",
		})
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

func GetServer(id string) (*dbtype.Server, error) {
	// Get server by ID
	data, err := database.DBS.Query(`
	SELECT * 
	FROM Server WHERE id=$id;`,
		map[string]string{
			"id": id,
		})

	if err != nil {
		return nil, err
	}

	server := make([]*dbtype.Server, 1)

	// Unmarshal data
	_, err = surrealdb.UnmarshalRaw(data, &server)
	if err != nil {
		return nil, err
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

	return server[0], nil
}

func DelServer(id string) error {
	_, err := database.DBS.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func GetServerPorts() []int32 {
	result, err := database.DBS.Query("SELECT id, port FROM type::table($tb);",
		map[string]string{
			"tb": "Server",
		})
	if err != nil {
		return nil
	}

	var servers []dbtype.Server

	_, err = surrealdb.UnmarshalRaw(result, &servers)
	if err != nil {
		return nil
	}

	ports := make([]int32, 0)
	for _, server := range servers {
		if server.Port != 0 {
			ports = append(ports, server.Port)
		}
	}
	return ports
}

func TickServer(id string) (*dbtype.Server, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body, err := GetServer(id)
	if err != nil || body == nil {
		return nil, err
	}

	body.Date_Updated = now
	body.Status = "Online"
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
