package server

import (
	"encoding/json/v2"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type apiContacts struct {
	Personal []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Number string `json:"number"`
	} `json:"personal"`
	Shared []struct {
		Name   string `json:"name"`
		Number string `json:"number"`
	} `json:"shared"`
}

func postJSON(t *testing.T, c *client, path string, payload any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	resp, respBody := c.do(http.MethodPost, path, body, "application/json")
	if resp.StatusCode >= 500 {
		t.Fatalf("POST %s: %d %s", path, resp.StatusCode, respBody)
	}
	return resp
}

func listContacts(t *testing.T, c *client) apiContacts {
	t.Helper()
	resp, body := c.do(http.MethodGet, "/api/contacts", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list contacts: %d %s", resp.StatusCode, body)
	}
	var got apiContacts
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("list contacts body %q: %v", body, err)
	}
	return got
}

func TestContactsAPIRequiresSession(t *testing.T) {
	server := newTestServer(t)
	anon := clientFor(t, server)
	anon.token = ""

	// GET hits the session gate directly; the state-changing verbs are
	// additionally behind CSRF — an anonymous caller must never succeed.
	resp, _ := anon.do(http.MethodGet, "/api/contacts", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous GET: %d (want 401)", resp.StatusCode)
	}
	for _, call := range []struct{ method, path string }{
		{http.MethodPost, "/api/contacts"},
		{http.MethodDelete, "/api/contacts?id=x"},
	} {
		resp, _ := anon.do(call.method, call.path, nil, "")
		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
			t.Errorf("anonymous %s %s: %d (want 401/403)", call.method, call.path, resp.StatusCode)
		}
	}
}

func TestContactsAPIRoundTrip(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	// Save answers 204 — the island re-fetches for the true id (the
	// store keeps the old id on a rename-upsert).
	resp := postJSON(t, c, "/api/contacts", map[string]string{"name": "Alice", "number": "+441632960961"})
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save: %d", resp.StatusCode)
	}

	got := listContacts(t, c)
	if len(got.Personal) != 1 {
		t.Fatalf("personal after save: %d rows (want 1): %+v", len(got.Personal), got)
	}
	saved := got.Personal[0]
	if saved.ID == "" || saved.Name != "Alice" || saved.Number != "+441632960961" {
		t.Errorf("saved row: %+v", saved)
	}
	if len(got.Shared) != 1 || got.Shared[0].Name != "Support" || got.Shared[0].Number != "2000" {
		t.Errorf("shared contacts in API response: %+v", got.Shared)
	}

	// Re-saving the same number renames (upsert), not duplicates — and
	// the id must stay stable, which is exactly why mutations answer 204
	// and the list is the only id source.
	resp = postJSON(t, c, "/api/contacts", map[string]string{"name": "Alice Renamed", "number": "+441632960961"})
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("rename: %d", resp.StatusCode)
	}
	got = listContacts(t, c)
	if len(got.Personal) != 1 || got.Personal[0].Name != "Alice Renamed" {
		t.Fatalf("upsert must rename in place: %+v", got.Personal)
	}
	if got.Personal[0].ID != saved.ID {
		t.Errorf("id drifted on rename: %q -> %q", saved.ID, got.Personal[0].ID)
	}

	// Validation mirrors the tab route: bad numbers 422, garbage 400.
	if resp := postJSON(t, c, "/api/contacts", map[string]string{"name": "Bad", "number": "not-a-phone"}); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("invalid number: %d (want 422)", resp.StatusCode)
	}
	resp, body := c.do(http.MethodPost, "/api/contacts", []byte("{"), "application/json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("malformed body: %d %.100s (want 400)", resp.StatusCode, body)
	}

	// Delete by the store's id, scoped to the owner.
	resp, _ = c.do(http.MethodDelete, "/api/contacts?id="+saved.ID, nil, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", resp.StatusCode)
	}
	if got := listContacts(t, c); len(got.Personal) != 0 {
		t.Errorf("personal after delete: %+v", got.Personal)
	}
	resp, _ = c.do(http.MethodDelete, "/api/contacts?id="+saved.ID, nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("double delete: %d (want 404)", resp.StatusCode)
	}
	resp, _ = c.do(http.MethodDelete, "/api/contacts?id=garbage", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("garbage id: %d (want 404)", resp.StatusCode)
	}
}

func TestContactsAPIExtensionIsolation(t *testing.T) {
	server := newTestServer(t)
	alice := signIn(t, server) // extension 1001
	bob := clientFor(t, server)
	bob.login("1002", "pw")

	if resp := postJSON(t, alice, "/api/contacts", map[string]string{"name": "Alice Contact", "number": "+441632960961"}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("alice save: %d", resp.StatusCode)
	}

	// Bob's list never shows Alice's rows.
	for _, contact := range listContacts(t, bob).Personal {
		if contact.Number == "+441632960961" {
			t.Errorf("bob sees alice's contact: %+v", contact)
		}
	}

	// Bob cannot delete Alice's row by its id.
	aliceID := listContacts(t, alice).Personal[0].ID
	resp, _ := bob.do(http.MethodDelete, "/api/contacts?id="+aliceID, nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("cross-extension delete: %d (want 404)", resp.StatusCode)
	}

	// Alice's row survives the attempt.
	if got := listContacts(t, alice); len(got.Personal) != 1 {
		t.Errorf("alice's contact must survive: %+v", got.Personal)
	}
}

// TestContactsAPIAndTabShareOneStore pins the single-home invariant end
// to end: a contact saved through the island's JSON API renders in the
// Contacts tab partial, and one saved through the tab answers the API.
func TestContactsAPIAndTabShareOneStore(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	postJSON(t, c, "/api/contacts", map[string]string{"name": "Via API", "number": "+441632960961"})
	_, body := c.do(http.MethodGet, "/partials/contacts", nil, "")
	if !strings.Contains(string(body), "Via API") {
		t.Errorf("tab partial missing the API-saved contact: %.400s", body)
	}

	form := url.Values{"name": {"Via Tab"}, "number": {"+493012345678"}}.Encode()
	resp, _ := c.do(http.MethodPost, "/contacts/save", []byte(form), "application/x-www-form-urlencoded")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("tab save: %d", resp.StatusCode)
	}
	numbers := map[string]bool{}
	for _, contact := range listContacts(t, c).Personal {
		numbers[contact.Number] = true
	}
	if !numbers["+441632960961"] || !numbers["+493012345678"] {
		t.Errorf("API list must see both homes' rows: %v", numbers)
	}
}
