package server

import (
	"encoding/json/v2"
	"strings"
	"testing"
)

// TestWireFixtureFilesPinTheConsumerContract decodes every rendered
// fixture and pins the keys the native shell's parsers read — a
// renamed key must red HERE (server side) and in the shell's vendored
// copy (its JVM tests), never at the first device boot.
func TestWireFixtureFilesPinTheConsumerContract(t *testing.T) {
	files := RenderWireFixtureFiles()

	wantNames := []string{
		"push-device-register-request.json",
		"push-fcm-draft-ready.json",
		"push-fcm-ring-end.json",
		"push-fcm-ring.json",
		"push-ntfy-draft-ready.json",
		"push-ntfy-ring-end.json",
		"push-ntfy-ring.json",
		"session-identity-brokered.json",
		"session-identity-password.json",
	}
	if len(files) != len(wantNames) {
		t.Fatalf("fixture set has %d files, want %d: %v", len(files), len(wantNames), files)
	}
	for _, name := range wantNames {
		body, ok := files[name]
		if !ok {
			t.Fatalf("fixture %s missing", name)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("fixture %s does not decode: %v", name, err)
		}
	}

	// The exactly-one-credential rule: each session arm carries its
	// credential field and NOT the other's.
	for name, credential := range map[string]string{
		"session-identity-password.json": "password",
		"session-identity-brokered.json": "sip_credential",
	} {
		body := string(files[name])
		if !strings.Contains(body, `"`+credential+`":"`) {
			t.Errorf("%s lacks its credential field %q: %s", name, credential, body)
		}
		other := "sip_credential"
		if credential == "sip_credential" {
			other = "password"
		}
		if strings.Contains(body, `"`+other+`"`) {
			t.Errorf("%s carries BOTH credential fields (%q present too): %s", name, other, body)
		}
		for _, key := range []string{"extension", "did"} {
			if !strings.Contains(body, `"`+key+`":`) {
				t.Errorf("%s lacks key %q: %s", name, key, body)
			}
		}
	}

	// The ring fact's keys on both transports — the shell's router keys.
	// FCM carries them flat in data; ntfy embeds the fact as compact JSON
	// inside the message field, so that string is parsed, not grepped.
	for name, body := range files {
		if !strings.HasPrefix(name, "push-fcm-") {
			continue
		}
		if !strings.Contains(string(body), `"event":"`) {
			t.Errorf("%s lacks the event key: %s", name, body)
		}
	}
	for name, body := range files {
		if !strings.HasPrefix(name, "push-ntfy-") {
			continue
		}
		var envelope struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			t.Fatalf("%s does not decode: %v", name, err)
		}
		var fact map[string]any
		if err := json.Unmarshal([]byte(envelope.Message), &fact); err != nil {
			t.Errorf("%s message field is not the compact-JSON fact (%q): %v", name, envelope.Message, err)
			continue
		}
		if _, ok := fact["event"]; !ok {
			t.Errorf("%s inner fact lacks the event key: %s", name, envelope.Message)
		}
	}

	// The registration request's keys — the shell ENCODES these.
	register := string(files["push-device-register-request.json"])
	for _, key := range []string{"device_id", "transport", "token"} {
		if !strings.Contains(register, `"`+key+`":`) {
			t.Errorf("register fixture lacks key %q: %s", key, register)
		}
	}
}
