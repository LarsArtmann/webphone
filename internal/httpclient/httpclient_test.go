package httpclient

import (
	"net/http"
	"testing"
	"time"
)

func TestIsolatedOwnsItsTransport(t *testing.T) {
	client := Isolated(7 * time.Second)
	if client.Timeout != 7*time.Second {
		t.Fatalf("Timeout = %v, want 7s", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T, want *http.Transport", client.Transport)
	}
	if transport == http.DefaultTransport {
		t.Fatal("Transport must be a clone, never the shared DefaultTransport")
	}
}
