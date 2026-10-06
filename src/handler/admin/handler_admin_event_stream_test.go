package handler_admin

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
	dbtype "viral-game-network/src/database/type"

	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type disconnectedStream struct {
	bytes.Buffer
	writes int
	failAt int
}

func (s *disconnectedStream) Write(p []byte) (int, error) {
	s.writes++
	if s.writes == s.failAt {
		return 0, io.ErrClosedPipe
	}
	return s.Buffer.Write(p)
}

func TestEventStreamDetectsQuietDisconnect(t *testing.T) {
	out := &disconnectedStream{failAt: 2}
	calls := 0
	streamSystemEvents(bufio.NewWriter(out), func(time.Time) ([]dbtype.SystemEvent, error) {
		calls++
		return nil, nil
	}, time.Millisecond, time.Now().Add(time.Second))
	if calls != 1 || out.writes != 2 || !strings.Contains(out.String(), ": keepalive") {
		t.Fatalf("quiet stream did not stop: calls=%d writes=%d", calls, out.writes)
	}
}

func TestEventStreamErrorsWaitForNextPoll(t *testing.T) {
	out := &disconnectedStream{failAt: 3}
	calls := 0
	started := time.Now()
	streamSystemEvents(bufio.NewWriter(out), func(time.Time) ([]dbtype.SystemEvent, error) {
		calls++
		return nil, errors.New("store unavailable")
	}, 10*time.Millisecond, time.Now().Add(time.Second))
	if calls != 2 || time.Since(started) < 15*time.Millisecond {
		t.Fatalf("query failure spins: calls=%d elapsed=%s", calls, time.Since(started))
	}
}

func TestEventStreamDeduplicatesOverlappingQueries(t *testing.T) {
	out := &disconnectedStream{failAt: 4}
	message := dbtype.SystemEvent{ID: &models.RecordID{Table: "System_Event", ID: "event"}, Date_Created: time.Now().UTC().Format(time.RFC3339), Message: "Once"}
	streamSystemEvents(bufio.NewWriter(out), func(time.Time) ([]dbtype.SystemEvent, error) {
		return []dbtype.SystemEvent{message}, nil
	}, time.Millisecond, time.Now().Add(time.Second))
	if count := strings.Count(out.String(), "data: "); count != 1 {
		t.Fatalf("event delivered %d times: %s", count, out.String())
	}
}

func TestEventStreamEndsAtSessionExpiration(t *testing.T) {
	var out bytes.Buffer
	started := time.Now()
	streamSystemEvents(bufio.NewWriter(&out), func(time.Time) ([]dbtype.SystemEvent, error) { return nil, nil }, time.Second, started.Add(20*time.Millisecond))
	if time.Since(started) >= time.Second {
		t.Fatal("stream remained open after session expired")
	}
	if !strings.Contains(out.String(), ": keepalive") {
		t.Fatal("live session did not start streaming")
	}
	out.Reset()
	streamSystemEvents(bufio.NewWriter(&out), func(time.Time) ([]dbtype.SystemEvent, error) {
		t.Fatal("expired session polled storage")
		return nil, nil
	}, time.Millisecond, time.Now().Add(-time.Second))
	if out.Len() != 0 {
		t.Fatal("expired session started streaming")
	}
}

func TestEventStreamDeliversUpdatedRecordWithinSameSecond(t *testing.T) {
	out := &disconnectedStream{failAt: 6}
	message := dbtype.SystemEvent{ID: &models.RecordID{Table: "System_Event", ID: "event"}, Date_Created: time.Now().UTC().Format(time.RFC3339), Message: "Pending"}
	calls := 0
	streamSystemEvents(bufio.NewWriter(out), func(time.Time) ([]dbtype.SystemEvent, error) {
		calls++
		if calls > 1 {
			message.Message = "Failed"
		}
		return []dbtype.SystemEvent{message}, nil
	}, time.Millisecond, time.Now().Add(time.Second))
	if count := strings.Count(out.String(), "data: "); count != 2 || !strings.Contains(out.String(), "Pending") || !strings.Contains(out.String(), "Failed") {
		t.Fatalf("updated event suppressed or duplicated: %s", out.String())
	}
}
