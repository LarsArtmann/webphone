package push

import "encoding/json"

// The consumer-side wire fixtures: byte-pinned examples of every payload
// this seam puts on either transport, rendered through the SAME pure
// payload builders the senders marshal (ringFcmPayload et al.) — the
// pinned contract can never diverge from the sent wire. The native
// shells (webphone-android first) vendor these files into their JVM
// tests and decode them through their own parsers, so field drift reds
// a LOCAL test on either side instead of surfacing at the first device
// boot. Rendered by `webphone contracts` into docs/contracts/wire/ and
// freshness-gated like the asyncapi document.
//
// The example VALUES are fixed constants (stable bytes, greppable); the
// example KEYS are the contract. Nothing here talks to a network.

// Fixture example values — deliberately fixed, deliberately fake.
const (
	FixtureFrom      = "+15550000001"
	FixtureContact   = "+15550000002"
	FixtureCallID    = "call-fixture-1"
	FixtureDraftID   = "draft-fixture-1"
	FixtureFCMToken  = "fixture-device-token"
	FixtureNtfyTopic = "fixture-topic"
)

// RenderWireFixtures renders the push seam's consumer fixtures as
// filename → JSON bytes. Marshal of these all-string structs cannot
// fail; mustMarshalRing's panic-free doctrine covers the same class.
func RenderWireFixtures() map[string][]byte {
	ring := Ring{Event: EventRing, From: FixtureFrom, CallID: FixtureCallID}
	ringEnd := Ring{Event: EventRingEnd, From: FixtureFrom, CallID: FixtureCallID}
	draft := Draft{Event: EventDraftReady, Contact: FixtureContact, DraftID: FixtureDraftID}

	fixtures := map[string]any{
		"push-fcm-ring.json":         ringFcmPayload(FixtureFCMToken, ring),
		"push-fcm-ring-end.json":     ringFcmPayload(FixtureFCMToken, ringEnd),
		"push-fcm-draft-ready.json":  draftFcmPayload(FixtureFCMToken, draft),
		"push-ntfy-ring.json":        ringNtfyBody(FixtureNtfyTopic, ring),
		"push-ntfy-ring-end.json":    ringNtfyBody(FixtureNtfyTopic, ringEnd),
		"push-ntfy-draft-ready.json": draftNtfyBody(FixtureNtfyTopic, draft),
	}

	rendered := make(map[string][]byte, len(fixtures))
	for name, payload := range fixtures {
		body, err := json.Marshal(payload)
		if err != nil {
			// Same class as mustMarshalRing: an all-string struct —
			// render as a stub rather than panic in a render path.
			rendered[name] = []byte(`{"error":"unrenderable"}`)
			continue
		}
		rendered[name] = body
	}
	return rendered
}
