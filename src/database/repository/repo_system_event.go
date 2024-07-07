package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"

	"github.com/surrealdb/surrealdb.go"
)

func GetSystemEvents(count int32) ([]dbtype.SystemEvent, error) {
	var err error

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

	// Get All Events
	result, err := database.DBS.Query(fmt.Sprintf("%s%s", lQuery, `;`), params)
	if err != nil {
		return nil, err
	}

	var events []dbtype.SystemEvent
	_, err = surrealdb.UnmarshalRaw(result, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func GetSystemEventsInFrame(seconds int32) ([]dbtype.SystemEvent, error) {
	var err error
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

	// Get All Events
	result, err := database.DBS.Query(fmt.Sprintf("%s%s", lQuery, `;`), params)
	if err != nil {
		return nil, err
	}

	var events []dbtype.SystemEvent
	_, err = surrealdb.UnmarshalRaw(result, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func PutSystemEvent(body dbtype.SystemEvent) (*dbtype.SystemEvent, error) {
	var err error
	var data interface{}
	var now = time.Now().UTC().Format(time.RFC3339)

	body.Date_Created = now
	data, err = database.DBS.Create("SystemEvent", body)

	if err != nil {
		return nil, err
	}

	// Unmarshal data
	event := make([]*dbtype.SystemEvent, 1)
	err = surrealdb.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return event[0], nil
}
