package server

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// transcriptBodyLimit caps one segment POST: a JSON envelope around a
// few seconds of TEXT (the audio already went through /api/transcribe),
// so single-digit KiB is the honest ceiling.
const transcriptBodyLimit = 16 << 10

// transcriptTextMax bounds one stored segment: the ASR seam returns
// trimmed plain text for ~4 s of speech; anything far beyond this is a
// misbehaving client, not a transcript.
const transcriptTextMax = 4000

// apiSaveTranscript persists one live-call transcript segment to the
// owner-scoped store behind the History tab's transcript section. The
// island's capture loop posts fire-and-forget: 204 whether stored or
// deliberately shaped, 4xx on island bugs (surfaced in dev), 503 when
// the store is unavailable. Gated on the same ASR flag as
// /api/transcribe — without the seam nothing produces segments, so the
// surface does not exist either.
func (h *handlers) apiSaveTranscript(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	if h.deps.ASR == nil || !h.deps.ASR.Enabled() {
		http.NotFound(w, r)
		return
	}
	if h.deps.Transcripts == nil {
		http.Error(w, "transcript storage is unavailable", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		CallID    string `json:"callId"`
		Direction string `json:"direction"`
		Remote    string `json:"remote"`
		StartedAt int64  `json:"startedAt"`
		Text      string `json:"text"`
	}
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, transcriptBodyLimit), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Direction != string(domain.DirectionInbound) && body.Direction != string(domain.DirectionOutbound) {
		http.Error(w, "direction must be in or out", http.StatusBadRequest)
		return
	}
	body.CallID = strings.TrimSpace(body.CallID)
	body.Remote = strings.TrimSpace(body.Remote)
	body.Text = strings.TrimSpace(body.Text)
	if body.CallID == "" || len(body.CallID) > 128 {
		http.Error(w, "callId is required", http.StatusBadRequest)
		return
	}
	if body.StartedAt <= 0 {
		http.Error(w, "startedAt is required", http.StatusBadRequest)
		return
	}
	if body.Remote == "" || len(body.Remote) > 64 {
		http.Error(w, "remote is required", http.StatusBadRequest)
		return
	}
	if body.Text == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}
	if len(body.Text) > transcriptTextMax {
		http.Error(w, "transcript segment too long", http.StatusRequestEntityTooLarge)
		return
	}

	if err := h.deps.Transcripts.Append(r.Context(), domain.CallTranscriptSegment{
		Owner:     sess.Extension,
		CallID:    body.CallID,
		Direction: domain.Direction(body.Direction),
		Remote:    body.Remote,
		StartedAt: time.UnixMilli(body.StartedAt),
		Text:      body.Text,
	}); err != nil {
		slog.Warn("transcript save failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "could not save the transcript", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
