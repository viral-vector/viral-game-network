package repository

import (
	"fmt"
	"os"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllLobby(count int, pager int) ([]dbtype.Lobby, int, error) {
	lQuery := `
	SELECT *
    ,lobby_host.*
	,lobby_server.*
    ,->Lobby_Users.* as lobby_users 
    ,array::first(SELECT id, name, guid FROM ->Lobby_Users.out) as lobby_users_user 
	FROM type::table($tb) 
	WHERE COUNT(lobby_users) < $mx
	ORDER BY date_created DESC
	`
	params := map[string]interface{}{
		"tb": "Lobby",
		"mx": os.Getenv("LOBBY_MAX_PLAYERS"),
	}

	if count > -1 {
		lQuery = fmt.Sprintf("%s LIMIT $ct START $pg", lQuery)
		params["ct"] = count
		params["pg"] = (pager - 1) * count
	}

	// Get all lobbies.
	lobbies, err := database.Query[dbtype.Lobby](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all lobbies.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table($tb) GROUP ALL;",
		map[string]interface{}{
			"tb": "Lobby",
		},
	)
	if err != nil || len(total) == 0 {
		return lobbies, 0, err
	}

	return lobbies, total[0].Total, nil
}

func GetLobby(id string) (*dbtype.Lobby, error) {
	// Get lobby by ID.
	lobbies, err := database.Query[dbtype.Lobby](
		`
	SELECT * 
	,lobby_host.*
	,lobby_server.* 
	,->Lobby_Users.* as lobby_users 
    ,array::first(SELECT id, name, guid FROM ->Lobby_Users.out) as lobby_users_user 
	FROM Lobby WHERE id=$id;`,
		map[string]interface{}{
			"id": id,
		})
	if err != nil {
		return nil, err
	}
	if len(lobbies) == 0 {
		return nil, fmt.Errorf("lobby not found")
	}
	return &lobbies[0], nil
}

func SetLobby(id string, body *dbtype.Lobby) (*dbtype.Lobby, error) {
	body.Date_Updated = time.Now().UTC().Format(time.RFC3339)
	// Update the lobby record.
	lobby, err := database.Update[dbtype.Lobby](body)
	if err != nil {
		return nil, err
	}
	return lobby, nil
}

func PutLobby(body *dbtype.Lobby, user *dbtype.User) (*dbtype.Lobby, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	body.Date_Created = now
	body.Date_Updated = now

	// Create the lobby.
	lobby, err := database.Create[dbtype.Lobby](body)
	if err != nil {
		return nil, err
	}
	if lobby == nil{
		return nil, fmt.Errorf("failed to create lobby")
	}

	// Link the host.
	err = LinkLobbyHost(lobby.ID.String(), user)
	if err != nil {
		return nil, err
	}

	lobby.Lobby_Host = user
	
	return lobby, nil
}

func DelLobby(id string) error {
	err := database.Delete[dbtype.Lobby](id)
	if err != nil {
		return err
	}
	return UnlinkLobbyAllUsers(id)
}

func LinkLobbyHost(id string, user *dbtype.User) error {
	_, err := database.Query[any](
		`UPDATE $id MERGE {
			lobby_host: $lobby_host
		}`,
		map[string]interface{}{
			"id":         id,
			"lobby_host": user.ID.String(),
		},
	)
	return err
}

func LinkLobbyUser(id string, user *dbtype.User) error {
	// First, remove any existing relation.
	_, err := database.Query[any](
		`DELETE FROM Lobby_Users WHERE out=$user;`,
		map[string]interface{}{
			"user": user.ID.String(),
		},
	)
	if err != nil {
		return err
	}

	// Create the relation between the lobby and the user.
	_, err = database.Query[any](
		`RELATE $lobby->Lobby_Users->$user 
		CONTENT {
			date_created: $date_created
		};`,
		map[string]interface{}{
			"lobby":        id,
			"user":         user.ID,
			"date_created": time.Now().UTC().Format(time.RFC3339),
		},
	)
	return err
}

func UnlinkLobbyAllUsers(id string) error {
	_, err := database.Query[any](
		`DELETE FROM Lobby_Users WHERE in=$lobby;`,
		map[string]interface{}{
			"lobby": id,
		},
	)
	return err
}

func LinkLobbyServer(id string, server *dbtype.Server) error {
	_, err := database.Query[any](
		`UPDATE $id MERGE {
			lobby_server: $lobby_server
		}`,
		map[string]interface{}{
			"id":           id,
			"lobby_server": server.ID,
		},
	)
	return err
}