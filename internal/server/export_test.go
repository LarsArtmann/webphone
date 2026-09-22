package server

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// readZipEntry returns one entry's bytes, failing the test when the name
// is absent — the zip IS the export contract, so a missing file is the
// finding, not an error to tolerate.
func readZipEntry(t *testing.T, archive *zip.Reader, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer rc.Close() //nolint:staticcheck,errcheck // test helper: one entry, process exits anyway
		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return content
	}
	t.Fatalf("export zip is missing %s", name)
	return nil
}

// TestExportDataZipsEverythingTheExtensionOwns pins the export contract:
// one session-gated GET returns a zip holding messages.json (every
// thread with its messages), faxes.json and contacts.vcf — all scoped to
// the signed-in extension, with attachment blobs deliberately reduced to
// their names in the JSON manifest.
func TestExportDataZipsEverythingTheExtensionOwns(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	owner := domain.MustParseExtension("1001")
	remote := domain.MustParsePhone("+441632960961")
	ctx := context.Background()

	threadID, err := server.messages.FindThread(ctx, owner, remote, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := server.messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "export-body-probe", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := server.faxes.Create(ctx, domain.FaxJob{
		ID: domain.GenerateFaxID(), Owner: owner, Remote: remote,
		Direction: domain.FaxOutbound, Status: domain.FaxTransmitted, Pages: 2,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	vcf := "BEGIN:VCARD\r\nVERSION:3.0\r\nFN:Carol\r\nTEL:+493012345678\r\nEND:VCARD\r\n"
	form, contentType := multipartBody(t, map[string]string{}, map[string]struct {
		Name    string
		Content []byte
	}{"vcard": {Name: "contacts.vcf", Content: []byte(vcf)}})
	if resp, _ := c.do(http.MethodPost, "/contacts/import", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("contact import status %d", resp.StatusCode)
	}

	resp, body := c.do(http.MethodGet, "/api/export", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export status %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/zip" {
		t.Errorf("content-type %q, want application/zip", got)
	}
	if disposition := resp.Header.Get("Content-Disposition"); !strings.Contains(disposition, "webphone-export-1001-") {
		t.Errorf("content-disposition %q must name the owning extension", disposition)
	}

	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("export body is not a zip: %v", err)
	}

	var threads []exportThread
	if err := json.Unmarshal(readZipEntry(t, archive, "messages.json"), &threads); err != nil {
		t.Fatalf("messages.json: %v", err)
	}
	found := false
	for _, thread := range threads {
		if thread.Remote != remote.String() {
			continue
		}
		for _, message := range thread.Messages {
			if message.Body == "export-body-probe" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("messages.json missing the seeded body (threads: %+v)", threads)
	}

	faxesJSON := string(readZipEntry(t, archive, "faxes.json"))
	if !strings.Contains(faxesJSON, remote.String()) || !strings.Contains(faxesJSON, `"pages":2`) {
		t.Errorf("faxes.json missing the seeded fax: %s", faxesJSON)
	}

	if vcfOut := string(readZipEntry(t, archive, "contacts.vcf")); !strings.Contains(vcfOut, "Carol") {
		t.Errorf("contacts.vcf missing the imported contact: %s", vcfOut)
	}
}

// TestExportDataRequiresSession: the export is a full-data read, so the
// anonymous request must bounce off the session gate.
func TestExportDataRequiresSession(t *testing.T) {
	server := newTestServer(t)
	anon := clientFor(t, server)

	resp, _ := anon.do(http.MethodGet, "/api/export", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous export status %d, want 401", resp.StatusCode)
	}
}
