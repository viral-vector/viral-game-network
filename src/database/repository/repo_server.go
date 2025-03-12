package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllServer() ([]dbtype.Server, error) {
	return database.Query[dbtype.Server]("SELECT * FROM type::table(Server);",
		map[string]interface{}{
		})
}

func GetServerByGuid(guid string) (*dbtype.Server, error) {
	results, err := database.Query[dbtype.Server]("SELECT * FROM type::table(Server) WHERE guid=$guid;",
		map[string]interface{}{
			"guid": guid,
		})
	if err != nil{
		return nil, err
	}
	if(len(results) == 0){
		return nil, nil
	}
	return &results[0], nil
}

func GetServer(id string) (*dbtype.Server, error) {
	return database.Select[dbtype.Server](id)
}

func SetServer(id string, server *dbtype.Server) (*dbtype.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	server.Date_Updated = now
	return database.Update[dbtype.Server](*server.ID, server)
}

func PutServer(server *dbtype.Server, lobby *dbtype.Lobby) (*dbtype.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	server.Date_Created = now
	return database.Create[dbtype.Server](server)
}

func DelServer(id string) error {
	server, err := GetServer(id)
	if err != nil {
		return fmt.Errorf("DelServer Error: %v", err)
	} 
	err = database.Delete(*server.ID)
	if err != nil {
		return fmt.Errorf("DelServer error: %s", err)
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
	server, err := GetServer(id)
	if err != nil {
		return nil, fmt.Errorf("TickServer Error: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	server.Date_Updated = now
	server.Status = "Online"
	return database.Update[dbtype.Server](*server.ID, server)
}