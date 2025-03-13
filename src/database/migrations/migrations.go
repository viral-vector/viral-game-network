package migrations

import (
	"fmt"
	"log"
	"time"
	dbtype "viral-game-network/src/database/type"
    "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// AllMigrations aggregates all migration steps.
var AllMigrations = []dbtype.Migration{}

// Run All(pending) Migrations
func Migrate(DBS *surrealdb.DB) error {
	log.Println("Checking Migrations")

	// Ensure that the Migrations table exists
	MigrationsTableScaffold := []string {
		`
		DEFINE TABLE Migrations SCHEMAFULL;
		DEFINE FIELD id ON TABLE Migrations TYPE string;
		DEFINE FIELD name ON TABLE Migrations TYPE string;
		DEFINE FIELD date_created ON TABLE Migrations TYPE string;
		DEFINE FIELD sql ON TABLE Migrations TYPE string;
		DEFINE INDEX idx_migrations_id ON TABLE Migrations COLUMNS id UNIQUE;
		DEFINE INDEX idx_migrations_name ON TABLE Migrations COLUMNS name UNIQUE;
		`,
	}

	for _, mig := range MigrationsTableScaffold {
		_, err := surrealdb.Query[any](DBS, mig, nil)
		if err != nil {
			return fmt.Errorf("Migrate error: %w", err)
		}
	} 

	// Retrieve applied migrations from the Migrations table
	results, err := surrealdb.Query[[]dbtype.Migration](DBS, "SELECT * FROM type::table(Migrations);", nil)
	if err != nil {
		return fmt.Errorf("Migrate error: %w", err)
	}

	var appliedMigrations []dbtype.Migration
	for _, qr := range *results {
		appliedMigrations = append(appliedMigrations, qr.Result...)
	}

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
		_, err := surrealdb.Query[any](DBS, mig.SQL, nil)
		if err != nil {
			return fmt.Errorf("Migrate error: %s: %w", mig.Name, err)
		}
 
		// Record the migration as applied.
		mig.Date_Created = time.Now().UTC().Format(time.RFC3339)
		_, err = surrealdb.Create[any](DBS, models.Table("Migrations"), mig)
		if err != nil {
			return fmt.Errorf("Migrate error: %s: %w", mig.Name, err)
		}
	}
	return nil
}
