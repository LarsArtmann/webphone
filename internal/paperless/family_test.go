// Adapter-level classification pins (family-adoption convention): the
// paperless seam creates NO errors of its own — every failure originates
// in the go-paperless SDK, which classifies at origin, and the adapter
// wraps family-neutrally. These pins prove the neutral wraps keep the
// SDK's classification reachable through the adapter's returns.
package paperless_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/paperless"
)

// familyStub answers every request with the status under test.
func familyStub(t *testing.T, status int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "stub", status)
	}))
	t.Cleanup(server.Close)
	return server
}

// TestAdapterPreservesSDKFamilies pins the seam's propagation contract:
// a rejected token stays a Rejection and a server hiccup stays a
// Transient after the adapter's neutral wraps — a fixed-family wrap here
// would clobber the SDK's origin classification.
func TestAdapterPreservesSDKFamilies(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		want   errorfamily.Family
	}{
		{"rejected token", http.StatusUnauthorized, errorfamily.Rejection},
		{"server error", http.StatusInternalServerError, errorfamily.Transient},
	} {
		t.Run(tc.name, func(t *testing.T) {
			archiver, err := paperless.NewArchiver(familyStub(t, tc.status).URL, "sekrit", nil)
			if err != nil {
				t.Fatal(err)
			}
			err = archiver.ArchiveFax(context.Background(), archiveJob(t), archivePDF)
			if err == nil {
				t.Fatal("stub status must fail the archive")
			}
			errorfamilytest.AssertFamily(t, err, tc.want)
		})
	}
}
