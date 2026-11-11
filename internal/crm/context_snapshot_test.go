package crm

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

// The snapshot in testdata/kith_context_snapshot.json is the producer's
// RENDERED bytes for one seeded scenario, captured from the CRM's own
// serializer (crm@85c7f2e, internal/machineapi/context.go contextResponse,
// json/v2 MarshalWrite): contact Ada Lovelace with company, one open deal
// (Retainer 2027, 1200000 cents EUR, proposal), two follow-ups with the
// earlier one overdue, and one journal call for the last-touch timestamp.
// Nothing was substituted — the contact id is the ephemeral projection id of
// the capture run. Re-record it the same way when the producer's shape moves,
// then reconcile this file and the fixture below in the same commit.

// TestKithContextSnapshotParsesToTheMappedContext proves the producer's real
// bytes — omitempty behavior, midnight-normalized due_at, the overdue flag —
// render exactly the chips the strip promises.
func TestKithContextSnapshotParsesToTheMappedContext(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/kith_context_snapshot.json")
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	client, _ := newContextClient(t, &contextStub{status: 200, body: string(raw)})

	got, err := client.ContactContext(t.Context(), NewContactRef("01M"))
	if err != nil {
		t.Fatalf("context: %v", err)
	}

	want := Context{
		Company: "Analytical Engines",
		Deals: []Deal{
			{Title: "Retainer 2027", Stage: "proposal", ValueCents: 1200000, Currency: "EUR"},
		},
		DealCount: 1,
		OpenTasks: 2,
		NextTask: &Task{
			Title: "Send the revised offer",
			DueAt: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC),
			// The producer computes overdue against ITS clock: the capture
			// ran after the task's due date, so the flag is true.
			Overdue: true,
		},
		LastActivity: time.Date(2026, 11, 11, 8, 31, 49, 0, time.UTC),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot mapped to %+v, want %+v", got, want)
	}
}

// producerContextKeys is the producer's DECLARED key universe, mirrored from
// contextResponse (crm@85c7f2e). omitempty keys may be absent from any given
// capture (this one carries no tags), but the fixture may not invent keys
// outside it.
var producerContextKeys = map[string]string{
	"id":                       "string",
	"name":                     "string",
	"email":                    "string",
	"company":                  "string",
	"tags":                     "array",
	"open_deals":               "array",
	"open_deal_count":          "number",
	"open_tasks":               "number",
	"next_task":                "object",
	"last_activity_at":         "string",
	"open_deals[].title":       "string",
	"open_deals[].stage":       "string",
	"open_deals[].value_cents": "number",
	"open_deals[].currency":    "string",
	"next_task.title":          "string",
	"next_task.due_at":         "string",
	"next_task.overdue":        "bool",
}

// TestConsumerFixtureCoversTheProducerShape keeps the hand-written fixture
// and the captured snapshot honest against each other: every key the
// producer actually emitted must exist in the fixture with the same JSON
// type, and every key the fixture carries must be inside the producer's
// declared universe. A producer rename therefore fails HERE, in the
// consumer's CI, at the next snapshot re-record — not on a customer's
// answered call.
func TestConsumerFixtureCoversTheProducerShape(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/kith_context_snapshot.json")
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	var snapshotAny, fixtureAny any
	if err := json.Unmarshal(raw, &snapshotAny); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if err := json.Unmarshal([]byte(fullContextJSON), &fixtureAny); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	snapshotShape := shapeOf(snapshotAny)
	fixtureShape := shapeOf(fixtureAny)

	for key, want := range snapshotShape {
		got, ok := fixtureShape[key]
		if !ok {
			t.Errorf("producer emits %q but the fixture never heard of it: re-record the snapshot and reconcile", key)

			continue
		}
		if got != want {
			t.Errorf("producer emits %q as %s, the fixture decodes it as %s", key, want, got)
		}
	}

	for key := range fixtureShape {
		if _, ok := producerContextKeys[key]; !ok {
			t.Errorf("fixture carries %q, which is outside the producer's declared key universe (crm contextResponse): invented fact", key)
		}
	}
}

// shapeOf flattens a decoded JSON value into path -> type: nested objects
// contribute "a.b" entries, array elements contribute "key[]" plus their
// element paths ("open_deals[].title").
func shapeOf(v any) map[string]string {
	out := map[string]string{}
	walkShape("", v, out)

	return out
}

func walkShape(path string, v any, out map[string]string) {
	switch val := v.(type) {
	case map[string]any:
		for k, sub := range val {
			p := k
			if path != "" {
				p = path + "." + k
			}
			out[p] = jsonTypeName(sub)
			walkShape(p, sub, out)
		}
	case []any:
		if len(val) == 0 {
			return
		}
		// The parent path already says "array"; only the element's nested
		// paths carry extra information, so no bare "key[]" entry.
		walkShape(path+"[]", val[0], out)
	}
}

func jsonTypeName(v any) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}
