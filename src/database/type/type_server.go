package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Server struct {
	ID      *models.RecordID `json:"id,omitempty"`
	Name    string `json:"name"`
	Guid 	string `json:"guid"`
	Status  string `json:"status"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Address string `json:"address"`
	Port int32 `json:"port"`
	Lobby *Lobby `json:"lobby,omitempty"`
}

func (u Server) TableName() string {
	return "Server"
}