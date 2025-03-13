package migrations

import (
	dbtype "viral-game-network/src/database/type"
)

func init() {
	// Append the migration to the AllMigrations slice.
	AllMigrations = append(AllMigrations, dbtype.Migration{
		Name:  "20250312-192559",
		SQL: "DEFINE TABLE Hello SCHEMAFULL;",
	})
}
