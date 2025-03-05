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
	FROM type::table($tb) 
	ORDER BY date_created DESC
	LIMIT $ct
	`
	params := map[string]interface{}{
		"tb": "SystemEvent",
		"ct": count,
	}

	events, err := database.Query[dbtype.SystemEvent](fmt.Sprintf("%s%s", lQuery, `;`), params)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func GetSystemEventsInFrame(seconds int32) ([]dbtype.SystemEvent, error) {
	startTime := time.Now().Add(time.Duration(-seconds) * time.Second).Format(time.RFC3339)

	lQuery := `
	SELECT *
	FROM type::table($tb) 
	WHERE date_created >= $dd
	ORDER BY date_created DESC
	`
	params := map[string]interface{}{
		"tb": "SystemEvent",
		"dd": startTime,
	}

	events, err := database.Query[dbtype.SystemEvent](fmt.Sprintf("%s%s", lQuery, `;`), params)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func PutSystemEvent(body *dbtype.SystemEvent) (*dbtype.SystemEvent, error) {
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Created = now
	event, err := database.Create[dbtype.SystemEvent](body)

	if err != nil {
		return nil, err
	}
	return event, nil
}
