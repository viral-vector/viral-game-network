package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllApplication(count int, pager int, search string) ([]dbtype.Application, int, error) {
	// Get All Applications
	lQuery := `
	SELECT * 
	FROM type::table(Application) 
	`
	params := map[string]interface{}{}

	// Search
	if search != "" {
		lQuery = fmt.Sprintf("%s WHERE [name, guid, group, image] ?~ $search", lQuery)
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

	// Get Applications.
	apps, err := database.Query[dbtype.Application](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all Applications.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table(Application) GROUP ALL;",
		map[string]interface{}{
			
		},
	)
	if err != nil || len(total) == 0 {
		return apps, 0, err
	}

	return apps, total[0].Total, nil
}

func GetApplication(id string) (*dbtype.Application, error) {
	// Get app by name or guid.
	apps, err := database.Query[dbtype.Application](
		`SELECT * FROM type::record($id);`,
		map[string]interface{}{
			"id": id,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(apps) == 0 {
		return nil, fmt.Errorf("Application not found")
	}

	return &apps[0], nil
}

func SelApplication(info *dbtype.Application) (*dbtype.Application, error) {
	// Get app by name or guid.
	apps, err := database.Query[dbtype.Application](
		`
		SELECT * 
		FROM type::table(Application) 
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

	if len(apps) == 0 {
		return nil, fmt.Errorf("Application not found")
	}

	return &apps[0], nil
}

func SetApplication(id string, app *dbtype.Application) (*dbtype.Application, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	app.Date_Updated = now
	return database.Update[dbtype.Application](*app.ID, app)
}

func PutApplication(app *dbtype.Application) (*dbtype.Application, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	app.Date_Created = now
	return database.Create[dbtype.Application](app)
}
