package repository

import (
	"os"
	"fmt"
	"time"
	"viral-game-network/src/database"
	"viral-game-network/src/database/type"
	"github.com/surrealdb/surrealdb.go"
)

func AllUser() ([]dbtype.User, error) {
	result, err := database.DBS.Query("SELECT * FROM type::table($tb) WHERE COUNT(users) < $mx AND (private != true);", 
	map[string]string{
		"tb": "User",
		"mx": os.Getenv("LOBBY_MAX_PLAYERS"),
	});
	if err != nil {
		return nil, err
	}

    var users []dbtype.User

    _, err = surrealdb.UnmarshalRaw(result, &users)
    if err != nil {
        return nil, err
    }

	return users, nil
}

func GetUser(info dbtype.User) (*dbtype.User, error) {
	// Get user by ID
	data, err := database.DBS.Query(`
		SELECT * 
		FROM type::table($tb) 
		WHERE 
			(name = $name OR guid = $guid) 
		LIMIT 1;`, 
	map[string]string{
		"tb": "User",
		"name": info.Name,
		"guid": info.Guid,
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal data
	user := make([]*dbtype.User, 1)
	_, err = surrealdb.UnmarshalRaw(data, &user)
	if err != nil {
		return nil, err
	}

	if user[0] == nil {
		return nil, fmt.Errorf("User not found")
	}

	return user[0], err
}

func SetUser(id string, body *dbtype.User) (*dbtype.User, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Updated = now
	data, err = database.DBS.Update(id, body)

	if err != nil {
		return nil, err
	}

	user := new(dbtype.User)
	err = surrealdb.Unmarshal(data, &user)
	if err != nil {
		return nil, err
	}
		
	return user, nil
}

func PutUser(body dbtype.User) (*dbtype.User, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Created = now
	body.Date_Updated = now
	data, err = database.DBS.Create("User", body)

	if err != nil {
		return nil, err
	}

	// Unmarshal data
	user := make([]*dbtype.User, 1)
	err = surrealdb.Unmarshal(data, &user)
	if err != nil {
		return nil, err
	}
	return user[0], nil
}