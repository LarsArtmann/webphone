package api

import (
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// writeJSONBody ends a JSON handler: the content type and the best-effort
// marshal live in one place. The response is already committed by the time
// the marshal can fail, so the failure cannot be repaired.
func writeJSONBody(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.MarshalWrite(w, body) //nolint:erraudit // best-effort write; the response is already committed
}

// storeAtCapacity answers a readmodels cap refusal with the calling
// surface's own noun: the extension's list is full, which the owner fixes
// by deleting an entry. Reports whether the error was a cap refusal.
func storeAtCapacity(w http.ResponseWriter, err error, message string) bool {
	if !errors.Is(err, readmodels.ErrListFull) {
		return false
	}
	http.Error(w, message, http.StatusUnprocessableEntity)
	return true
}
