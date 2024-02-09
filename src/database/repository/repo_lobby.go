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
    ,lobby_host.*
    ,->Lobby_Users.* as lobby_users 
    ,array::first(SELECT id, name, guid FROM ->Lobby_Users.out) as lobby_users.user 
	FROM type::table($tb) 
	WHERE COUNT(lobby_users) < $mx
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
	// Get lobby by ID
	data, err := database.DBS.Query(`
	SELECT * 
	,lobby_host.* 
	,->Lobby_Users.* as lobby_users 
    ,array::first(SELECT id, name, guid FROM ->Lobby_Users.out) as lobby_users.user 
	FROM Lobby WHERE id=$id;`, 
	map[string]string{
		"id": id,
	});

	if err != nil {
		return nil, err
	}
  
	lobby := make([]*dbtype.Lobby, 1)

	// Unmarshal data
	_, err = surrealdb.UnmarshalRaw(data, &lobby)
	if err != nil {
		return nil, err
	}

	return lobby[0], nil
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

func PutLobby(body *dbtype.Lobby, user *dbtype.User) (*dbtype.Lobby, error) {
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

	// Link Host
	_, err = database.DBS.Query(`UPDATE $id MERGE {
		lobby_host:$lobby_host
	}`, 
	map[string]interface{}{
		"id": lobby[0].ID,
		"lobby_host": user.ID,
	})
	
	if err != nil {
		return nil, err
	}
	
	lobby[0].Lobby_Host = user

	return lobby[0], nil
}

func LinkLobbyUser(id string, user *dbtype.User) (error) {
	_, err := database.DBS.Query(`DELETE FROM Lobby_Users WHERE out=$user;`, 
	map[string]string{
		"user": user.ID,
	});
	if err != nil {
		return err
	}

	_, err = database.DBS.Query(`RELATE $lobby->Lobby_Users->$user 
		CONTENT {
			date_created: $date_created
		};`, 
	map[string]string{
		"lobby": id,
		"user": user.ID,
		"date_created": time.Now().UTC().Format(time.RFC3339),
	});
	if err != nil {
		return err
	}

	return nil
}