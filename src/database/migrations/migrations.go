package migrations

import (
	"fmt"
	"log"
	"time"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/database/sqlquery"
    "github.com/surrealdb/surrealdb.go"
)

// AllMigrations aggregates all migration steps.
var AllMigrations = []dbtype.Migration{}


func Migrations_Add(m dbtype.Migration) {
	AllMigrations = append(AllMigrations, m)
}

// Run All(pending) Migrations
func Migrations_Run(DBS *surrealdb.DB) error {
	log.Println("Checking Migrations")

	// Ensure that the Migrations table exists
	MigrationsTableScaffold := []string {
		`
		DEFINE TABLE IF NOT EXISTS Migrations SCHEMAFULL;
		DEFINE FIELD IF NOT EXISTS id ON TABLE Migrations TYPE string;
		DEFINE FIELD IF NOT EXISTS name ON TABLE Migrations TYPE string;
		DEFINE FIELD IF NOT EXISTS date_created ON TABLE Migrations TYPE string;
		DEFINE FIELD IF NOT EXISTS sql ON TABLE Migrations TYPE string;
		DEFINE INDEX IF NOT EXISTS idx_migrations_id ON TABLE Migrations COLUMNS id UNIQUE;
		DEFINE INDEX IF NOT EXISTS idx_migrations_name ON TABLE Migrations COLUMNS name UNIQUE;
		`,
	}

	for _, mig := range MigrationsTableScaffold {
		_, err := sqlquery.Query[any](DBS, mig, nil)
		if err != nil {
			return fmt.Errorf("Migrate error: %w", err)
		}
	} 

	// Retrieve applied migrations from the Migrations table
	results, err := sqlquery.Query[dbtype.Migration](DBS, "SELECT * FROM type::table(Migrations);", nil)
	if err != nil {
		return fmt.Errorf("Migrate error: %w", err)
	}

	appliedMigrations := results

	isApplied := func(name string) bool {
		for _, applied := range appliedMigrations {
			if applied.Name == name {
				return true
			}
		}
		return false
	}

	// Apply pending migrations
	for _, mig := range AllMigrations {
		fmt.Println("Checking Migration:", mig.Name)
		if isApplied(mig.Name) {
			fmt.Println("Skipping Migration:", mig.Name)
			continue
		}
		fmt.Println("Applying migration", mig.Name)
        // Apply the schema and its history record atomically.
        mig.Date_Created = time.Now().UTC().Format(time.RFC3339)
        _, err := sqlquery.Query[any](DBS, "BEGIN TRANSACTION;"+mig.SQL+"CREATE Migrations CONTENT $migration; COMMIT TRANSACTION;", map[string]interface{}{"migration": mig})
        if err != nil {
            return fmt.Errorf("Migrate error: %s: %w", mig.Name, err)
        }

	}
	return nil
}
