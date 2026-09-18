package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
)

func configGateway(url, secret string) config.Gateway {
	return config.Gateway{Mode: config.GatewayWebhook, WebhookURL: url, WebhookSecret: secret}
}

func webhookGateway(url, secret string, client *http.Client) *Webhook {
	return &Webhook{provider{cfg: configGateway(url, secret), client: client}}
}

func faxWebhookGateway(url, secret string, client *http.Client) *FaxWebhook {
	return &FaxWebhook{provider{cfg: configGateway(url, secret), client: client}}
}

// capturedRequest records what the fake provider received.
type capturedRequest struct {
	path        string
	auth        string
	contentType string
	form        map[string]string
	attachment  filePart
	document    filePart
}

type filePart struct {
	name    string
	content []byte
	present bool
}

func captureProvider(t *testing.T, respond func(w http.ResponseWriter)) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{form: map[string]string{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.path = r.URL.Path
		captured.auth = r.Header.Get("Authorization")
		captured.contentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("provider could not parse multipart form: %v", err)
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		for _, key := range []string{"kind", "owner", "to", "body"} {
			captured.form[key] = r.FormValue(key)
		}
		captured.attachment = readFilePart(r, "attachment")
		captured.document = readFilePart(r, "document")
		respond(w)
	}))
	t.Cleanup(srv.Close)
	return srv, captured
}

func readFilePart(r *http.Request, field string) filePart {
	file, header, err := r.FormFile(field)
	if err != nil {
		return filePart{}
	}
	content, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil {
		return filePart{}
	}
	return filePart{name: header.Filename, content: content, present: true}
}

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestWebhookSendMessagePostsMultipartWithBearer(t *testing.T) {
	srv, captured := captureProvider(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"provider_ref":"gw-123"}`))
	})

	attPath := writeTempFile(t, t.TempDir(), "pic.jpg", "jpeg-bytes")
	gw := webhookGateway(srv.URL, "s3cret", srv.Client())
	receipt, err := gw.SendMessage(context.Background(), OutboundMessage{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParsePhone("+441632960961"),
		Body:  "hello there",
		Attachments: []OutboundAttachment{
			{Name: "pic.jpg", MimeType: "image/jpeg", Path: attPath},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if receipt.ProviderRef != "gw-123" {
		t.Errorf("provider_ref: got %q, want gw-123", receipt.ProviderRef)
	}
	if captured.path != "/message" {
		t.Errorf("path: got %q, want /message", captured.path)
	}
	if captured.auth != "Bearer s3cret" {
		t.Errorf("authorization: got %q, want Bearer s3cret", captured.auth)
	}
	if !strings.HasPrefix(captured.contentType, "multipart/form-data") {
		t.Errorf("content-type: got %q", captured.contentType)
	}
	if captured.form["kind"] != "message" || captured.form["owner"] != "1001" ||
		captured.form["to"] != "+441632960961" || captured.form["body"] != "hello there" {
		t.Errorf("form fields: %+v", captured.form)
	}
	if !captured.attachment.present || captured.attachment.name != "pic.jpg" ||
		string(captured.attachment.content) != "jpeg-bytes" {
		t.Errorf("attachment part: %+v", captured.attachment)
	}
}

func TestWebhookSendMessageWithoutSecretOmitsAuthorization(t *testing.T) {
	srv, captured := captureProvider(t, func(w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"provider_ref":"ref"}`))
	})

	gw := webhookGateway(srv.URL, "", srv.Client())
	if _, err := gw.SendMessage(context.Background(), OutboundMessage{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParsePhone("+441632960961"),
		Body:  "hi",
	}); err != nil {
		t.Fatal(err)
	}
	if captured.auth != "" {
		t.Errorf("authorization header sent without a configured secret: %q", captured.auth)
	}
}

func TestWebhookSendMessageAcceptsBareTokenReceipt(t *testing.T) {
	srv, _ := captureProvider(t, func(w http.ResponseWriter) {
		_, _ = w.Write([]byte("plain-ref-42\n"))
	})

	gw := webhookGateway(srv.URL, "s", srv.Client())
	receipt, err := gw.SendMessage(context.Background(), OutboundMessage{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParsePhone("+441632960961"),
		Body:  "hi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.ProviderRef != "plain-ref-42" {
		t.Errorf("bare-token receipt: got %q", receipt.ProviderRef)
	}
}

func TestWebhookProviderErrorsStayErrors(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr []string
	}{
		{name: "5xx rejection", status: http.StatusInternalServerError, body: "boom", wantErr: []string{"HTTP 500", "boom"}},
		{name: "malformed receipt", status: http.StatusOK, body: `{"provider_ref"`, wantErr: []string{"decode provider receipt"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := captureProvider(t, func(w http.ResponseWriter) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			gw := webhookGateway(srv.URL, "s", srv.Client())
			_, err := gw.SendMessage(context.Background(), OutboundMessage{
				Owner: domain.MustParseExtension("1001"),
				To:    domain.MustParsePhone("+441632960961"),
				Body:  "hi",
			})
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err.Error(), want)
				}
			}
		})
	}
}

func TestFaxWebhookPostsPDFDocument(t *testing.T) {
	srv, captured := captureProvider(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"provider_ref":"fax-7"}`))
	})

	pdfPath := writeTempFile(t, t.TempDir(), "doc.pdf", "%PDF-fake")
	gw := faxWebhookGateway(srv.URL, "fax-secret", srv.Client())
	receipt, err := gw.SendFax(context.Background(), OutboundFax{
		Owner:   domain.MustParseExtension("1001"),
		To:      domain.MustParsePhone("+441632960961"),
		PDFPath: pdfPath,
	})
	if err != nil {
		t.Fatal(err)
	}

	if receipt.ProviderRef != "fax-7" {
		t.Errorf("provider_ref: got %q", receipt.ProviderRef)
	}
	if captured.path != "/fax" {
		t.Errorf("path: got %q, want /fax", captured.path)
	}
	if captured.auth != "Bearer fax-secret" {
		t.Errorf("authorization: got %q", captured.auth)
	}
	if captured.form["kind"] != "fax" || captured.form["owner"] != "1001" ||
		captured.form["to"] != "+441632960961" {
		t.Errorf("form fields: %+v", captured.form)
	}
	if !captured.document.present || captured.document.name != "fax.pdf" ||
		string(captured.document.content) != "%PDF-fake" {
		t.Errorf("document part: %+v", captured.document)
	}
	if captured.attachment.present {
		t.Error("fax form must not carry an attachment part")
	}
}

func TestFaxWebhookMissingPDFIsAnError(t *testing.T) {
	srv, _ := captureProvider(t, func(w http.ResponseWriter) {})
	gw := faxWebhookGateway(srv.URL, "s", srv.Client())
	if _, err := gw.SendFax(context.Background(), OutboundFax{
		Owner:   domain.MustParseExtension("1001"),
		To:      domain.MustParsePhone("+441632960961"),
		PDFPath: filepath.Join(t.TempDir(), "absent.pdf"),
	}); err == nil {
		t.Error("missing PDF must be an error")
	}
}
