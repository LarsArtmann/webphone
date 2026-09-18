package gateway

import (
	"context"
	"encoding/json"
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
//	form fields : kind=message, owner, to, body, secret
//	files       : one part per attachment (field "attachment", filename kept)
//
// A 2xx answer with body {"provider_ref": "..."} (or a bare token) is the
// acceptance receipt; anything else is an error and the message stays failed.
type provider struct {
	cfg    config.Gateway
	client *http.Client
}

type Webhook struct {
	provider
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

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)
	//nolint:errcheck // documented pattern: writes to a strings.Builder cannot fail
	writer.WriteField("kind", "fax")
	//nolint:errcheck // see above
	writer.WriteField("owner", fax.Owner.String())
	//nolint:errcheck // see above
	writer.WriteField("to", fax.To.String())
	part, err := writer.CreateFormFile("document", "fax.pdf")
	if err != nil {
		return Receipt{}, fmt.Errorf("create fax form file: %w", err)
	}
	if _, err := io.Copy(part, pdf); err != nil {
		return Receipt{}, fmt.Errorf("copy fax pdf: %w", err)
	}
	if err := writer.Close(); err != nil {
		return Receipt{}, fmt.Errorf("close fax form: %w", err)
	}

	return w.post(ctx, w.cfg.WebhookURL+"/fax", writer.FormDataContentType(), strings.NewReader(buf.String()))
}

func messageForm(kind, owner, to, body string, attachments []OutboundAttachment) (io.Reader, string, error) {
	var buf strings.Builder
	writer := multipart.NewWriter(&buf)
	//nolint:errcheck // documented pattern: writes to a strings.Builder cannot fail
	writer.WriteField("kind", kind)
	//nolint:errcheck // see above
	writer.WriteField("owner", owner)
	//nolint:errcheck // see above
	writer.WriteField("to", to)
	//nolint:errcheck // see above
	writer.WriteField("body", body)

	for _, att := range attachments {
		file, err := os.Open(att.Path)
		if err != nil {
			return nil, "", fmt.Errorf("open attachment %s: %w", att.Name, err)
		}
		part, err := writer.CreateFormFile("attachment", att.Name)
		if err != nil {
			_ = file.Close()
			return nil, "", fmt.Errorf("create attachment form file: %w", err)
		}
		if _, err := io.Copy(part, file); err != nil {
			_ = file.Close()
			return nil, "", fmt.Errorf("copy attachment %s: %w", att.Name, err)
		}
		_ = file.Close()
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("close message form: %w", err)
	}

	return strings.NewReader(buf.String()), writer.FormDataContentType(), nil
}

func (p provider) post(
	ctx context.Context, url, contentType string, body io.Reader,
) (Receipt, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return Receipt{}, fmt.Errorf("build provider request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	if p.cfg.WebhookSecret != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.WebhookSecret)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return Receipt{}, fmt.Errorf("provider call: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Receipt{}, fmt.Errorf("provider rejected: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}

	var receipt struct {
		ProviderRef string `json:"provider_ref"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&receipt); err != nil {
		return Receipt{}, fmt.Errorf("decode provider receipt: %w", err)
	}

	return Receipt{ProviderRef: receipt.ProviderRef}, nil
}

// DefaultClient is the shared HTTP client for provider calls.
func DefaultClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
