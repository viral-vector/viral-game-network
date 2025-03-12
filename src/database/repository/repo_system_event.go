package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func GetSystemEvents(count int32) ([]dbtype.SystemEvent, error) {
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

func GetSystemEventsInFrame(seconds int32) ([]dbtype.SystemEvent, error) {
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
