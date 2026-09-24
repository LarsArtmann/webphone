// Family classification at the outbound seam (SUPERB error-excellence
// T02/T03): every failure the webhook POST can produce carries a family
// BEFORE it leaves the gateway, so the server's user-surface switch can
// consume errorfamily.Classify instead of a growing type ladder. The
// codes pin that the classification is OURS — the library's fail-open
// default would classify any untagged error as Transient too, so a
// missing wrap must fail these tests, not pass by accident.
package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
)

func postProbe(t *testing.T, gw *Webhook, attachments ...OutboundAttachment) error {
	t.Helper()
	_, err := gw.SendMessage(context.Background(), OutboundMessage{
		Owner:       domain.MustParseExtension("1001"),
		To:          domain.MustParsePhone("+441632960961"),
		Body:        "family probe",
		Attachments: attachments,
	})
	if err == nil {
		t.Fatal("expected an error")
	}

	return err
}

func TestPostFailuresCarryFamilies(t *testing.T) {
	newRejected := func(t *testing.T, status int) error {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "provider says no", status)
		}))
		t.Cleanup(srv.Close)

		return postProbe(t, webhookGateway(srv.URL, "s", srv.Client()))
	}

	newReceiptDecode := func(t *testing.T) error {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"broken`)) // 2xx, unparseable, not a bare token
		}))
		t.Cleanup(srv.Close)

		return postProbe(t, webhookGateway(srv.URL, "s", srv.Client()))
	}

	transport := func(t *testing.T) error {
		t.Helper()
		return postProbe(t, webhookGateway("http://127.0.0.1:1", "s", http.DefaultClient))
	}

	formBuild := func(t *testing.T) error {
		t.Helper()
		gw := webhookGateway("http://127.0.0.1:1", "s", http.DefaultClient)

		return postProbe(t, gw, OutboundAttachment{
			Name: "gone.png", MimeType: "image/png",
			Path: filepath.Join(t.TempDir(), "does-not-exist.png"),
		})
	}

	for _, tc := range []struct {
		name       string
		err        error
		wantFamily errorfamily.Family
		wantCode   string
	}{
		{
			name:       "transport failure is transient",
			err:        func() error { return transport(t) }(),
			wantFamily: errorfamily.Transient,
			wantCode:   "gateway.transport",
		},
		{
			name:       "provider 4xx answer is a rejection",
			err:        func() error { return newRejected(t, http.StatusForbidden) }(),
			wantFamily: errorfamily.Rejection,
		},
		{
			name:       "provider 5xx answer is transient",
			err:        func() error { return newRejected(t, http.StatusBadGateway) }(),
			wantFamily: errorfamily.Transient,
		},
		{
			name:       "unparseable 2xx receipt is transient",
			err:        func() error { return newReceiptDecode(t) }(),
			wantFamily: errorfamily.Transient,
			wantCode:   "gateway.receipt",
		},
		{
			name:       "form-build failure (attachment unreadable) is infrastructure",
			err:        func() error { return formBuild(t) }(),
			wantFamily: errorfamily.Infrastructure,
			wantCode:   "gateway.form",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errorfamilytest.AssertFamily(t, tc.err, tc.wantFamily)
			errorfamilytest.AssertCode(t, tc.err, tc.wantCode)
		})
	}

	t.Run("families survive the service %w wrap", func(t *testing.T) {
		// messaging.Send/fax.Send wrap gateway errors with
		// fmt.Errorf("gateway: %w", err) — the classification must ride
		// the chain so the server switch never sees an unclassified error.
		for _, raw := range []error{
			newRejected(t, http.StatusForbidden),
			newRejected(t, http.StatusBadGateway),
			transport(t),
		} {
			wrapped := fmt.Errorf("gateway: %w", raw)
			if errorfamily.Classify(wrapped) != errorfamily.Classify(raw) {
				t.Errorf("wrap changed the family of %v", raw)
			}
		}
	})

	t.Run("rejections keep their type for the detail fast path", func(t *testing.T) {
		raw := newRejected(t, http.StatusForbidden)
		wrapped := fmt.Errorf("gateway: %w", raw)
		if _, ok := errors.AsType[*ErrProviderRejected](wrapped); !ok {
			t.Errorf("family tagging must not hide *ErrProviderRejected: %v", wrapped)
		}
	})
}

func TestFaxFormFailureIsInfrastructure(t *testing.T) {
	gw := faxWebhookGateway("http://127.0.0.1:1", "s", http.DefaultClient)
	pdf := filepath.Join(t.TempDir(), "missing.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.4 fax"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := gw.SendFax(context.Background(), OutboundFax{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParsePhone("+441632960961"),
		// The spooled PDF is deleted between spool and send: the open
		// fails before any provider traffic, classifying the lane as ours.
		PDFPath: filepath.Join(filepath.Dir(pdf), "gone.pdf"),
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, err, "gateway.form")
}
