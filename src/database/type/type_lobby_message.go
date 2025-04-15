package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type LobbyMessage struct {
	ID    *models.RecordID `json:"id,omitempty"`
	User_ID  string `json:"user_id"`
	Date_Created string `json:"date_created"`
	Body string `json:"body"`
}

func (n LobbyMessage) ModelID() string {
	return n.ID.String()
}

func (n LobbyMessage) TableName() string {
	return "Lobby_Message"
}