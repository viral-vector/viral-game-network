package database

import (
	"fmt"
	"log"
	"os"
	"time"
	dbtype "viral-game-network/src/database/type"
	migrations "viral-game-network/src/database/migrations"
    "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"github.com/google/uuid"
)

var DBS *surrealdb.DB
var err error

func init() {
	log.Println("Database Initialize!")

	DBS, err = surrealdb.New(os.Getenv("STORE_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}

	if err = DBS.Use(os.Getenv("STORE_DATABASE"), "vgn"); err != nil {
		log.Fatal(err)
	}

	authData := &surrealdb.Auth{
		Username: os.Getenv("STORE_USERNAME"), // use your setup username
		Password: os.Getenv("STORE_PASSWORD"), // use your setup password
	}
	
	if _, err := DBS.SignIn(authData); err != nil {
		log.Fatal(err)
	}

	// Run Migration
	go func() {
		time.Sleep(3 * time.Second)

		err := migrations.Migrations_Run(DBS)
		if err != nil{
			log.Println("Migrations_Run Error!", err)
		}
	}()
	
	log.Println("Database Connected!")
}

func Query[T any](query string, params map[string]interface{}) ([]T, error) {
	// First attempt: decode each QueryResult.Result as []T.
	queryResultsSlice, err := surrealdb.Query[[]T](DBS, query, params)
	if err == nil {
		resultsSlice := *queryResultsSlice
		var output []T
		for _, qr := range resultsSlice {
			// Expecting qr.Result to be of type []T, so flatten it.
			output = append(output, qr.Result...)
		}
		return output, nil
	}

	// If the first decoding fails, try to decode as a single T.
	queryResultsSingle, err2 := surrealdb.Query[T](DBS, query, params)
	if err2 != nil {
		// Return the original error if both attempts fail.
		return nil, fmt.Errorf("failed decoding as []T: %w; also failed decoding as T: %v", err, err2)
	}

	resultsSingle := *queryResultsSingle
	var output []T
	for _, qr := range resultsSingle {
		// Here qr.Result is a single T value; append it directly.
		output = append(output, qr.Result)
	}
	return output, nil
}

func Create[T dbtype.Model](record *T) (*T, error) {
	table := (*record).TableName()
	return surrealdb.Create[T](DBS, models.Table(table), record)
}

func Update[T dbtype.Model](id models.RecordID, record *T) (*T, error) {
	return surrealdb.Update[T](DBS, id, record)
}

func Upsert[T dbtype.Model](record *T) (*T, error) {
	table := (*record).TableName()
	queryResultsSingle, err := surrealdb.Upsert[[]T](DBS, models.Table(table), record)
	if err != nil {
		return nil, err
	}

	return &(*queryResultsSingle)[0], nil
}

func Delete[T models.RecordID](id models.RecordID) (error) {
	_, err := surrealdb.Query[any](DBS, "DELETE type::record($id);", map[string]interface{}{
		"id": id.String(), 
	})
	if err != nil {
		log.Panicf("Database Delete Error: ", err)
	}
	return err
}

func Select[T any](id string) (*T, error) {
	result, err := surrealdb.Query[T](DBS, "SELECT * FROM ONLY type::record($id);", map[string]interface{}{
		"id": id,
	})
	if err != nil {
		log.Panicf("Database Select Error: ", err)
		return nil, err
	}
	return &(*result)[0].Result, err
}

func Relate(in *models.RecordID, out *models.RecordID, table string, data map[string]interface{}) (error) {
	// Create a new relationship.
	relationship := &surrealdb.Relationship{
		In:       *in,
		Out:      *out,
		Relation: models.Table(table),
		Data: data,
	}
	err := surrealdb.Relate(DBS, relationship)
	if err != nil {
		log.Panicf("Database Relate Error: ", err)
	}
	return err
}

func GetUUID() string {
	return uuid.New().String()
}