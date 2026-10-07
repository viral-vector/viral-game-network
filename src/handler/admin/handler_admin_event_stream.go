package handler_admin

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"

	"github.com/gofiber/fiber/v3"
)

func Handle_SSEvents(c fiber.Ctx) error {
	expires, ok := c.Locals("session_expires").(time.Time)
	if !ok || !expires.After(time.Now()) {
		return fiber.ErrForbidden
	}
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	// Capture session data before Fiber returns the context to its pool.
	c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
		streamSystemEvents(w, repository.SelSystemEventsSince, 5*time.Second, expires)
	})
	return nil
}

// A single loop owns polling and writing. Heartbeats detect disconnects even
// when no events arrive or storage is unavailable; there is no producer to leak.
func streamSystemEvents(w *bufio.Writer, load func(time.Time) ([]dbtype.SystemEvent, error), interval time.Duration, expires time.Time) {
	if !expires.After(time.Now()) {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	expiry := time.NewTimer(time.Until(expires))
	defer expiry.Stop()
	since := time.Now().Add(-interval)
	type deliveredEvent struct {
		created time.Time
		payload [32]byte
	}
	seen := make(map[string]deliveredEvent)
	for {
		if !expires.After(time.Now()) {
			return
		}
		if _, err := w.WriteString(": keepalive\n\n"); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}
		started := time.Now()
		messages, err := load(since)
		if err == nil {
			for _, message := range messages {
				if !expires.After(time.Now()) {
					return
				}
				id := message.ModelID()
				encoded, err := json.Marshal(message)
				if err != nil {
					continue
				}
				payload := sha256.Sum256(encoded)
				if previous, duplicate := seen[id]; id != "" && duplicate && previous.payload == payload {
					continue
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", encoded); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
				if id != "" {
					created, err := time.Parse(time.RFC3339, message.Date_Created)
					if err != nil {
						created = started
					}
					seen[id] = deliveredEvent{created: created, payload: payload}
				}
			}
			// Overlap by one second because stored timestamps have second precision.
			since = started.Add(-time.Second).Truncate(time.Second)
			for id, delivered := range seen {
				if delivered.created.Before(since) {
					delete(seen, id)
				}
			}
		}
		select {
		case <-ticker.C:
		case <-expiry.C:
			return
		}
	}
}
