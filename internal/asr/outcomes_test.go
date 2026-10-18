package asr

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOutcomeSnapshotCountsAndLatency(t *testing.T) {
	var o outcomes
	o.record(10*time.Millisecond, nil)
	o.record(30*time.Millisecond, nil)
	o.record(5*time.Millisecond, errors.New("provider answered 500"))

	s := o.snapshot()
	if s.Calls != 3 || s.Failures != 1 {
		t.Fatalf("counts: calls=%d failures=%d, want 3/1", s.Calls, s.Failures)
	}
	if s.AvgLatency != 15*time.Millisecond {
		t.Errorf("avg latency %s, want 15ms (the rolling mean)", s.AvgLatency)
	}
	if s.LastError == "" || !strings.Contains(s.LastError, "500") {
		t.Errorf("last error %q, want the recorded cause", s.LastError)
	}
	if s.LastSuccess.IsZero() {
		t.Error("last success must be stamped")
	}
}

func TestOutcomesHealthCheckVerdicts(t *testing.T) {
	var o outcomes
	if err := o.HealthCheck(context.Background()); err != nil {
		t.Errorf("no traffic yet must be healthy, got %v", err)
	}

	var degraded outcomes
	degraded.record(time.Millisecond, errors.New("connection refused"))
	if err := degraded.HealthCheck(context.Background()); err == nil {
		t.Fatal("failures without a later success must degrade")
	} else if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("degraded verdict must name the cause, got %v", err)
	}

	var recovered outcomes
	recovered.record(time.Millisecond, errors.New("connection refused"))
	time.Sleep(2 * time.Millisecond)
	recovered.record(time.Millisecond, nil)
	if err := recovered.HealthCheck(context.Background()); err != nil {
		t.Errorf("a success after the last failure must recover, got %v", err)
	}
}

// TestClientFeedsOutcomes pins the latency metric wiring: real round
// trips through the OpenAI-compatible client land in the snapshot both
// providers expose.
func TestClientFeedsOutcomes(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls <= 2 {
			_, _ = w.Write([]byte(`{"text":"hi"}`))
			return
		}
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.Transcribe(ctx, Request{Audio: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Transcribe(ctx, Request{Audio: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Transcribe(ctx, Request{Audio: []byte("x")}); err == nil {
		t.Fatal("third call must fail (provider 502)")
	}

	s := client.RecentOutcomes()
	if s.Calls != 3 || s.Failures != 1 {
		t.Fatalf("snapshot calls=%d failures=%d, want 3/1", s.Calls, s.Failures)
	}
	if s.AvgLatency <= 0 {
		t.Errorf("avg latency %s, want a positive round-trip mean", s.AvgLatency)
	}
	if err := client.HealthCheck(ctx); err == nil {
		t.Error("failing-since-last-success must degrade the health leg")
	}
}

// TestOutcomeSnapshotConcurrentRecords keeps the tracker race-clean:
// many goroutines recording at once must never tear the ring.
func TestOutcomeSnapshotConcurrentRecords(t *testing.T) {
	var o outcomes
	var wg sync.WaitGroup
	for i := range 64 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n%3 == 0 {
				o.record(time.Millisecond, errors.New("x"))
				return
			}
			o.record(time.Millisecond, nil)
		}(i)
	}
	wg.Wait()
	s := o.snapshot()
	if s.Calls != 64 {
		t.Fatalf("calls %d, want 64 (no lost records)", s.Calls)
	}
}
