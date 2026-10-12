package textagent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/go-error-family"
)

// TestFamilyPins locks the seam's error taxonomy: Rejection for a
// mis-deployed secret and for agent verdicts below 500 (the caller must
// surface them, not retry), Transient for agent-side 5xx and transport
// failures (load to retry behind).
func TestFamilyPins(t *testing.T) {
	_, missing := New("http://127.0.0.1:8071", filepath.Join(t.TempDir(), "absent"))
	errorfamilytest.AssertFamily(t, missing, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, missing, "textagent.secret")

	emptyPath := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(emptyPath, []byte("  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, emptyErr := New("http://127.0.0.1:8071", emptyPath)
	errorfamilytest.AssertFamily(t, emptyErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, emptyErr, "textagent.secret")

	verdict := &Error{Status: http.StatusConflict, Message: "already sent"}
	errorfamilytest.AssertFamily(t, verdict, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, verdict, "textagent.http")
	errorfamilytest.AssertFamily(t, fmt.Errorf("approve: %w", verdict), errorfamily.Rejection)

	overload := &Error{Status: http.StatusServiceUnavailable, Message: "fail-closed"}
	errorfamilytest.AssertFamily(t, overload, errorfamily.Transient)
	errorfamilytest.AssertCode(t, overload, "textagent.http")
}

// TestTransportFamilyPin drives the real transport path: a closed port
// classifies Transient with the reach code.
func TestTransportFamilyPin(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	secret := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secret, []byte("s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	client, err := New(dead.URL, secret)
	if err != nil {
		t.Fatal(err)
	}
	_, dialErr := client.Reviews(t.Context())
	if dialErr == nil {
		t.Fatal("unreachable agent must fail")
	}
	errorfamilytest.AssertFamily(t, dialErr, errorfamily.Transient)
	errorfamilytest.AssertCode(t, dialErr, "textagent.reach")
}

// TestBadIDFamilyPin: the draft-id guard answers a coded Rejection.
func TestBadIDFamilyPin(t *testing.T) {
	client := &Client{base: "http://127.0.0.1:1", secret: "x", http: http.DefaultClient}
	_, err := client.Approve(t.Context(), "../escape", "")
	if err == nil {
		t.Fatal("a path-escaping id must be rejected")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "textagent.badid")
}
