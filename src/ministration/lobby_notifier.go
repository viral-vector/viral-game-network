package ministration

import (
	"encoding/json"
	"strconv"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/pubsub"
)

func Service_Lobby_Notify(lobby_id string, msg string, user_id string) error {
	channel := "lobby:" + lobby_id + ":channel"

	tmp := dbtype.LobbyMessage{
		User_ID:      user_id,
		Date_Created: strconv.FormatInt(time.Now().Unix(), 10),
		Body:         string(msg),
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
	err = cache.Exp(channel, 60*time.Minute)

	return nil
}
