package repository

import (
	"crypto/sha256"
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
		map[string]interface{}{})
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
	return SelSystemEventsSince(time.Now().Add(-time.Duration(seconds) * time.Second))
}

func SelSystemEventsSince(since time.Time) ([]dbtype.SystemEvent, error) {
	startTime := since.UTC().Format(time.RFC3339)

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
	if event == nil || event.Ref_ID == "" || event.Ref_Source == "" {
		return nil, fmt.Errorf("system event requires a source and reference ID")
	}
	// A stable record ID makes polling and retries idempotent across workers.
	key := sha256.Sum256([]byte(event.Ref_Source + "\x00" + event.Ref_ID))
	rows, err := database.Query[dbtype.SystemEvent](`
		UPSERT type::record($id) SET
			date_created = IF date_created != NONE AND message = $message
				AND severity = $severity AND ref_target = $target AND ref_source = $source
				{ date_created } ELSE { $created },
			ref_id = $ref, ref_source = $source, ref_target = $target,
			message = $message, severity = $severity
		RETURN AFTER;
	`, map[string]interface{}{
		"id": fmt.Sprintf("System_Event:%x", key), "ref": event.Ref_ID,
		"source": event.Ref_Source, "target": event.Ref_Target,
		"message": event.Message, "severity": event.Severity,
		"created": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, fmt.Errorf("database returned no system event")
	}
	return &rows[0], nil
}
