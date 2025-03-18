package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Lobby struct {
	ID    *models.RecordID `json:"id,omitempty"`
	Name  string `json:"name"`
	Guid  string `json:"guid,omitempty"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Configs string `json:"configs"`
	Private bool `json:"private"`
	Code string `json:"code"`
	Lobby_Application *Application `json:"lobby_application,omitempty"`
	Lobby_Host *User `json:"lobby_host,omitempty"`
	Lobby_Users []*User `json:"lobby_users,omitempty"`
	Lobby_Server *Server `json:"lobby_server,omitempty"` 
}

func (u Lobby) TableName() string {
	return "Lobby"
}