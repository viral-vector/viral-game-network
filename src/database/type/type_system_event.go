package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type SystemEvent struct {
	ID           *models.RecordID `json:"id,omitempty"`
	Date_Created string `json:"date_created"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
}

func (u SystemEvent) TableName() string {
	return "System_Event"
}