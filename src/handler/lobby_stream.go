package handler

import (
	"encoding/json"
	"sync"
	dbtype "viral-game-network/src/database/type"
)

// lobbyLiveWriter suppresses the overlap between the history snapshot and the
// already-open subscription. Only snapshot IDs are retained, so memory stays
// bounded by replay history even for long-lived connections.
func lobbyLiveWriter(history []string, send func([]byte) error) func([]byte) error {
	replayed := make(map[string]struct{}, len(history))
	messageID := func(raw []byte) string {
		var message dbtype.LobbyMessage
		if json.Unmarshal(raw, &message) != nil {
			return ""
		}
		return message.ModelID()
	}
	for _, raw := range history {
		if id := messageID([]byte(raw)); id != "" {
			replayed[id] = struct{}{}
		}
	}
	var writer sync.Mutex
	return func(raw []byte) error {
		writer.Lock()
		defer writer.Unlock()
		if id := messageID(raw); id != "" {
			if _, exists := replayed[id]; exists {
				delete(replayed, id)
				return nil
			}
		}
		return send(raw)
	}
}
