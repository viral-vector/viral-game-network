package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"

	"github.com/surrealdb/surrealdb.go"
)

func AllUser(count int, pager int) ([]dbtype.User, int, error) {
	// Get All Users
	result, err := database.DBS.Query(`
	SELECT * 
	FROM type::table($tb) 
	ORDER BY date_created DESC 
	LIMIT $ct START $pg;`,
		map[string]interface{}{
			"tb": "User",
			"ct": count,
			"pg": (pager - 1) * count,
		})
	if err != nil {
		return nil, 0, err
	}

	var users []dbtype.User
	_, err = surrealdb.UnmarshalRaw(result, &users)
	if err != nil {
		return nil, 0, err
	}

	// Count All Users
	result, err = database.DBS.Query("SELECT count() AS total FROM type::table($tb) GROUP ALL;",
		map[string]interface{}{
			"tb": "User",
		})
	if err != nil {
		return users, 0, err
	}

	var total []struct {
		Total int `json:"total"`
	}
	_, err = surrealdb.UnmarshalRaw(result, &total)
	if err != nil || len(total) == 0 {
		return users, 0, err
	}

	return users, total[0].Total, nil
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
			"tb":   "User",
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
