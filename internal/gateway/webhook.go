package gateway

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/config"
)

// Webhook forwards outbound messages to a provider URL as multipart/form-data:
//
//	form fields : kind=message, owner, to, body
//	files       : one part per attachment (field "attachment", filename kept)
//
// The secret rides the Authorization header, not a form field. A 2xx
// answer with body {"provider_ref": "..."} (or a bare token) is the
// acceptance receipt; anything else is an error and the message stays
// failed.
type provider struct {
	cfg    config.Gateway
	client *http.Client
}

type Webhook struct {
	provider
}

// ErrProviderRejected is a provider ANSWER refusing the send (non-2xx with
// a detail body) — e.g. "Source and destination cannot be the same number".
// Unlike transport failures these are user-actionable and safe to surface:
// the detail comes from the provider's own rejection, not our internals.
type ErrProviderRejected struct {
	Status int
	Detail string
}

func (e *ErrProviderRejected) Error() string {
	return fmt.Sprintf("provider rejected: HTTP %d: %s", e.Status, e.Detail)
}

// unwrapErrorJSON extracts the inner "error" string from a JSON error body
// ({"error": "…"}) so users see the reason, not the envelope. Non-JSON
// bodies pass through trimmed.
func unwrapErrorJSON(raw []byte) string {
	trimmed := strings.TrimSpace(string(raw))
	var envelope struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(trimmed), &envelope); err == nil && strings.TrimSpace(envelope.Error) != "" {
		return strings.TrimSpace(envelope.Error)
	}
	return trimmed
}

// SendMessage posts the message to the provider.
func (w *Webhook) SendMessage(ctx context.Context, msg OutboundMessage) (Receipt, error) {
	body, contentType, err := messageForm("message", msg.Owner.String(), msg.To.String(), msg.Body, msg.Attachments)
	if err != nil {
		return Receipt{}, fmt.Errorf("build message form: %w", err)
	}
	return w.post(ctx, w.cfg.WebhookURL+"/message", contentType, body)
}

// FaxWebhook forwards outbound faxes to a provider URL as multipart/form-data
// with the PDF as the sole file part.
type FaxWebhook struct {
	provider
}

// SendFax posts the fax job to the provider.
func (w *FaxWebhook) SendFax(ctx context.Context, fax OutboundFax) (Receipt, error) {
	pdf, err := os.Open(fax.PDFPath)
	if err != nil {
		return Receipt{}, fmt.Errorf("open fax pdf: %w", err)
	}
	defer func() { _ = pdf.Close() }()

	body, contentType, err := providerForm("fax", fax.Owner.String(), fax.To.String(),
		func(writer *multipart.Writer) error {
			part, err := writer.CreateFormFile("document", "fax.pdf")
			if err != nil {
				return fmt.Errorf("create fax form file: %w", err)
			}
			if _, err := io.Copy(part, pdf); err != nil {
				return fmt.Errorf("copy fax pdf: %w", err)
			}
			return nil
		})
	if err != nil {
		return Receipt{}, fmt.Errorf("build fax form: %w", err)
	}

	return w.post(ctx, w.cfg.WebhookURL+"/fax", contentType, body)
}

// providerForm writes the provider envelope every webhook shares (the
// kind/owner/to fields), hands the writer to addFiles for the payload
// parts, then closes the form. Field writes are infallible: a
// strings.Builder cannot fail.
func providerForm(kind, owner, to string, addFiles func(*multipart.Writer) error) (io.Reader, string, error) {
	var buf strings.Builder
	writer := multipart.NewWriter(&buf)
	//nolint:errcheck // documented pattern: writes to a strings.Builder cannot fail
	writer.WriteField("kind", kind)
	//nolint:errcheck // see above
	writer.WriteField("owner", owner)
	//nolint:errcheck // see above
	writer.WriteField("to", to)
	if err := addFiles(writer); err != nil {
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("close provider form: %w", err)
	}

	return strings.NewReader(buf.String()), writer.FormDataContentType(), nil
}

func messageForm(kind, owner, to, body string, attachments []OutboundAttachment) (io.Reader, string, error) {
	return providerForm(kind, owner, to, func(writer *multipart.Writer) error {
		//nolint:errcheck // documented pattern: writes to a strings.Builder cannot fail
		writer.WriteField("body", body)

		for _, att := range attachments {
			file, err := os.Open(att.Path)
			if err != nil {
				return fmt.Errorf("open attachment %s: %w", att.Name, err)
			}
			part, err := writer.CreateFormFile("attachment", att.Name)
			if err != nil {
				_ = file.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
				return fmt.Errorf("create attachment form file: %w", err)
			}
			if _, err := io.Copy(part, file); err != nil {
				_ = file.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
				return fmt.Errorf("copy attachment %s: %w", att.Name, err)
			}
			_ = file.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
		}
		return nil
	})
}

func (p provider) post(
	ctx context.Context, url, contentType string, body io.Reader,
) (Receipt, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return Receipt{}, fmt.Errorf("build provider request (content-type %s): %w", contentType, err)
	}
	req.Header.Set("Content-Type", contentType)
	if p.cfg.WebhookSecret != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.WebhookSecret)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return Receipt{}, fmt.Errorf("provider call to %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512)) //nolint:erraudit // best-effort write; the response is already committed
		return Receipt{}, &ErrProviderRejected{
			Status: resp.StatusCode,
			Detail: unwrapErrorJSON(detail),
		}
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return Receipt{}, fmt.Errorf("read provider receipt (content-type %s): %w", contentType, err)
	}
	var receipt struct {
		ProviderRef string `json:"provider_ref"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		// The documented contract also allows a bare provider token as
		// the acceptance receipt: any small answer that is not a JSON
		// object is taken verbatim as the ref.
		if ref := strings.TrimSpace(string(raw)); ref != "" && len(ref) <= 256 && !strings.ContainsRune(ref, '{') {
			return Receipt{ProviderRef: ref}, nil
		}
		return Receipt{}, fmt.Errorf("decode provider receipt (content-type %s): %w", contentType, err)
	}

	return Receipt{ProviderRef: receipt.ProviderRef}, nil
}

// DefaultClient is the shared HTTP client for provider calls.
func DefaultClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
