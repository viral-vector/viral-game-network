package sqlquery

import (
	"context"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/surrealdb/surrealdb.go"
	"time"
)

// Query checks every statement status, including failures inside transactions.
func Query[T any](db *surrealdb.DB, query string, params map[string]interface{}) ([]T, error) {
	if db == nil {
		return nil, fmt.Errorf("database is not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results, err := surrealdb.Query[cbor.RawMessage](ctx, db, query, params)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return nil, fmt.Errorf("database returned no query response")
	}
	var output []T
	for _, result := range *results {
		if result.Status != "OK" {
			var message string
			if err := cbor.Unmarshal(result.Result, &message); err != nil {
				message = result.Status
			}
			return nil, fmt.Errorf("database statement failed: %s", message)
		}
		var rows []T
		if err := cbor.Unmarshal(result.Result, &rows); err == nil {
			output = append(output, rows...)
			continue
		}
		var row T
		if err := cbor.Unmarshal(result.Result, &row); err != nil {
			return nil, fmt.Errorf("decode query result: %w", err)
		}
		output = append(output, row)
	}
	return output, nil
}
