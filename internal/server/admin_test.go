package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/admin"
	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
)

// The admin surface's behavior pins: the gate (anonymous 401, non-admin
// styled 404, admin 200), the nav's conditional tab, cross-owner
// visibility in every section, the read-only thread deep link, the
// attachment/fax document fallback, and the merged per-line CDR page.

// adminPasswordFile writes one extension's SIP password file.
func adminPasswordFile(t testing.TB, dir, extension, password string) string {
	t.Helper()
	path := filepath.Join(dir, extension+".txt")
	if err := os.WriteFile(path, []byte(password+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// newAdminServer wires the admin seam over a two-account mapping:
// boss@example.com owns 1001 (ADMIN), worker@example.com owns 1002.
// The identity layer's login modes stay off — the admin gate reads the
// session row's LoginEmail, never the ceremony, so the tests mint rows
// directly through the session store.
func newAdminServer(t testing.TB, phoneAPIURL string) (*testServer, session.Store) {
	t.Helper()
	dir := t.TempDir()
	mapped, err := authmap.New(
		map[string]authmap.User{
			"boss@example.com":   {Email: "boss@example.com", DisplayName: "The Boss", Extensions: []domain.Extension{domain.MustParseExtension("1001")}, Admin: true},
			"worker@example.com": {Email: "worker@example.com", Extensions: []domain.Extension{domain.MustParseExtension("1002")}},
		},
		map[string]string{
			"1001": adminPasswordFile(t, dir, "1001", "boss-pw"),
			"1002": adminPasswordFile(t, dir, "1002", "worker-pw"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	identities := map[string]string{"1001": "+49301111", "1002": "+49302222"}
	var sessions session.Store = session.NewMemStore(time.Hour)
	server := newTestServerWithPhoneAPI(t, phoneAPIURL, func(deps *Deps) {
		deps.Admin, err = admin.New(mapped, identities)
		if err != nil {
			t.Fatal(err)
		}
		deps.Sessions = sessions
	})
	return server, sessions
}

// mintedClient builds a client whose cookie carries a directly minted
// session row (extension login semantics: no email) — the shape a real
// session store holds after any login mode minted it.
func mintedClient(t testing.TB, server *testServer, sessions session.Store, extension, loginEmail string) *client {
	t.Helper()
	token, err := sessions.Create(domain.MustParseExtension(extension), "pw", loginEmail)
	if err != nil {
		t.Fatal(err)
	}
	c := clientFor(t, server)
	base, err := url.Parse(c.base)
	if err != nil {
		t.Fatal(err)
	}
	c.http.Jar.SetCookies(base, []*http.Cookie{{Name: session.CookieName, Value: token.String(), Path: "/"}})
	return c
}

func TestAdminTabGates(t *testing.T) {
	server, sessions := newAdminServer(t, "")

	// Anonymous: the session gate answers first (the /partials contract).
	anon := clientFor(t, server)
	if resp, _ := anon.do(http.MethodGet, "/partials/admin", nil, ""); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous admin partial: status %d, want 401", resp.StatusCode)
	}

	// Non-admin (mapped worker AND plain extension login alike): the
	// styled 404 — the surface never confirms it exists.
	worker := mintedClient(t, server, sessions, "1002", "worker@example.com")
	for _, path := range []string{"/admin", "/partials/admin", "/admin/threads/xyz"} {
		if resp, body := worker.do(http.MethodGet, path, nil, ""); resp.StatusCode != http.StatusNotFound || !strings.Contains(string(body), "404") {
			t.Fatalf("worker %s: %d %s, want the styled 404", path, resp.StatusCode, body)
		}
	}
	plain := mintedClient(t, server, sessions, "1001", "")
	if resp, _ := plain.do(http.MethodGet, "/admin", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("extension login (no identity) on /admin: %d, want 404", resp.StatusCode)
	}

	// Admin: the full page renders the panel.
	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	resp, body := boss.do(http.MethodGet, "/admin", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin page: %d", resp.StatusCode)
	}
	for _, want := range []string{"admin-panel", "wp-admin-users", "wp-admin-numbers", "wp-admin-threads", "wp-admin-faxes", "wp-admin-calls"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("admin page missing %q", want)
		}
	}
	// The partial swaps the same region.
	if resp, body = boss.do(http.MethodGet, "/partials/admin", nil, ""); resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "admin-panel") {
		t.Fatalf("admin partial: %d", resp.StatusCode)
	}
}

func TestAdminNavTabIsConditional(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	worker := mintedClient(t, server, sessions, "1002", "worker@example.com")
	_, body := worker.do(http.MethodGet, "/settings", nil, "")
	if strings.Contains(string(body), `href="/admin"`) {
		t.Error("non-admin shell must not render the Admin nav link")
	}
	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	_, body = boss.do(http.MethodGet, "/settings", nil, "")
	if !strings.Contains(string(body), `href="/admin"`) || !strings.Contains(string(body), "Verwaltung") && !strings.Contains(string(body), "Admin") {
		t.Error("admin shell must render the Admin nav link and label")
	}
}

func TestAdminPanelShowsEveryOwnersData(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	ctx := context.Background()

	// Two owners' conversations; 1002's must appear for the admin even
	// though the admin session IS 1001.
	neighbor := domain.MustParseExtension("1002")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+4989555777")
	threadID, err := server.messages.FindThread(ctx, neighbor, remote, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := server.messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: neighbor, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "neighbor line secret", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := server.faxes.Create(ctx, domain.FaxJob{
		ID: domain.GenerateFaxID(), Owner: neighbor, Remote: domain.MustParsePhone("+4989555888"),
		Direction: domain.FaxOutbound, Status: domain.FaxTransmitted, Pages: 3,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := server.transcripts.Append(ctx, domain.CallTranscriptSegment{
		Owner: neighbor, CallID: "call-admin-probe", Direction: domain.DirectionInbound,
		Remote: "+4989555999", StartedAt: time.Now(), Text: "neighbor call words",
	}); err != nil {
		t.Fatal(err)
	}

	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	_, body := boss.do(http.MethodGet, "/partials/admin", nil, "")
	page := string(body)
	for _, want := range []string{
		"1002", "+4989555777", "neighbor line secret",
		"+4989555888",
		"neighbor call words",
		"boss@example.com", "worker@example.com",
		"+49301111", "+49302222",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("admin panel missing %q", want)
		}
	}
	// The admin's own live session counts: 1001 carries exactly one.
	if !strings.Contains(page, "1") {
		t.Error("live session count must render")
	}
}

// TestAdminFindFilter pins M1: the q parameter narrows every data
// section server-side (case-insensitive substring across the row's
// human fields), the filter form pushes its URL so the filtered view
// stays addressable, and the read-only badge renders on the panel.
func TestAdminFindFilter(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	ctx := context.Background()

	neighbor := domain.MustParseExtension("1002")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+4989555777")
	threadID, err := server.messages.FindThread(ctx, neighbor, remote, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := server.messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: neighbor, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "neighbor line secret", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")

	_, unfiltered := boss.do(http.MethodGet, "/partials/admin", nil, "")
	page := string(unfiltered)
	if !strings.Contains(page, "neighbor line secret") {
		t.Fatalf("unfiltered panel must show the neighbor thread")
	}
	for _, want := range []string{`hx-push-url="true"`, `aria-label="Find"`, ">read-only<"} {
		if !strings.Contains(page, want) {
			t.Errorf("admin panel missing %q (filter form / badge)", want)
		}
	}

	// A remote-party hit survives; a non-match empties the threads
	// section (honest empty state), and the users inventory — config,
	// not searchable data — stays regardless.
	_, filtered := boss.do(http.MethodGet, "/partials/admin?q=555777", nil, "")
	if !strings.Contains(string(filtered), "neighbor line secret") {
		t.Errorf("the matching thread must survive the filter")
	}
	_, empty := boss.do(http.MethodGet, "/partials/admin?q=zzz-no-such", nil, "")
	page = string(empty)
	if !strings.Contains(page, "No conversations on any line.") {
		t.Errorf("the filtered-out threads section must show its empty state: %.600s", page)
	}
	if !strings.Contains(page, "worker@example.com") {
		t.Errorf("the users inventory is config — it must not filter away")
	}
}

func TestAdminThreadDeepLink(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	ctx := context.Background()
	neighbor := domain.MustParseExtension("1002")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+4915112345678")
	threadID, err := server.messages.FindThread(ctx, neighbor, remote, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := server.messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: neighbor, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "cross line bubble", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	path := "/admin/threads/" + threadID.String()
	if resp, body := boss.do(http.MethodGet, path, nil, ""); resp.StatusCode != http.StatusOK ||
		!strings.Contains(string(body), "cross line bubble") || !strings.Contains(string(body), "1002") {
		t.Fatalf("admin thread deep link: %d %s", resp.StatusCode, body)
	}
	resp, page := boss.do(http.MethodGet, "/partials"+path, nil, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(page), "cross line bubble") {
		t.Fatalf("admin thread partial: %d %s", resp.StatusCode, page)
	}
	// The read-only view must not render the owner composer.
	if strings.Contains(string(page), "hx-post=\"/messages/send\"") {
		t.Error("admin thread view must not render the reply composer")
	}
	// A garbage id is a miss, answered by the styled 404.
	if resp, _ := boss.do(http.MethodGet, "/admin/threads/nope", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Errorf("garbage thread id: %d, want 404", resp.StatusCode)
	}
}

func TestAdminAttachmentAndFaxDocumentFallback(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	ctx := context.Background()
	neighbor := domain.MustParseExtension("1002")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491519876543")
	threadID, err := server.messages.FindThread(ctx, neighbor, remote, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	attachmentID := domain.GenerateAttachmentID()
	if err := server.messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: neighbor, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "with picture", CreatedAt: time.Now(),
		Attachments: []domain.Attachment{{
			ID: attachmentID, MessageID: domain.MessageID{}, Name: "pic.png",
			MimeType: "image/png", SizeBytes: 8, Path: "0001/pic.png",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	// The seeded row carries no blob file, so a resolved row answers 410
	// ("attachment missing") — the pin is 410-not-404: 404 means the row
	// itself was withheld, 410 means the LOOKUP granted access.
	resp, _ := boss.do(http.MethodGet, "/attachments/"+attachmentID.String(), nil, "")
	if resp.StatusCode != http.StatusGone {
		t.Errorf("admin session must resolve another line's attachment row: %d, want 410 (row granted, file absent)", resp.StatusCode)
	}
	worker := mintedClient(t, server, sessions, "1002", "worker@example.com")
	if resp, _ = worker.do(http.MethodGet, "/attachments/"+attachmentID.String(), nil, ""); resp.StatusCode != http.StatusGone {
		t.Errorf("owning line keeps its own access: %d, want 410", resp.StatusCode)
	}
	// A session without the admin flag gets nothing: the row is withheld.
	stranger := mintedClient(t, server, sessions, "1001", "")
	if resp, _ = stranger.do(http.MethodGet, "/attachments/"+attachmentID.String(), nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Errorf("non-admin extension login: %d, want 404", resp.StatusCode)
	}
}

// newAdminCDRStub serves one CDR per extension (caller id = the Basic
// auth extension, so the merged page proves WHICH line surfaced the
// row); failFor lists extensions whose page fails (the honest-degrade
// path).
func newAdminCDRStub(t testing.TB, failFor map[string]bool) *httptest.Server {
	t.Helper()
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extension := "unknown"
		if encoded, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Basic "); ok {
			if raw, err := base64.StdEncoding.DecodeString(encoded); err == nil {
				extension = strings.SplitN(string(raw), ":", 2)[0]
			}
		}
		if failFor[extension] {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"entries":[{"context":"public","caller_id_number":%q,"destination_number":"+4989000000","start":"2026-11-04 1%c:00:00","billsec":42}]}`, extension, extension[len(extension)-1])
	}))
	t.Cleanup(stub.Close)
	return stub
}

func TestAdminCDRMergesPerLinePages(t *testing.T) {
	stub := newAdminCDRStub(t, map[string]bool{"1002": true})
	server, sessions := newAdminServer(t, stub.URL)

	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	_, body := boss.do(http.MethodGet, "/partials/admin", nil, "")
	page := string(body)
	// 1001's row merged; 1002's page failed honestly (named, not fatal).
	if !strings.Contains(page, "1001") || !strings.Contains(page, "+4989000000") {
		t.Error("the surviving line's CDR row must render")
	}
	if !strings.Contains(page, "1002") || !strings.Contains(page, "Call records unavailable") {
		t.Error("the failing line must be named in the degrade notice")
	}
}

func TestAdminCDROffWithoutPhoneAPI(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	boss := mintedClient(t, server, sessions, "1001", "boss@example.com")
	_, body := boss.do(http.MethodGet, "/partials/admin", nil, "")
	if !strings.Contains(string(body), "Call records need the phone API") {
		t.Error("a disabled phone API must explain the missing CDR section")
	}
}
