package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// FuzzHookJSONDecode drives the webhook decode path (json/v2 plus the
// flexPages custom unmarshaller) with hostile bodies. The invariant: any
// input either decodes or answers a 400 — never a panic and never a
// stored/garbage state. Deep nesting and huge strings are classic
// decoder killers, hence the seeded corpus.
func FuzzHookJSONDecode(f *testing.F) {
	f.Add([]byte(`{"owner":"1001","from":"+441632960961","body":"hi"}`))
	f.Add([]byte(`{"provider_ref":"gw-1","status":"transmitted","pages":3}`))
	f.Add([]byte(`{"provider_ref":"gw-1","status":"transmitted","pages":"3"}`))
	f.Add([]byte(`{"pages":-1}`))
	f.Add([]byte(`{"pages":"many"}`))
	f.Add([]byte(`{"pages":null}`))
	f.Add([]byte(`{"attachments":[{"data_base64":"!!!"}]}`))
	f.Add([]byte(`{"a":` + strings.Repeat(`[`, 4096) + strings.Repeat(`]`, 4096) + `}`))
	f.Add([]byte(`{"body":"` + strings.Repeat("x", 1<<20) + `"}`))
	f.Add([]byte("\xff\xfe\x00garbage"))

	f.Fuzz(func(t *testing.T, data []byte) {
		r := httptest.NewRequest(http.MethodPost, "/hooks/message", bytes.NewReader(data))
		w := httptest.NewRecorder()

		var msg struct {
			Owner       string `json:"owner"`
			From        string `json:"from"`
			Body        string `json:"body"`
			Attachments []struct {
				Name     string `json:"name"`
				MimeType string `json:"mime_type"`
				DataB64  string `json:"data_base64"`
			} `json:"attachments"`
		}
		_ = decodeJSON(w, r, &msg)

		var fax struct {
			ProviderRef string `json:"provider_ref"`
			Status      string `json:"status"`
			faxStatusPages
			Error  string `json:"error"`
			PDFB64 string `json:"pdf_base64"`
		}
		r2 := httptest.NewRequest(http.MethodPost, "/hooks/fax/status", bytes.NewReader(data))
		w2 := httptest.NewRecorder()
		err := decodeJSON(w2, r2, &fax)
		if err != nil && w2.Code != http.StatusBadRequest {
			t.Fatalf("decode error answered %d, want 400: %v", w2.Code, err)
		}
		// Semantic validity (status whitelist, ref presence) lives in the
		// handlers and is covered there; the decode layer's only job is
		// to never panic and to answer 400 on malformed input.
	})
}
