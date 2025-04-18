package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllSystemEvents(count int, pager int) ([]dbtype.SystemEvent, int, error) {
	// Get All System Events
	sevents, err := database.Query[dbtype.SystemEvent](`
		SELECT * 
		FROM type::table(System_Event) 
		ORDER BY date_created DESC 
		LIMIT $ct START $pg;`,
		map[string]interface{}{
			"ct": count,
			"pg": (pager - 1) * count,
		})
	if err != nil {
		return nil, 0, err
	}

	// Count All System Events
	total, err := database.Query[dbtype.Total]("SELECT count() AS total FROM type::table(System_Event) GROUP ALL;",
		map[string]interface{}{
		})
	if err != nil {
		return sevents, 0, err
	}
	return sevents, total[0].Total, nil
}

func SelSystemEvents(count int64) ([]dbtype.SystemEvent, error) {
	lQuery := `
	SELECT *
	FROM type::table(System_Event) 
	ORDER BY date_created DESC
	LIMIT $ct
	`
	params := map[string]interface{}{
		"ct": count,
	}

	return database.Query[dbtype.SystemEvent](fmt.Sprintf("%s%s", lQuery, `;`), params)
}

func SelSystemEventsInFrame(seconds int32) ([]dbtype.SystemEvent, error) {
	startTime := time.Now().Add(time.Duration(-seconds) * time.Second).Format(time.RFC3339)

	lQuery := `
	SELECT *
	FROM type::table(System_Event) 
	WHERE date_created >= $dd
	ORDER BY date_created DESC
	`
	params := map[string]interface{}{
		"dd": startTime,
	}

	return database.Query[dbtype.SystemEvent](fmt.Sprintf("%s%s", lQuery, `;`), params)
}

func PutSystemEvent(event *dbtype.SystemEvent) (*dbtype.SystemEvent, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	event.Date_Created = now
	return database.Create[dbtype.SystemEvent](event)
}

func PopSystemevent(event *dbtype.SystemEvent) (*dbtype.SystemEvent, error) {
	return database.Upsert(event)
}
