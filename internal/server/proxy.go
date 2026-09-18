package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/larsartmann/webphone/internal/session"
)

// proxyPhoneAPI forwards the island's /phone-api/* calls to the configured
// upstream, injecting Basic auth from the server session. The island's
// panels (voicemail, history) keep working byte-for-byte: same paths, same
// JSON, no password in the browser beyond the island's own memory copy.
func (h *handlers) proxyPhoneAPI(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	if !h.deps.PhoneAPI.Enabled() {
		http.Error(w, "phone api not configured", http.StatusServiceUnavailable)
		return
	}

	rest := strings.TrimPrefix(r.URL.Path, "/phone-api")
	upstream := h.deps.PhoneAPI.ResolvePath(rest, r.URL.RawQuery)

	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstream, nil)
	if err != nil {
		http.Error(w, "build upstream request", http.StatusBadGateway)
		return
	}
	req.SetBasicAuth(sess.Extension.String(), sess.Password)
	req.Header.Set("Accept", "application/json")
	if r.Body != nil && r.Body != http.NoBody {
		req.Body = io.NopCloser(io.LimitReader(r.Body, 1<<20))
	}

	resp, err := h.deps.PhoneAPI.HTTPClient().Do(req)
	if err != nil {
		http.Error(w, "phone api unreachable", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	for _, name := range []string{"Content-Type", "Cache-Control"} {
		if value := resp.Header.Get(name); value != "" {
			w.Header().Set(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body) //nolint:erraudit // best-effort write; the response is already committed

	// The island polls its voicemail after registration and after each
	// ended call — exactly when a deposit may have landed. Forwarding
	// those reads as "voicemail" nudges keeps an open voicemail tab live
	// without any server-side polling of the PBX. The nudge carries no
	// payload; the panel re-fetches its partial on receipt.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && strings.HasPrefix(rest, "/voicemail/") {
		h.deps.Hubs.Publish(sess.Extension, sseEventVoicemail, "")
	}
}
