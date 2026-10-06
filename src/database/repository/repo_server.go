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
	filter := ""

	// Search
	if search != "" {
		filter = " WHERE [name, guid, address, port] ?~ $search"
		lQuery += filter
		params["search"] = fmt.Sprintf("%s", search)
	}

	// Ordering
	lQuery = fmt.Sprintf("%s ORDER BY date_created DESC, id ASC", lQuery)

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
		fmt.Sprintf("SELECT count() AS total FROM type::table(Server)%s GROUP ALL;", filter),
		params,
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

// SyncServerPodState keeps heartbeat readiness when the pod is still running.
// Updating only pod fields also prevents stale snapshots from restoring metadata.
func SyncServerPodState(id, phase, address string, port int32) (*dbtype.Server, error) {
	rows, err := database.Query[dbtype.Server](`UPDATE type::record($id) SET
  status=IF $phase='Running' AND status='Online' { 'Online' } ELSE { $phase },
  address=$address, port=$port,
  date_updated=IF $phase='Running' AND status='Online' { date_updated } ELSE { $now }
  RETURN AFTER;`, map[string]interface{}{"id": id, "phase": phase, "address": address, "port": port, "now": time.Now().UTC().Format(time.RFC3339)})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("server not found")
	}
	return &rows[0], nil
}

func TickServer(id string) (*dbtype.Server, error) {
	rows, err := database.Query[dbtype.Server](`UPDATE type::record($id) SET status='Online', date_updated=$now RETURN AFTER;`, map[string]interface{}{"id": id, "now": time.Now().UTC().Format(time.RFC3339)})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("server not found")
	}
	return &rows[0], nil
}
