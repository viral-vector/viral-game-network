package database

import (
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
	// Call the SDK's Query which returns a pointer to a slice of QueryResult[T]
	queryResults, err := surrealdb.Query[T](DBS, query, params)
	if err != nil {
		return nil, err
	}

	// Dereference the pointer
	resultsSlice := *queryResults

	// Extract the actual results from each QueryResult element.
	// This assumes that each QueryResult has a field `Result` of type T.
	var output []T
	for _, qr := range resultsSlice {
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