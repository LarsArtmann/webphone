package crm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The /context product is consumed over ONE wire shape, and the producer is a
// different repository: a change to the path, the auth header or the query
// would break the strip silently (the phone renders nothing for a miss), so the
// REQUEST is pinned byte-shape here rather than left implicit in a stub that
// only happens to match (P2-2).
func TestContactContextRequestShapeIsPinned(t *testing.T) {
	t.Parallel()

	var gotMethod, gotPath, gotAuth, gotQuery, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"company":"Analytical Engines Ltd"}`))
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.URL, "sekrit", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ref := NewContactRef("01M")
	if _, err := client.ContactContext(context.Background(), ref); err != nil {
		t.Fatalf("ContactContext: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/contacts/01M/context" {
		t.Errorf("path = %q, want /api/contacts/01M/context", gotPath)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want none (the ref addresses the contact; nothing else may drift in)", gotQuery)
	}
	if gotAuth != "Bearer sekrit" {
		t.Errorf("Authorization = %q, want the bearer the config provides", gotAuth)
	}
	if gotAccept == "" {
		t.Error("the request must state an Accept header — a proxy that content-negotiates must not answer HTML")
	}
}

// TestContactContextFailureClassesStayDistinct is the degraded-answer half: an
// absent context is a STATE (the strip renders nothing), a rejected bearer is a
// DIFFERENT state (the operator's token is wrong), and anything else is an
// infrastructure error. Collapsing them would hide a revoked token behind
// "no context published".
func TestContactContextFailureClassesStayDistinct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   error
	}{
		{"absent", http.StatusNotFound, ErrNoContext},
		{"revoked", http.StatusUnauthorized, ErrUnauthorized},
		{"forbidden", http.StatusForbidden, ErrUnauthorized},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, _ := newContextClient(t, &contextStub{status: tc.status})
			ref := NewContactRef("01M")
			if _, err := client.ContactContext(context.Background(), ref); !errors.Is(err, tc.want) {
				t.Fatalf("status %d = %v, want %v", tc.status, err, tc.want)
			}
		})
	}
}
