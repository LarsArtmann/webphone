package server

import (
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/larsartmann/webphone/internal/asr"
)

// transcribeBodyLimit caps one audio segment. MediaRecorder chunks are
// tens of KB; a stored voicemail clip is a few MB; the bound exists so a
// runaway client cannot stream unbounded bytes into the provider seam.
const transcribeBodyLimit = 25 << 20

// apiTranscribe is the ONE server endpoint behind every transcription
// surface (live calls, voicemail, MMS audio): the island and shell POST a
// whole audio segment here and get the text back. Session-gated (the
// signed-in extension is the owner by construction) and CSRF-protected
// like the other /api JSON posts. A disabled seam answers 404 — the
// surface simply does not exist, and the island hides every affordance
// behind the same flag.
func (h *handlers) apiTranscribe(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireSession(w, r); !ok {
		return
	}
	if h.deps.ASR == nil || !h.deps.ASR.Enabled() {
		http.NotFound(w, r)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, transcribeBodyLimit)
	audio, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read the audio", http.StatusBadRequest)
		return
	}
	if len(audio) == 0 {
		http.Error(w, "no audio to transcribe", http.StatusBadRequest)
		return
	}

	text, err := h.deps.ASR.Transcribe(r.Context(), asr.Request{
		Audio:       audio,
		Filename:    r.URL.Query().Get("filename"),
		ContentType: r.Header.Get("Content-Type"),
		Language:    r.URL.Query().Get("lang"),
	})
	if err != nil {
		h.transcribeFailed(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.MarshalWrite(w, struct { //nolint:erraudit // best-effort write; the response is already committed
		Text string `json:"text"`
	}{Text: text})
}

// transcribeFailed maps a provider failure to one visible surface. The
// provider is an external dependency, so every failure is a 502 from the
// user's point of view (the text could not be produced); the operator
// gets the classified cause in the log. Never a 401: an upstream
// credential problem is the deployment's, not the signed-in user's.
func (h *handlers) transcribeFailed(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, asr.ErrUnauthorized):
		slog.Warn("asr: provider rejected the configured credentials", "error", err)
		http.Error(w, "transcription provider rejected the configured credentials", http.StatusBadGateway)
	default:
		slog.Warn("asr: transcription failed", "error", err)
		http.Error(w, "transcription failed", http.StatusBadGateway)
	}
}
