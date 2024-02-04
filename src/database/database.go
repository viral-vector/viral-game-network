package database

import (
	"log"
	"os"
    "github.com/surrealdb/surrealdb.go"
)

var DBS *surrealdb.DB
var err error

func init() {
	log.Println("Database Initialize!")

	DBS, err = surrealdb.New(os.Getenv("STORE_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}

	if _, err = DBS.Signin(map[string]interface{}{
		"user": os.Getenv("STORE_USERNAME"),
		"pass": os.Getenv("STORE_PASSWORD"),
	}); err != nil {
		log.Fatal(err)
	}

	if _, err = DBS.Use("test", "test"); err != nil {
		log.Fatal(err)
	}

	log.Println("Database Connected!")
}