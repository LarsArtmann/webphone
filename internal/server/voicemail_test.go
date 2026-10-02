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
	if !strings.Contains(page, `data-sms="+441632960961"`) || !strings.Contains(page, `data-save-contact="+441632960961"`) {
		t.Errorf("voicemail row missing the sms/save-contact affordances: %.400s", page)
	}
	if got := strings.Count(page, "data-dial="); got != 1 {
		t.Errorf("messages without a CID number must render no dial button: %d data-dial attributes", got)
	}
}

// TestVoicemailRowsCarryStableMorphIds pins the im-preserve audit's one
// stateful morph surface: the voicemail panel re-fetches its partial into
// #tab-content on every "voicemail" nudge (hx-swap morph), and idiomorph
// keeps an element whose id exists in BOTH trees morphing it in place —
// so rows that move and an <audio> mid-playback survive the swap instead
// of being re-created (playback would reset). The uuid-derived ids ARE
// the opt-out; dropping them silently reintroduces the re-creation bug.
func TestVoicemailRowsCarryStableMorphIds(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(uri, "/phone-api/voicemail/") && strings.HasSuffix(uri, "/summary"):
			_, _ = w.Write([]byte(`{"new":1,"old":0}`))
		case strings.HasPrefix(uri, "/phone-api/voicemail/"):
			_, _ = w.Write([]byte(`{"messages":[
				{"uuid":"vm-1","cid_number":"+441632960961","cid_name":"Alice","seconds":12,"created":1758300000,"read":false,"audio_url":"/phone-api/voicemail/1001/messages/vm-1/audio"}
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
	for _, want := range []string{
		`id="vm-vm-1"`,
		`id="vm-audio-vm-1"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("voicemail partial missing %s: idiomorph cannot persist the node across the re-fetch morph: %.400s", want, page)
		}
	}
}

// TestVoicemailRowsRenderPlayerChrome pins the M13 player contract: the
// audio element ships WITHOUT native controls (the custom chrome drives
// it), the play/speed buttons and waveform canvas carry the uuid ids the
// shell wires, and both label variants ride the server-rendered data
// attributes so the per-extension language survives the JS toggle.
func TestVoicemailRowsRenderPlayerChrome(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(uri, "/phone-api/voicemail/") && strings.HasSuffix(uri, "/summary"):
			_, _ = w.Write([]byte(`{"new":1,"old":0}`))
		case strings.HasPrefix(uri, "/phone-api/voicemail/"):
			_, _ = w.Write([]byte(`{"messages":[
				{"uuid":"vm-1","cid_number":"+441632960961","cid_name":"Alice","seconds":12,"created":1758300000,"read":false,"audio_url":"/phone-api/voicemail/1001/messages/vm-1/audio"},
				{"uuid":"vm-3","cid_number":"","cid_name":"","seconds":5,"created":1758300200,"read":true,"audio_url":""}
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
	for _, want := range []string{
		`id="vm-play-vm-1"`,
		`data-vm-play="vm-1"`,
		`data-label-play="Play"`,
		`data-label-pause="Pause"`,
		`id="vm-speed-vm-1"`,
		`data-vm-speed="vm-1"`,
		`id="vm-wave-vm-1"`,
		`role="slider"`,
		`aria-label="Position in message"`,
		`id="vm-time-vm-1"`,
		`0:00 / 0:12`,
		`aria-valuemax="12"`,
		`class="wp-vm-player"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("player chrome missing %s: %.600s", want, page)
		}
	}
	if strings.Contains(page, `<audio class="wp-audio" controls`) {
		t.Error("the audio element must ship controls-free; the native UI is only the fallback")
	}
	if strings.Contains(page, "vm-play-vm-3") || strings.Contains(page, "vm-audio-vm-3") {
		t.Error("a row without audio_url must not render player chrome")
	}
}
