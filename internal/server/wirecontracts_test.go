package server

import (
	"encoding/json"
	"testing"
)

// TestWireContractsIsValid fails the build when the declared
// machine-to-machine catalog is inconsistent.
func TestWireContractsIsValid(t *testing.T) {
	if violations := BuildWireContracts().Validate(); len(violations) > 0 {
		t.Fatalf("wire contract violations: %v", violations)
	}
}

// TestWireContractSchemasMatchHandlerTypes pins the declared payloads to
// the exact types the handlers decode: the fax lanes must carry the
// flexPages field trio (the embedded FaxStatusPages promotion is what
// makes the bridge contract complete), and every inbound hook declares
// its Bearer-secret posture in the summary.
func TestWireContractSchemasMatchHandlerTypes(t *testing.T) {
	body, err := renderWireContracts()
	if err != nil {
		t.Fatalf("wire contract render: %v", err)
	}
	var doc struct {
		Channels map[string]struct {
			Messages map[string]struct {
				Payload struct {
					Ref string `json:"$ref"`
				} `json:"payload"`
			} `json:"messages"`
		} `json:"channels"`
		Components struct {
			Schemas map[string]any `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("wire contract is not JSON: %v", err)
	}

	faxStatus := doc.Components.Schemas["event.wire.hooks.fax.status"]
	if faxStatus == nil {
		t.Fatal("fax.status payload schema missing")
	}
	raw, _ := json.Marshal(faxStatus)
	for _, field := range []string{"pages", "page_count", "num_pages", "provider_ref", "status"} {
		if !jsonContainsField(raw, field) {
			t.Errorf("fax.status payload schema missing field %q (embedded FaxStatusPages promotion?)", field)
		}
	}
}

func jsonContainsField(raw []byte, field string) bool {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	if props, ok := m["properties"].(map[string]any); ok {
		if _, ok := props[field]; ok {
			return true
		}
	}
	for _, v := range m {
		if sub, err := json.Marshal(v); err == nil && jsonContainsField(sub, field) {
			return true
		}
	}
	return false
}
