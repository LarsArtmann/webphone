package app

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
)

func captureBufLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

// TestLoggingDeadLetterStoreWarnsOnPoisonLanding pins the poison-moment
// surface: the library's sendToDLQ is silent, so the decorator's WARN is
// the only immediate trace a poisoned projection leaves in the journal.
func TestLoggingDeadLetterStoreWarnsOnPoisonLanding(t *testing.T) {
	log, buf := captureBufLogger()
	decorated := loggingDeadLetterStore{DeadLetterStore: projectionhost.NewMemoryDeadLetterStore(), log: log}

	entry := projectionhost.DeadLetterEntry{
		ProjectionName: "messages-projection",
		EventID:        "evt-1",
		EventType:      "MessageReceived",
		Error:          "fold: payload decode failed",
		FailedAt:       time.Now(),
	}
	if err := decorated.Store(context.Background(), entry); err != nil {
		t.Fatalf("Store through decorator: %v", err)
	}

	out := buf.String()
	for _, needle := range []string{
		"projection dead-lettered an event",
		"messages-projection",
		"MessageReceived",
		"evt-1",
		"projection_dead_letters",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("poison WARN must carry %q, got:\n%s", needle, out)
		}
	}
}

// TestWarnDLQDepth pins the restart surface: durable entries from a
// previous boot re-WARN at Start, an empty queue stays silent, and a
// store without the admin surface skips instead of failing.
func TestWarnDLQDepth(t *testing.T) {
	ctx := context.Background()

	t.Run("seeded queue names count and oldest age", func(t *testing.T) {
		log, buf := captureBufLogger()
		store := projectionhost.NewMemoryDeadLetterStore()
		old := projectionhost.DeadLetterEntry{ProjectionName: "p", EventID: "old", FailedAt: time.Now().Add(-2 * time.Hour)}
		recent := projectionhost.DeadLetterEntry{ProjectionName: "p", EventID: "recent", FailedAt: time.Now().Add(-5 * time.Minute)}
		for _, entry := range []projectionhost.DeadLetterEntry{old, recent} {
			if err := store.Store(ctx, entry); err != nil {
				t.Fatalf("seed: %v", err)
			}
		}

		warnDLQDepth(ctx, store, log)

		out := buf.String()
		for _, needle := range []string{
			"projection dead-letter queue carries poison events",
			"count=2",
			"oldest_age=",
		} {
			if !strings.Contains(out, needle) {
				t.Errorf("depth WARN must carry %q, got:\n%s", needle, out)
			}
		}
	})

	t.Run("empty queue stays silent", func(t *testing.T) {
		log, buf := captureBufLogger()
		warnDLQDepth(ctx, projectionhost.NewMemoryDeadLetterStore(), log)
		if buf.Len() != 0 {
			t.Fatalf("empty queue must not WARN, got:\n%s", buf.String())
		}
	})

	t.Run("store without the admin surface skips silently", func(t *testing.T) {
		log, buf := captureBufLogger()
		// Interface embedding yields exactly DeadLetterStore's method
		// set: this value deliberately fails the Admin type-assert.
		nonAdmin := struct{ projectionhost.DeadLetterStore }{projectionhost.NewMemoryDeadLetterStore()}
		if err := nonAdmin.Store(ctx, projectionhost.DeadLetterEntry{ProjectionName: "p", EventID: "x", FailedAt: time.Now()}); err != nil {
			t.Fatalf("seed non-admin store: %v", err)
		}
		warnDLQDepth(ctx, nonAdmin, log)
		if buf.Len() != 0 {
			t.Fatalf("non-admin store must skip, not WARN, got:\n%s", buf.String())
		}
	})
}
