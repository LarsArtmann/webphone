// CRM family pins (family-adoption train, 2026-09-30): sentinels
// classify via registration and keep identity (errIsMiss compares by
// equality); transport/decode are Transient; the status arms split
// Rejection/Transient mirroring the pbx and gateway seams. The
// resolver's negative-cache path (errIsMiss) must stay family-blind.
package crm_test

import (
	"strings"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/crm"
)

func TestSentinelsClassify(t *testing.T) {
	errorfamilytest.AssertFamily(t, crm.ErrDisabled, errorfamily.Infrastructure)
	errorfamilytest.AssertFamily(t, crm.ErrUnauthorized, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, crm.ErrNotFound, errorfamily.Rejection)
}

func TestLookupStatusSplitAndTransport(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		family errorfamily.Family
	}{
		{"4xx is a rejection", http.StatusNotFound, errorfamily.Rejection},
		{"5xx is transient", http.StatusInternalServerError, errorfamily.Transient},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			client, err := crm.NewClient(srv.URL, "token", "")
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.LookupByPhone(t.Context(), "+441632960961")
			if err == nil {
				t.Fatal("non-200: want error")
			}
			errorfamilytest.AssertFamily(t, err, tc.family)
			errorfamilytest.AssertCode(t, err, "crm.http")
		})
	}

	t.Run("dial refused is transient", func(t *testing.T) {
		closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := closed.URL
		closed.Close()

		client, err := crm.NewClient(url, "token", "")
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.LookupByPhone(t.Context(), "+441632960961")
		if err == nil {
			t.Fatal("dial refused: want error")
		}
		errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
		errorfamilytest.AssertCode(t, err, "crm.transport")
	})
}

// TestCAFileRejectionIsFamilyPinned pins issue #9's boot-side lane: a
// mistyped or certificate-free CA bundle is operator input, so it is a
// Rejection named by its own code, never a transport surprise at the
// first lookup.
func TestCAFileRejectionIsFamilyPinned(t *testing.T) {
	_, err := crm.NewClient("https://crm.example.org", "token", filepath.Join(t.TempDir(), "absent.pem"))
	if err == nil {
		t.Fatal("missing ca bundle: want error")
	}

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "crm.ca_file")
}

func TestCallLogStatusSplit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		family errorfamily.Family
		code   string
	}{
		{"4xx rejection", http.StatusConflict, errorfamily.Rejection, "crm.call_log.rejected"},
		{"5xx transient", http.StatusBadGateway, errorfamily.Transient, "crm.call_log.failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			client, err := crm.NewClient(srv.URL, "token", "")
			if err != nil {
				t.Fatal(err)
			}
			err = client.LogCall(t.Context(), crm.NewContactRef("contact-1"), "out", "+441632960961", 60, "completed")
			if err == nil {
				t.Fatal("non-204: want error")
			}
			errorfamilytest.AssertFamily(t, err, tc.family)
			errorfamilytest.AssertCode(t, err, tc.code)
		})
	}
}

// TestOversizedAnswersAreTheirOwnTransient (T18 F52): a CRM answer
// beyond the read cap is its own Transient verdict (crm.too_large) at
// both read sites — lookup and context — never a boundless buffer or a
// truncation-induced decode surprise.
func TestOversizedAnswersAreTheirOwnTransient(t *testing.T) {
	filler := strings.Repeat(" ", (1<<20)+1) // readCap+1
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, filler)
	}))
	defer srv.Close()

	client, err := crm.NewClient(srv.URL, "token", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.LookupByPhone(t.Context(), "+441632960961")
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertCode(t, err, "crm.too_large")

	_, err = client.ContactContext(t.Context(), crm.NewContactRef("contact-1"))
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertCode(t, err, "crm.too_large")
}
