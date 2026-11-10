package server

// The consumer-side wire fixtures for the session-adjacent seams the
// native shells ride: the sessionIdentity answer (both credential
// arms of its exactly-one-credential rule) and the push-device
// registration request the shell ENCODES. The push transports' payload
// fixtures live beside these (internal/push.RenderWireFixtures) —
// rendered through the same pure payload builders the senders marshal.
//
// Everything is byte-pinned by the same freshness gate as the asyncapi
// document (see docs/contracts/wire/); the android shell vendors these
// files and decodes them in JVM tests, so a renamed key reds a LOCAL
// test on either side. The example VALUES are fixed fakes; the KEYS
// are the contract.

import (
	"encoding/json/v2"

	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/server/api"
)

// Fixture example values — the session/push-device seam's share of the
// fixed fake vocabulary (see internal/push/fixtures.go).
const (
	fixtureExtension   = "1001"
	fixtureDID         = "+15551234567"
	fixtureDisplayName = "Jane Operator"
)

// RenderWireFixtureFiles renders every consumer wire fixture the
// `webphone contracts` command writes under docs/contracts/wire/ as
// filename → JSON bytes. Marshal of these structs cannot fail; a stub
// keeps the render path panic-free (mustMarshalRing's doctrine).
func RenderWireFixtureFiles() map[string][]byte {
	legacy := sessionIdentity{
		Extension: fixtureExtension,
		DID:       fixtureDID,
		Password:  "fixture-directory-password",
	}
	brokered := sessionIdentity{
		Extension:     fixtureExtension,
		DID:           fixtureDID,
		DisplayName:   fixtureDisplayName,
		Numbers:       []string{fixtureDID},
		Extensions:    []string{fixtureExtension, "1002"},
		SIPCredential: "fixture-sip-credential",
	}
	register := api.PushDeviceRegisterRequest{
		DeviceID:  "fixture-device-0001",
		Transport: push.TransportNtfy,
		Token:     push.FixtureNtfyTopic,
	}

	files := map[string]any{
		"session-identity-password.json":    legacy,
		"session-identity-brokered.json":    brokered,
		"push-device-register-request.json": register,
	}
	rendered := push.RenderWireFixtures()
	for name, payload := range files {
		body, err := json.Marshal(payload)
		if err != nil {
			rendered[name] = []byte(`{"error":"unrenderable"}`)
			continue
		}
		rendered[name] = body
	}
	return rendered
}
