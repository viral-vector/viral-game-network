package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllUser(count int, pager int) ([]dbtype.User, int, error) {
	// Get All Users
	users, err := database.Query[dbtype.User](`
	SELECT * 
	FROM type::table(User) 
	ORDER BY date_created DESC 
	LIMIT $ct START $pg;`,
		map[string]interface{}{
			"ct": count,
			"pg": (pager - 1) * count,
		})
	if err != nil {
		return nil, 0, err
	}

	// Count All Users
	total, err := database.Query[dbtype.Total]("SELECT count() AS total FROM type::table(User) GROUP ALL;",
		map[string]interface{}{
		})
	if err != nil {
		return users, 0, err
	}
	return users, total[0].Total, nil
}

func GetUser(info *dbtype.User) (*dbtype.User, error) {
	// Get user by name or guid.
	users, err := database.Query[dbtype.User](
		`
		SELECT * 
		FROM type::table(User) 
		WHERE 
			(name = $name OR guid = $guid) 
		LIMIT 1;`,
		map[string]interface{}{
			"name": info.Name,
			"guid": info.Guid,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("User not found")
	}

	return &users[0], nil
}

func SetUser(id string, user *dbtype.User) (*dbtype.User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	user.Date_Updated = now
	return database.Update[dbtype.User](*user.ID, user)
}

func PutUser(user *dbtype.User) (*dbtype.User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	user.Date_Created = now
	user.Guid = database.GetUUID()
	return database.Create[dbtype.User](user)
}
