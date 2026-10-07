package database

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"time"
	migrations "viral-game-network/src/database/migrations"
	"viral-game-network/src/database/sqlquery"
	dbtype "viral-game-network/src/database/type"
)

var DBS *surrealdb.DB

// Connect initializes storage explicitly so importing a package has no network effects.
// Migrations finish before callers can bootstrap or serve requests.
func Connect(endpoint, namespace, name, username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := surrealdb.FromEndpointURLString(ctx, endpoint)
	if err != nil {
		return err
	}
	if _, err = db.SignIn(ctx, &surrealdb.Auth{Username: username, Password: password}); err != nil {
		db.Close(context.Background())
		return err
	}
	if err = db.Use(ctx, namespace, name); err != nil {
		db.Close(context.Background())
		return err
	}
	if err = migrations.Migrations_Run(db); err != nil {
		db.Close(context.Background())
		return err
	}
	DBS = db
	return nil
}

func Close() {
	if DBS != nil {
		DBS.Close(context.Background())
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rows, err := surrealdb.Create[[]T](ctx, DBS, models.Table(table), record)
	if err != nil {
		return nil, err
	}
	if rows == nil || len(*rows) != 1 {
		return nil, fmt.Errorf("database returned no created record")
	}
	return &(*rows)[0], nil
}

func Update[T dbtype.Model](id models.RecordID, record *T) (*T, error) {
	if DBS == nil {
		return nil, fmt.Errorf("database is not connected")
	}
	if record == nil {
		return nil, fmt.Errorf("missing record")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return surrealdb.Update[T](ctx, DBS, id, record)
}

func Upsert[T dbtype.Model](record *T) (*T, error) {
	if DBS == nil {
		return nil, fmt.Errorf("database is not connected")
	}
	if record == nil {
		return nil, fmt.Errorf("missing record")
	}
	table := (*record).TableName()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	queryResultsSingle, err := surrealdb.Upsert[[]T](ctx, DBS, models.Table(table), record)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := surrealdb.Relate[any](ctx, DBS, relationship)
	return err
}

func GetUUID() string {
	return uuid.New().String()
}
