package server

import (
	"net/http"
	"testing"
)

// A malformed identifier in a path or query must answer 404, not panic the
// request: a signed-in client can put anything in a URL, and every panic
// is a stack trace in the operator's log for input nobody can fix. The
// Parse*/Must* split in internal/domain is the mechanism — this pins the
// handler behavior.
func TestMalformedIdentifiersAnswer404NotPanic(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"thread partial", http.MethodGet, "/partials/messages/not-a-thread-id"},
		{"mark read", http.MethodPost, "/messages/not-a-thread-id/read"},
		{"fax document", http.MethodGet, "/fax/not-a-fax-id/document"},
		{"attachment", http.MethodGet, "/attachments/not-an-attachment-id"},
		{"delete contact", http.MethodPost, "/contacts/delete?id=short"},
		{"branded id of wrong length", http.MethodGet, "/partials/messages/Thread:tooshort"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := c.do(tt.method, tt.path, nil, "")
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("%s %s = %d, want 404", tt.method, tt.path, resp.StatusCode)
			}
		})
	}
}
