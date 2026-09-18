package server

import (
	"encoding/json/v2"
	"net/http"
)

// toastDetail mirrors cqrshtmx.ToastDetail — the HX-Trigger wire shape the
// island's toast listener consumes ({message, kind} under the
// "showMessage" event). Kept local so the server package does not need
// the dispatch-layer types for one struct.
type toastDetail struct {
	Message string `json:"message"`
	Kind    string `json:"kind"`
}

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
