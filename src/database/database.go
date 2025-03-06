package database

import (
	"fmt"
	"log"
	"os"
    "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

var DBS *surrealdb.DB
var err error

type Model interface {
	TableName() string
}

func init() {
	log.Println("Database Initialize!")

	DBS, err = surrealdb.New(os.Getenv("STORE_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}

	if err = DBS.Use(os.Getenv("APP_NAME"), "vgn"); err != nil {
		log.Fatal(err)
	}

	authData := &surrealdb.Auth{
		Username: os.Getenv("STORE_USERNAME"), // use your setup username
		Password: os.Getenv("STORE_PASSWORD"), // use your setup password
	}
	
	if _, err := DBS.SignIn(authData); err != nil {
		log.Fatal(err)
	}
	
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

func Create[T Model](record *T) (*T, error) {
	table := (*record).TableName()
	return surrealdb.Create[T](DBS, models.Table(table), record)
}

func Update[T Model](record *T) (*T, error) {
	table := (*record).TableName()
	return surrealdb.Update[T](DBS, models.Table(table), record)
}

func Delete[T any](id string) (error) {
	_, err := surrealdb.Delete[T, string](DBS, id)
	return err
}

func Select[T any](id string) (*T, error) {
	record, err := surrealdb.Select[T, string](DBS, id)
	if err != nil {
		return nil, err
	}
	return record, nil
}