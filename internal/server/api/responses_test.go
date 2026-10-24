package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/webphone/internal/readmodels"
)

func TestStoreAtCapacityAnswersOnlyTheCapRefusal(t *testing.T) {
	t.Run("cap refusal", func(t *testing.T) {
		rec := httptest.NewRecorder()
		if handled := storeAtCapacity(rec, readmodels.ErrListFull, "list is full"); !handled {
			t.Fatal("storeAtCapacity = false, want true for ErrListFull")
		}
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
		}
		if body := rec.Body.String(); body != "list is full\n" {
			t.Fatalf("body = %q, want the caller's noun", body)
		}
	})

	t.Run("other error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		if handled := storeAtCapacity(rec, errors.New("boom"), "list is full"); handled {
			t.Fatal("storeAtCapacity = true for an unrelated error")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want no write", rec.Code)
		}
	})
}

func TestWriteJSONBodySetsTheJSONContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSONBody(rec, map[string]string{"ok": "1"})
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if body := rec.Body.String(); body != `{"ok":"1"}` {
		t.Fatalf("body = %q", body)
	}
}
