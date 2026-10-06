package database

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	migrations "viral-game-network/src/database/migrations"
	"viral-game-network/src/database/sqlquery"
	dbtype "viral-game-network/src/database/type"
)

var DBS *surrealdb.DB

// Connect initializes storage explicitly so importing a package has no network effects.
// Migrations finish before callers can bootstrap or serve requests.
func Connect(endpoint, namespace, name, username, password string) error {
	db, err := surrealdb.New(endpoint)
	if err != nil {
		return err
	}
	if _, err = db.SignIn(&surrealdb.Auth{Username: username, Password: password}); err != nil {
		db.Close()
		return err
	}
	if err = db.Use(namespace, name); err != nil {
		db.Close()
		return err
	}
	if err = migrations.Migrations_Run(db); err != nil {
		db.Close()
		return err
	}
	DBS = db
	return nil
}

func Close() {
	if DBS != nil {
		DBS.Close()
		DBS = nil
	}
}

func Query[T any](query string, params map[string]interface{}) ([]T, error) {
	return sqlquery.Query[T](DBS, query, params)
}

func Create[T dbtype.Model](record *T) (*T, error) {
	if DBS == nil {
		return nil, fmt.Errorf("database is not connected")
	}
	if record == nil {
		return nil, fmt.Errorf("missing record")
	}
	table := (*record).TableName()
	return surrealdb.Create[T](DBS, models.Table(table), record)
}

func Update[T dbtype.Model](id models.RecordID, record *T) (*T, error) {
	if DBS == nil {
		return nil, fmt.Errorf("database is not connected")
	}
	if record == nil {
		return nil, fmt.Errorf("missing record")
	}
	return surrealdb.Update[T](DBS, id, record)
}

func Upsert[T dbtype.Model](record *T) (*T, error) {
	if DBS == nil {
		return nil, fmt.Errorf("database is not connected")
	}
	if record == nil {
		return nil, fmt.Errorf("missing record")
	}
	table := (*record).TableName()
	queryResultsSingle, err := surrealdb.Upsert[[]T](DBS, models.Table(table), record)
	if err != nil {
		return nil, err
	}

	if queryResultsSingle == nil || len(*queryResultsSingle) == 0 {
		return nil, fmt.Errorf("database returned no upserted record")
	}
	return &(*queryResultsSingle)[0], nil
}

func Delete[T models.RecordID](id models.RecordID) error {
	_, err := Query[any]("DELETE type::record($id);", map[string]interface{}{
		"id": id.String(),
	})
	return err
}

func Select[T any](id string) (*T, error) {
	result, err := Query[T]("SELECT * FROM type::record($id);", map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

func Relate(in *models.RecordID, out *models.RecordID, table string, data map[string]interface{}) error {
	if DBS == nil {
		return fmt.Errorf("database is not connected")
	}
	if in == nil || out == nil {
		return fmt.Errorf("missing relationship endpoint")
	}
	// Create a new relationship.
	relationship := &surrealdb.Relationship{
		In:       *in,
		Out:      *out,
		Relation: models.Table(table),
		Data:     data,
	}
	return surrealdb.Relate(DBS, relationship)
}

func GetUUID() string {
	return uuid.New().String()
}
