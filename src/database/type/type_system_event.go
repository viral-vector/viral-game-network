package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type SystemEvent struct {
	ID           *models.RecordID `json:"id,omitempty"`
	Ref_ID       string `json:"ref_id,omitempty"`
	Ref_Source   string `json:"ref_source,omitempty"`
	Ref_Target   string `json:"ref_target,omitempty"`
	Date_Created string `json:"date_created"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
}

func (n SystemEvent) ModelID() string {
	if n.ID == nil {
		return ""
	}
	return n.ID.String()
}

func (n SystemEvent) TableName() string {
	return "System_Event"
}