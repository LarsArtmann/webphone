package crm

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

// stubCRM is a configurable Ledger machine-API double.
type stubCRM struct {
	lookupBodies atomic.Int64
	lookupStatus int
	lookupJSON   string
	logStatus    int
	logCalls     []map[string]any
}

func (s *stubCRM) handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer sekrit" {
		w.Header().Set("WWW-Authenticate", `Bearer realm="ledger-api"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/contacts/by-phone":
		s.lookupBodies.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(s.lookupStatus)
		_, _ = w.Write([]byte(s.lookupJSON)) //nolint:erraudit // test stub write
	case r.Method == http.MethodPost && r.URL.Path == "/api/contacts/01M/contact/calls":
		var got map[string]any
		_ = json.UnmarshalRead(r.Body, &got) //nolint:errcheck // test stub decode
		s.logCalls = append(s.logCalls, got)
		w.WriteHeader(s.logStatus)
	default:
		http.NotFound(w, r)
	}
}

func newStubServer(t *testing.T, stub *stubCRM) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(stub.handler))
	t.Cleanup(srv.Close)
	return srv
}

func TestClientDisabledByDefault(t *testing.T) {
	client, err := NewClient("", "token")
	if err != nil {
		t.Fatalf("NewClient empty: %v", err)
	}
	if client.Enabled() {
		t.Fatal("empty base URL must mean disabled")
	}

	resolver := NewResolver(client, nil)
	if resolver.Enabled() {
		t.Fatal("resolver over a disabled client must be disabled")
	}
	if _, ok := resolver.Resolve(context.Background(), "+493012345678"); ok {
		t.Fatal("disabled resolver must never match")
	}
	if err := resolver.LogCall(context.Background(), "x", "in", "+493012345678", 1, "answered"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("LogCall on disabled client: %v", err)
	}

	// A nil *Client (a Deps field nobody wired) must behave exactly like
	// a zero client: disabled, never a panic on the base-URL deref.
	var nilClient *Client
	if nilClient.Enabled() {
		t.Fatal("nil client must report disabled")
	}
	if _, err := nilClient.LookupByPhone(context.Background(), "+493012345678"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("LookupByPhone on nil client: %v", err)
	}
	if err := nilClient.LogCall(context.Background(), "x", "in", "+493012345678", 1, "answered"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("LogCall on nil client: %v", err)
	}
}

func TestLookupByPhone(t *testing.T) {
	stub := &stubCRM{
		lookupStatus: http.StatusOK,
		lookupJSON:   `{"results":[{"id":"01M","first_name":"Ada","last_name":"Lovelace","email":"ada@x","phones":["+49 30 12345678"]}]}`,
		logStatus:    http.StatusNoContent,
	}
	srv := newStubServer(t, stub)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	match, err := client.LookupByPhone(context.Background(), "+49 30 12345678")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if match.ID != "01M" || match.Name != "Ada Lovelace" {
		t.Fatalf("match: %+v", match)
	}

	// The stub answers every number with the same document; an empty
	// result list is the miss case (the CRM owns number matching).
	stub.lookupJSON = `{"results":[]}`
	if _, err := client.LookupByPhone(context.Background(), "+449900000000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("miss: want ErrNotFound, got %v", err)
	}

	bad, err := NewClient(srv.URL, "wrong")
	if err != nil {
		t.Fatalf("NewClient wrong: %v", err)
	}
	if _, err := bad.LookupByPhone(context.Background(), "+493012345678"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("bad token: want ErrUnauthorized, got %v", err)
	}
}

func TestLogCall(t *testing.T) {
	stub := &stubCRM{logStatus: http.StatusNoContent}
	srv := newStubServer(t, stub)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if err := client.LogCall(context.Background(), "01M/contact", "in", "+49 30 12345678", 123, "answered"); err != nil {
		t.Fatalf("log call: %v", err)
	}

	if len(stub.logCalls) != 1 {
		t.Fatalf("upstream saw %d calls", len(stub.logCalls))
	}

	got := stub.logCalls[0]
	if got["direction"] != "in" || got["seconds"] != float64(123) || got["outcome"] != "answered" {
		t.Fatalf("forwarded body: %+v", got)
	}

	stub.logStatus = http.StatusInternalServerError
	if err := client.LogCall(context.Background(), "01M/contact", "in", "+49 30 12345678", 0, ""); err == nil {
		t.Fatal("5xx must surface as an error")
	}
}

func TestResolverCachesLookups(t *testing.T) {
	stub := &stubCRM{
		lookupStatus: http.StatusOK,
		lookupJSON:   `{"results":[{"id":"01M","first_name":"Ada"}]}`,
	}
	srv := newStubServer(t, stub)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resolver := NewResolver(client, nil)
	ctx := context.Background()

	for range 3 {
		if got := resolver.Name(ctx, "+49 30 12345678"); got != "Ada" {
			t.Fatalf("name: %q", got)
		}
	}

	if hits := stub.lookupBodies.Load(); hits != 1 {
		t.Fatalf("upstream hit %d times for 3 lookups (cache failed)", hits)
	}

	// Negative caching: a miss is remembered too.
	stub.lookupJSON = `{"results":[]}`
	if _, ok := resolver.Resolve(ctx, "+449900000000"); ok {
		t.Fatal("unexpected match for unknown number")
	}
	stub.lookupJSON = `{"results":[{"id":"late","first_name":"Late"}]}`
	if _, ok := resolver.Resolve(ctx, "+449900000000"); ok {
		t.Fatal("negative cache entry must hold until its TTL expires")
	}
}

func TestResolverDegradesQuietly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resolver := NewResolver(client, nil)
	if name := resolver.Name(context.Background(), "+493012345678"); name != "" {
		t.Fatalf("CRM failure must degrade to empty name, got %q", name)
	}
}

// A burst of concurrent Resolves for one number (history and tab renders
// racing) must cost the CRM ONE upstream lookup: the first caller leads
// the round-trip, the rest join its flight. Also pins the upstream
// outcome counters (2026-09-22 single-flight train).
func TestResolveSingleFlightsConcurrentLookups(t *testing.T) {
	stub := &stubCRM{
		lookupStatus: http.StatusOK,
		lookupJSON:   `{"results":[{"id":"01M/contact","first_name":"Ada"}]}`,
	}
	srv := newStubServer(t, stub)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resolver := NewResolver(client, nil)

	const racers = 10
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range racers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, ok := resolver.Resolve(context.Background(), "+493012345678"); !ok {
				t.Error("a joined Resolve must inherit the leader's match")
			}
		}()
	}
	close(start)
	wg.Wait()

	if got := stub.lookupBodies.Load(); got != 1 {
		t.Fatalf("concurrent Resolve for one number hit the CRM %d times (want 1)", got)
	}
	hit, miss, failure := resolver.LookupCounters()
	if hit != 1 || miss != 0 || failure != 0 {
		t.Fatalf("counters after one upstream hit: hit=%d miss=%d failure=%d", hit, miss, failure)
	}
}

func TestResolverLookupCounters(t *testing.T) {
	stub := &stubCRM{
		lookupStatus: http.StatusOK,
		lookupJSON:   `{"results":[{"id":"01M/contact","first_name":"Ada"}]}`,
	}
	srv := newStubServer(t, stub)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resolver := NewResolver(client, nil)
	ctx := context.Background()

	if _, ok := resolver.Resolve(ctx, "+493011111111"); !ok {
		t.Fatal("known number must match")
	}
	stub.lookupJSON = `{"results":[]}`
	if _, ok := resolver.Resolve(ctx, "+493022222222"); ok {
		t.Fatal("unknown number must miss")
	}

	deadClient, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	srv.Close() // lookups via deadClient now fail at the transport
	deadResolver := NewResolver(deadClient, nil)
	if _, ok := deadResolver.Resolve(ctx, "+493033333333"); ok {
		t.Fatal("dead CRM must not match")
	}

	hit, miss, failure := resolver.LookupCounters()
	if hit != 1 || miss != 1 || failure != 1 {
		t.Fatalf("counters after hit+miss+failure: hit=%d miss=%d failure=%d", hit, miss, failure)
	}

	// A cache hit is not CRM traffic: the counters must not move.
	if _, ok := resolver.Resolve(ctx, "+493011111111"); !ok {
		t.Fatal("cached number must still match")
	}
	if hit, _, _ := resolver.LookupCounters(); hit != 1 {
		t.Fatalf("cache hit counted as an upstream lookup (%d)", hit)
	}

	if _, miss, failure := deadResolver.LookupCounters(); miss != 0 || failure != 0 {
		t.Fatalf("counters must be per-resolver, got miss=%d failure=%d", miss, failure)
	}
}
