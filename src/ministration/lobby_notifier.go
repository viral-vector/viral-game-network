package ministration

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"strconv"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
)

func Service_Lobby_Notify(lobby_id string, msg string, user_id string) error {
	channel := "lobby:" + lobby_id + ":channel"

	tmp := dbtype.LobbyMessage{
		ID:           &models.RecordID{Table: "Lobby_Message", ID: uuid.NewString()},
		User_ID:      user_id,
		Date_Created: strconv.FormatInt(time.Now().Unix(), 10),
		Body:         string(msg),
	}
	fin, err := json.Marshal(tmp)
	if err != nil {
		return err
	}
	return cache.PublishHistory(channel, string(fin), 60*time.Minute)
}
