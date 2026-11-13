package asr

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/larsartmann/go-error-family"
)

// transcriptResponse is the OpenAI-compatible /v1/audio/transcriptions
// wire shape; providers that return a bare object or extra fields are
// tolerated (unknown fields ignored by the decoder).
type transcriptResponse struct {
	Text string `json:"text"`
}

func decodeTranscript(r io.Reader) (string, error) {
	raw, err := io.ReadAll(io.LimitReader(r, 1<<20))
	if err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.read", "asr: read response")
	}
	var parsed transcriptResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.decode", "asr: decode response")
	}
	return strings.TrimSpace(parsed.Text), nil
}

// providerRoundTrip executes one provider request and classifies the
// HTTP verdict — the ONE transport tail every Provider dialect shares
// (the OpenAI-compatible wire and the Google recognize wire differ in
// body shape and decoding, not in how a round trip fails). The returned
// release func drains and closes the response body; the caller defers
// it once done reading. A 429 is load, not misconfiguration: it maps to
// a Transient error so the caller may retry after a beat.
func providerRoundTrip(client *http.Client, httpReq *http.Request) (*http.Response, func(), error) {
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, nil, errorfamily.WrapTransientf(err, "asr.transport", "asr: request")
	}
	release := func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		defer release()
		return nil, nil, ErrUnauthorized
	case resp.StatusCode == http.StatusTooManyRequests:
		defer release()
		return nil, nil, errorfamily.Newf(errorfamily.Transient, "asr.http", "asr: provider answered 429")
	case resp.StatusCode != http.StatusOK:
		defer release()
		family := errorfamily.Transient
		if resp.StatusCode < 500 {
			family = errorfamily.Rejection
		}
		return nil, nil, errorfamily.Newf(family, "asr.http", "asr: provider answered %d", resp.StatusCode)
	}
	return resp, release, nil
}

// buildMultipart assembles the multipart body: the file part carries the
// HONEST Content-Type (a provider that sniffs trusts the header less than
// the bytes, but a truthful header never hurts), plus the model and the
// optional language hint as fields.
func buildMultipart(model string, req Request) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := textproto.MIMEHeader{}
	filename := req.Filename
	if filename == "" {
		filename = defaultFilename
	}
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: create file part")
	}
	if _, err := part.Write(req.Audio); err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: write audio part")
	}
	if err := writer.WriteField("model", model); err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: write model field")
	}
	if req.Language != "" {
		if err := writer.WriteField("language", req.Language); err != nil {
			return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: write language field")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: close multipart body")
	}

	return body, writer.FormDataContentType(), nil
}
