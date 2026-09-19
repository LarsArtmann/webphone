package server

import (
	"encoding/json/v2"
	"net/http"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
)

// toastDetail is the library's HX-Trigger wire shape ({message, kind} under
// the "showMessage" event) — aliased, not re-declared, so a shape change
// upstream fails this build instead of silently breaking the island's toast
// listener (the same aliasing adminui and dashboardui use).
type toastDetail = cqrshtmx.ToastDetail

// notifyToast queues an island toast through the HX-Trigger response
// header. Must run BEFORE the response body renders — headers stop riding
// along once the partial has been written. Kinds follow the island's
// vocabulary: "ok" for success paths, "error" for failures.
func notifyToast(w http.ResponseWriter, kind, message string) {
	payload, err := json.Marshal(map[string]toastDetail{
		"showMessage": {Message: message, Kind: kind},
	})
	if err != nil {
		return // header is best-effort; the swap itself is the primary feedback
	}
	w.Header().Set("HX-Trigger", string(payload))
}
