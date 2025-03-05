package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type LobbyUser struct {
	ID    *models.RecordID `json:"id,omitempty"`
	User  User `json:"user"`
	Date_Created string `json:"date_created"`
}

func (u LobbyUser) TableName() string {
	return "Lobby_User"
}