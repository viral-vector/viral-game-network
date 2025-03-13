package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Migration struct {
	ID  *models.RecordID `json:"id,omitempty"`
	Name  string `json:"name"`
	Date_Created string `json:"date_created"`
	SQL  string `json:"sql"`
}

func (u Migration) TableName() string {
	return "Migrations"
}