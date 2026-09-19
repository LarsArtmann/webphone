package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestVoicemailRowsOfferCallBack pins the callback affordance: every
// voicemail row whose sender left a CID number carries a data-dial
// button; anonymous senders render no dead button.
func TestVoicemailRowsOfferCallBack(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(uri, "/phone-api/voicemail/") && strings.HasSuffix(uri, "/summary"):
			_, _ = w.Write([]byte(`{"new":1,"old":1}`))
		case strings.HasPrefix(uri, "/phone-api/voicemail/"):
			_, _ = w.Write([]byte(`{"messages":[
				{"uuid":"vm-1","cid_number":"+441632960961","cid_name":"Alice","seconds":12,"created":1758300000,"read":false,"audio_url":"/phone-api/voicemail/1001/messages/vm-1/audio"},
				{"uuid":"vm-2","cid_number":"","cid_name":"Withheld","seconds":8,"created":1758300100,"read":true,"audio_url":"/phone-api/voicemail/1001/messages/vm-2/audio"}
			]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/voicemail", nil, "")
	page := string(body)
	if !strings.Contains(page, `data-dial="+441632960961"`) {
		t.Errorf("voicemail row missing the callback button: %.400s", page)
	}
	if got := strings.Count(page, "data-dial="); got != 1 {
		t.Errorf("messages without a CID number must render no dial button: %d data-dial attributes", got)
	}
}
