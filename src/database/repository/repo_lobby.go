package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllLobby(count int, pager int) ([]dbtype.Lobby, int, error) {
	lQuery := `
	SELECT *
	,array::first(SELECT * FROM ->Lobby_Application.out) AS lobby_application
    ,(IF count(SELECT id FROM ->Lobby_Host.out) > 0
       {array::first(SELECT id, name, guid FROM ->Lobby_Host.out)} ELSE {NULL}) AS lobby_host
    ,(IF count(SELECT id FROM ->Lobby_Server.out) > 0
       {array::first(SELECT id, name, guid FROM ->Lobby_Server.out)} ELSE {NULL}) AS lobby_server
    ,(SELECT * FROM ->Lobby_Users.out) AS lobby_users
	FROM type::table(Lobby)
	ORDER BY date_created DESC
	`
	params := map[string]interface{}{
		
	}

	if count > -1 {
		lQuery = fmt.Sprintf("%s LIMIT $ct START $pg", lQuery)
		params["ct"] = count
		params["pg"] = (pager - 1) * count
	}

	// Get all lobbies.
	lobbies, err := database.Query[dbtype.Lobby](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		fmt.Println("AllLobby: ", err)
		return nil, 0, err
	}

	// Count all lobbies.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table(Lobby) GROUP ALL;",
		map[string]interface{}{
			
		},
	)
	if err != nil || len(total) == 0 {
		return lobbies, 0, err
	}

	return lobbies, total[0].Total, nil
}

func GetLobby(id string) (*dbtype.Lobby, error) {
	// Get lobby by ID.
	lobbies, err := database.Query[dbtype.Lobby](`
	SELECT * 
	,array::first(SELECT * FROM ->Lobby_Application.out) AS lobby_application
    ,(IF count(SELECT id FROM ->Lobby_Host.out) > 0
       {array::first(SELECT id, name, guid FROM ->Lobby_Host.out)} ELSE {NULL}) AS lobby_host
    ,(IF count(SELECT id FROM ->Lobby_Server.out) > 0
       {array::first(SELECT * FROM ->Lobby_Server.out)} ELSE {NULL}) AS lobby_server
    ,(SELECT * FROM ->Lobby_Users.out) AS lobby_users
	FROM type::record($id);`,
		map[string]interface{}{
			"id": id,
		})
	if err != nil {
		return nil, err
	}
	if len(lobbies) == 0 {
		return nil, nil
	}
	return &lobbies[0], nil
}

func SetLobby(id string, lobby *dbtype.Lobby) (*dbtype.Lobby, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	lobby.Date_Updated = now
	return database.Update[dbtype.Lobby](*lobby.ID, lobby)
}

func PutLobby(body *dbtype.Lobby, app *dbtype.Application, user *dbtype.User) (*dbtype.Lobby, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	body.Date_Created = now

	// Create the lobby.
	lobby, err := database.Create[dbtype.Lobby](body)
	if err != nil {
		return nil, err
	}
	// Link the app.
	err = LinkLobbyApplication(lobby, app)
	if err != nil {
		return nil, err
	}
	// Link the host.
	err = LinkLobbyHost(lobby, user)
	if err != nil {
		return nil, err
	}
	
	return GetLobby(lobby.ID.String())
}

func DelLobby(id string) error {
	lobby, err := GetLobby(id)
	if err != nil || lobby == nil {
		return fmt.Errorf("DelLobby error: lobby not found %s", err)
	}
	err = database.Delete(*lobby.ID)
	if err != nil {
		return fmt.Errorf("DelLobby error: lobby not found %s", err)
	}
	return nil
}

func LinkLobbyApplication(lobby *dbtype.Lobby, app *dbtype.Application) error {
	return database.Relate(lobby.ID, app.ID, "Lobby_Application", map[string]interface{}{

	})
}

func LinkLobbyHost(lobby *dbtype.Lobby, user *dbtype.User) error {
	return database.Relate(lobby.ID, user.ID, "Lobby_Host", map[string]interface{}{

	})
}

func LinkLobbyUser(lobby *dbtype.Lobby, user *dbtype.User) error {
	// First, remove any existing relation.
	_, err := database.Query[any](
		`DELETE FROM Lobby_Users WHERE out=$user RETURN *;`,
		map[string]interface{}{
			"user": user.ID.String(),
		},
	)
	if err != nil {
		return err
	}

	return database.Relate(lobby.ID, user.ID, "Lobby_Users", map[string]interface{}{
		"date_created": time.Now().UTC().Format(time.RFC3339),
	})
}

func UnlinkLobbyAllUsers(id string) error {
	_, err := database.Query[any](
		`DELETE FROM Lobby_Users WHERE in=$lobby RETURN *;`,
		map[string]interface{}{
			"lobby": id,
		},
	)
	return err
}

func LinkLobbyServer(lobby *dbtype.Lobby, server *dbtype.Server) error {
	return database.Relate(lobby.ID, server.ID, "Lobby_Server", map[string]interface{}{
		"date_created": time.Now().UTC().Format(time.RFC3339),
	})
}