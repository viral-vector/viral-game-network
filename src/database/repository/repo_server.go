package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllServer() ([]dbtype.Server, error) {
	servers, err := database.Query[dbtype.Server]("SELECT * FROM type::table($tb);",
		map[string]interface{}{
			"tb": "Server",
		})
	
		return servers, err
}

func GetServer(id string) (*dbtype.Server, error) {
	server, err := database.Select[dbtype.Server](id)

	return server, err
}

func SetServer(id string, body *dbtype.Server) (*dbtype.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	body.Date_Updated = now
	server, err := database.Update[dbtype.Server](body)
	
	return server, err
}

func PutServer(body *dbtype.Server, lobby *dbtype.Lobby) (*dbtype.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	body.Date_Created = now
	body.Date_Updated = now
	server, err := database.Create[dbtype.Server](body)

	return server, err
}

func DelServer(id string) error {
	_, err := GetServer(id)
	if err != nil {
		return fmt.Errorf("DelServer Error: %v", err)
	} 
	return nil
}

func GetServerPorts() []int32 {
	servers, err := database.Query[dbtype.Server]("SELECT id, port FROM Server;",
		map[string]interface{}{})
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
	now := time.Now().UTC().Format(time.RFC3339)
	server, err := GetServer(id)
	if err != nil || server == nil {
		return nil, err
	}

	server.Date_Updated = now
	server.Status = "Online"
	updated, err := database.Update[dbtype.Server](server)
	if err != nil {
		return nil, err
	}

	return updated, nil
}