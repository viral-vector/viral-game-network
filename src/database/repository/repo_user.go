package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllUser(count int, pager int, search string) ([]dbtype.User, int, error) {
	// Get All Users.
	lQuery := `
	SELECT * 
	FROM type::table(User) 
	`
	params := map[string]interface{}{}

	// Search
	if search != "" {
		lQuery = fmt.Sprintf("%s WHERE [name, guid] ?~ $search", lQuery)
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

	// Get User.
	users, err := database.Query[dbtype.User](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all User.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table(User) GROUP ALL;",
		map[string]interface{}{
			
		},
	)
	if err != nil || len(total) == 0 {
		return users, 0, err
	}

	return users, total[0].Total, nil
}

func GetUser(id string) (*dbtype.User, error) {
	// Get user by name or guid.
	users, err := database.Query[dbtype.User](
		`SELECT * FROM type::record($id);`,
		map[string]interface{}{
			"id": id,
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

func SelUser(info *dbtype.User) (*dbtype.User, error) {
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

func DelUser(id string, user *dbtype.User) error {
	return database.Delete(*user.ID)
}