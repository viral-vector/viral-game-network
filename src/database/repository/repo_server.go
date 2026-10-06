package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllServer(count int, pager int, search string) ([]dbtype.Server, int, error) {
	lQuery := `
	SELECT *
	,(IF count(SELECT id FROM <-Lobby_Server.in) > 0
		{array::first(SELECT * FROM <-Lobby_Server.in)} ELSE {NULL}) AS lobby 
	FROM type::table(Server)`

	params := map[string]interface{}{}

	// Search
	if search != "" {
		lQuery = fmt.Sprintf("%s WHERE [name, guid, address, port] ?~ $search", lQuery)
		params["search"] = fmt.Sprintf("%s", search)
	}

	// Ordering
	lQuery = fmt.Sprintf("%s ORDER BY date_created DESC", lQuery)

	// Limit and Pagination
	if count > -1 {
		lQuery = fmt.Sprintf("%s LIMIT $ct START $pg", lQuery)
		params["ct"] = count
		params["pg"] = (pager - 1) * count
	}

	servers, err := database.Query[dbtype.Server](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all servers.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table(Server) GROUP ALL;",
		map[string]interface{}{},
	)
	if err != nil || len(total) == 0 {
		return servers, 0, err
	}

	return servers, total[0].Total, nil

}

func GetServerByGuid(guid string) (*dbtype.Server, error) {
	results, err := database.Query[dbtype.Server]("SELECT * FROM type::table(Server) WHERE guid=$guid;",
		map[string]interface{}{
			"guid": guid,
		})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

func GetServer(id string) (*dbtype.Server, error) {
	// Get lobby by ID.
	servers, err := database.Query[dbtype.Server](`
	SELECT *
	,array::first(SELECT * FROM <-Lobby_Server.in) AS lobby 
	FROM type::record($id)
	ORDER BY date_created DESC;`,
		map[string]interface{}{
			"id": id,
		})
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return nil, nil
	}
	return &servers[0], nil
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

func DelServer(id string, server *dbtype.Server) error {
	return database.Delete(*server.ID)
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

	if server == nil || server.ID == nil {
		return nil, fmt.Errorf("server not found")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	server.Date_Updated = now
	server.Status = "Online"
	return database.Update[dbtype.Server](*server.ID, server)
}
