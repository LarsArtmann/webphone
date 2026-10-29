package crm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// contextStub is a Ledger machine-API double for the context product: it
// answers one path, counts upstream hits, and lets each test choose the status
// and body so absence, revocation and failure can be told apart.
type contextStub struct {
	hits   atomic.Int64
	status int
	body   string
}

func (s *contextStub) handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer sekrit" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	if r.Method != http.MethodGet || r.URL.Path != "/api/contacts/01M/context" {
		http.NotFound(w, r)

		return
	}

	s.hits.Add(1)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s.status)
	_, _ = w.Write([]byte(s.body))
}

func newContextClient(t *testing.T, stub *contextStub) (*Client, *Resolver) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(stub.handler))
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.URL, "sekrit")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client, NewResolver(client, "", nil)
}

// fullContextJSON is the producer's published shape, extra identity fields
// included. The phone ignores id/name/tags on purpose (identity arrives from
// the lookup product), so this fixture doubles as the forward-compatibility
// proof: unknown fields decode without a complaint.
const fullContextJSON = `{
  "id": "01M",
  "name": "Ada Lovelace",
  "company": "Analytical Engines Ltd",
  "tags": ["lead", "conference"],
  "open_deals": [
    {"title": "Retainer 2027", "stage": "proposal", "value_cents": 1200000, "currency": "EUR"}
  ],
  "open_deal_count": 4,
  "open_tasks": 2,
  "next_task": {"title": "Send the revised offer", "due_at": "2026-11-02T09:00:00Z", "overdue": false},
  "last_activity_at": "2026-10-24T11:03:00Z"
}`

// TestContactContextDecodesTheProducersShape pins the wire contract the CRM
// publishes: every fact the strip renders, decoded from the producer's exact
// field names. A field rename on the producer side must break THIS test, not a
// customer's page.
func TestContactContextDecodesTheProducersShape(t *testing.T) {
	t.Parallel()

	client, _ := newContextClient(t, &contextStub{status: http.StatusOK, body: fullContextJSON})

	got, err := client.ContactContext(context.Background(), NewContactRef("01M"))
	if err != nil {
		t.Fatalf("context: %v", err)
	}

	if got.Company != "Analytical Engines Ltd" {
		t.Errorf("company = %q", got.Company)
	}

	if len(got.Deals) != 1 || got.Deals[0].Title != "Retainer 2027" ||
		got.Deals[0].Stage != "proposal" || got.Deals[0].ValueCents != 1200000 ||
		got.Deals[0].Currency != "EUR" {
		t.Errorf("deals: %+v", got.Deals)
	}

	if got.DealCount != 4 {
		t.Errorf("deal count = %d, want the producer's TRUE total (4)", got.DealCount)
	}

	if got.OpenTasks != 2 || got.NextTask == nil || got.NextTask.Title != "Send the revised offer" {
		t.Errorf("tasks: %d, next %+v", got.OpenTasks, got.NextTask)
	}

	if got.NextTask.DueAt.UTC().Format("2006-01-02T15:04:05Z") != "2026-11-02T09:00:00Z" {
		t.Errorf("next task due = %s", got.NextTask.DueAt)
	}

	if got.NextTask.Overdue {
		t.Errorf("a future task must not be overdue")
	}

	if got.LastActivity.IsZero() {
		t.Errorf("last activity not parsed")
	}
}

// TestContactContextAbsenceIsNotAnError: a 404 means "no context for this
// contact / not published". The phone renders nothing, so it must be its own
// sentinel rather than a transient failure the caller might retry or toast.
func TestContactContextAbsenceIsNotAnError(t *testing.T) {
	t.Parallel()

	client, _ := newContextClient(t, &contextStub{status: http.StatusNotFound})

	_, err := client.ContactContext(context.Background(), NewContactRef("01M"))
	if !errors.Is(err, ErrNoContext) {
		t.Fatalf("404 = %v, want ErrNoContext", err)
	}
}

// TestContactContextRejectsARevokedToken keeps the half-configuration loud:
// a rejected bearer is named, not folded into "no context".
func TestContactContextRejectsARevokedToken(t *testing.T) {
	t.Parallel()

	client, _ := newContextClient(t, &contextStub{status: http.StatusUnauthorized})

	if _, err := client.ContactContext(context.Background(), NewContactRef("01M")); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("401 = %v, want ErrUnauthorized", err)
	}
}

// TestContactContextToleratesAMinimalAnswer: a CRM that publishes less (no
// company, no deals, no tasks, no timestamps) must still decode. The strip
// then renders fewer items instead of failing the whole surface.
func TestContactContextToleratesAMinimalAnswer(t *testing.T) {
	t.Parallel()

	client, _ := newContextClient(t, &contextStub{status: http.StatusOK, body: `{"id":"01M","name":"Ada"}`})

	got, err := client.ContactContext(context.Background(), NewContactRef("01M"))
	if err != nil {
		t.Fatalf("minimal context: %v", err)
	}

	if got.Company != "" || len(got.Deals) != 0 || got.NextTask != nil {
		t.Fatalf("minimal answer: %+v", got)
	}

	if !got.LastActivity.IsZero() {
		t.Errorf("absent last activity must stay zero, got %s", got.LastActivity)
	}
}

// TestResolverContextOfMemoizesAndCachesAbsence is the cost contract: an open
// conversation renders the strip once per page view, so repeated reads inside
// the window must cost ONE upstream round trip — and a CRM that publishes no
// context must be asked once per window, not once per page.
func TestResolverContextOfMemoizesAndCachesAbsence(t *testing.T) {
	t.Parallel()

	stub := &contextStub{status: http.StatusOK, body: fullContextJSON}
	_, resolver := newContextClient(t, stub)

	ref := NewContactRef("01M")
	for range 3 {
		if _, ok := resolver.ContextOf(context.Background(), ref); !ok {
			t.Fatalf("context must resolve")
		}
	}

	if hits := stub.hits.Load(); hits != 1 {
		t.Fatalf("upstream hit %d times for 3 reads (the memo failed)", hits)
	}

	absent := &contextStub{status: http.StatusNotFound}
	_, absentResolver := newContextClient(t, absent)

	for range 3 {
		if _, ok := absentResolver.ContextOf(context.Background(), ref); ok {
			t.Fatalf("a 404 must resolve to no context")
		}
	}

	if hits := absent.hits.Load(); hits != 1 {
		t.Fatalf("absent context hit upstream %d times for 3 reads (absence must be cached)", hits)
	}
}

// TestResolverContextOfDoesNotCacheTransportFailures: a flaky CRM must be
// retried on the next open. Caching a timeout would blind the strip for
// minutes after a single blip.
func TestResolverContextOfDoesNotCacheTransportFailures(t *testing.T) {
	t.Parallel()

	stub := &contextStub{status: http.StatusInternalServerError}
	_, resolver := newContextClient(t, stub)

	ref := NewContactRef("01M")
	for range 2 {
		if _, ok := resolver.ContextOf(context.Background(), ref); ok {
			t.Fatalf("a 500 must resolve to no context")
		}
	}

	if hits := stub.hits.Load(); hits != 2 {
		t.Fatalf("upstream hit %d times for 2 reads (a failure must NOT be cached)", hits)
	}
}

// TestResolverContextOfIsDisabledSafe: with no CRM configured the strip
// neither renders nor dials out. A nil resolver is the test-composition case.
func TestResolverContextOfIsDisabledSafe(t *testing.T) {
	t.Parallel()

	var nilResolver *Resolver
	if _, ok := nilResolver.ContextOf(context.Background(), NewContactRef("01M")); ok {
		t.Fatal("a nil resolver must never resolve a context")
	}

	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, ok := NewResolver(client, "", nil).ContextOf(context.Background(), NewContactRef("01M")); ok {
		t.Fatal("a disabled resolver must never resolve a context")
	}

	if _, ok := NewResolver(client, "", nil).ContextOf(context.Background(), ContactRef{}); ok {
		t.Fatal("a zero ref must never resolve a context")
	}
}
