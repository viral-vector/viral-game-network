package service

import (
	"time"
	"strconv"
	"encoding/json"
	"github.com/google/uuid"
	"viral-game-network/src/cache"
	"viral-game-network/src/pubsub"
	"viral-game-network/src/database/type"
)


func Service_Lobby_Notify(lobby_id string, msg string, user_id string) error {
	
	channel := "lobby:"+lobby_id+":channel" 
	
	tmp := dbtype.LobbyMessage{
		ID: uuid.New().String(),
		User_ID: user_id,
		Date_Created: strconv.FormatInt(time.Now().Unix(), 10), 
		Body: string(msg),
	}
	fin, err := json.Marshal(tmp)
	if err != nil {
		return err
	}
	err = pubsub.Pub(channel, fin)
	if err != nil {
		return err
	}
	err = cache.Add(channel, string(fin))
	err = cache.Exp(channel, 60 * time.Minute)

	return nil
}