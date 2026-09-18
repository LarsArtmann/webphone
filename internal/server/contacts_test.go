package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestContactsVCardImportExport(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	vcf := "BEGIN:VCARD\r\n" +
		"VERSION:3.0\r\n" +
		"FN:Alice\r\n" +
		"TEL:+441632960961\r\n" +
		"END:VCARD\r\n" +
		"BEGIN:VCARD\r\n" +
		"FN:Bad Number\r\n" +
		"TEL:not-a-phone\r\n" +
		"END:VCARD\r\n" +
		"BEGIN:VCARD\r\n" +
		"FN:Bob\r\n" +
		"TEL:+493012345678\r\n" +
		"END:VCARD\r\n"
	form, contentType := multipartBody(t,
		map[string]string{},
		map[string]struct {
			Name    string
			Content []byte
		}{"vcard": {Name: "contacts.vcf", Content: []byte(vcf)}})
	resp, body := c.do(http.MethodPost, "/contacts/import", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("import: %d %.300s", resp.StatusCode, body)
	}
	panel := string(body)
	if !strings.Contains(panel, "Alice") || !strings.Contains(panel, "+441632960961") {
		t.Errorf("Alice missing after import: %.400s", panel)
	}
	if !strings.Contains(panel, "Bob") {
		t.Errorf("Bob missing after import: %.400s", panel)
	}
	if strings.Contains(panel, "Bad Number") {
		t.Errorf("invalid number must be skipped: %.400s", panel)
	}

	// Re-importing the same file renames instead of duplicating (the
	// store upserts by number).
	resp, _ = c.do(http.MethodPost, "/contacts/import", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("re-import: %d", resp.StatusCode)
	}
	_, panelBytes := c.do(http.MethodGet, "/partials/contacts", nil, "")
	// The row renders the number twice (text + data-dial attribute);
	// exactly one row means exactly two occurrences.
	if got := strings.Count(string(panelBytes), `data-dial="+441632960961"`); got != 1 {
		t.Errorf("duplicate import must not create a second row: %d rows", got)
	}

	// Export round-trips the stored contacts.
	resp, exported := c.do(http.MethodGet, "/contacts/export", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/vcard") {
		t.Errorf("export content type: %s", ct)
	}
	exportedText := string(exported)
	for _, want := range []string{"BEGIN:VCARD", "FN:Alice", "+441632960961", "FN:Bob", "END:VCARD"} {
		if !strings.Contains(exportedText, want) {
			t.Errorf("export missing %q: %s", want, exportedText)
		}
	}
	if strings.Contains(exportedText, "not-a-phone") {
		t.Error("invalid number must never reach the store or export")
	}

	// Anonymous export is gated.
	anon := clientFor(t, server)
	anon.token = ""
	resp, _ = anon.do(http.MethodGet, "/contacts/export", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous export: %d (want 401)", resp.StatusCode)
	}
}
