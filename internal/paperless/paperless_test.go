package paperless_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/paperless"
)

// paperlessStub is a minimal Paperless-ngx API stand-in: the Ensure*
// lookups answer "not found" once and create on POST; the upload accepts
// the multipart form and answers a fixed task id; the task endpoint
// answers the configured terminal verdict.
type paperlessStub struct {
	t *testing.T

	mu           sync.Mutex
	lookups      map[string]int // path → GET count (metadata caching assertions)
	authHeaders  []string
	uploadForm   map[string]string
	uploadedFile []byte
	taskStatus   string
	taskResult   string // raw result_data object
}

func newPaperlessStub(t *testing.T) *paperlessStub {
	t.Helper()
	return &paperlessStub{
		t:          t,
		lookups:    map[string]int{},
		taskStatus: "success",
		taskResult: `{"document_id":42}`,
	}
}

func (s *paperlessStub) serve(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.authHeaders = append(s.authHeaders, r.Header.Get("Authorization"))

		switch {
		case r.Method == http.MethodGet && isMetadataPath(r.URL.Path):
			s.lookups[r.URL.Path]++
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"results":[]}`)
		case r.Method == http.MethodPost && isMetadataPath(r.URL.Path):
			var created struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}
			created.ID = metadataStubID(r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(created)
		case r.Method == http.MethodPost && r.URL.Path == "/api/documents/post_document/":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				http.Error(w, "bad multipart", http.StatusBadRequest)
				return
			}
			s.uploadForm = map[string]string{}
			for _, key := range []string{"title", "created", "document_type", "tags", "custom_fields"} {
				s.uploadForm[key] = r.FormValue(key)
			}
			file, _, err := r.FormFile("document")
			if err != nil {
				http.Error(w, "missing document part", http.StatusBadRequest)
				return
			}
			defer file.Close()
			s.uploadedFile, _ = io.ReadAll(file)
			_, _ = io.WriteString(w, `"task-1"`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/tasks/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"results":[{"task_id":"task-1","status":"`+s.taskStatus+
				`","result_data":`+s.taskResult+`}]}`)
		default:
			http.Error(w, "unexpected "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func isMetadataPath(path string) bool {
	switch path {
	case "/api/tags/", "/api/document_types/", "/api/custom_fields/":
		return true
	}
	return false
}

func metadataStubID(path string) int {
	switch path {
	case "/api/tags/":
		return 7
	case "/api/document_types/":
		return 3
	case "/api/custom_fields/":
		return 9
	}
	return 0
}

func archiveJob(t *testing.T) domain.FaxJob {
	t.Helper()
	received := time.Date(2026, 10, 1, 12, 30, 0, 0, time.UTC)
	return domain.FaxJob{
		ID:           domain.GenerateFaxID(),
		Owner:        domain.MustParseExtension("100"),
		Remote:       domain.MustParsePhone("+4930123456"),
		Direction:    domain.FaxInbound,
		Status:       domain.FaxReceived,
		DocumentPath: "faxes/abc123.pdf",
		CreatedAt:    received,
		UpdatedAt:    received,
	}
}

var archivePDF = []byte("%PDF-1.4\n%archive\n%%EOF\n")

// TestArchiveFaxFilesTaggedTypedDocument pins the wire offer: one upload
// per fax, tagged + typed + provenance-stamped, titled "Fax from <remote>
// <date>", with the spooled PDF bytes as the document part.
func TestArchiveFaxFilesTaggedTypedDocument(t *testing.T) {
	stub := newPaperlessStub(t)
	server := stub.serve(t)
	archiver, err := paperless.NewArchiver(server.URL, "sekrit", nil)
	if err != nil {
		t.Fatal(err)
	}

	job := archiveJob(t)
	if err := archiver.ArchiveFax(context.Background(), job, archivePDF); err != nil {
		t.Fatalf("archive: %v", err)
	}

	if string(stub.uploadedFile) != string(archivePDF) {
		t.Error("uploaded bytes differ from the spooled pdf")
	}
	want := map[string]string{
		"title":         "Fax from +4930123456 2026-10-01",
		"created":       "2026-10-01",
		"document_type": "3",
		"tags":          "7",
	}
	for key, want := range want {
		if stub.uploadForm[key] != want {
			t.Errorf("form %q = %q, want %q", key, stub.uploadForm[key], want)
		}
	}
	var fields []struct {
		Field int    `json:"field"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(stub.uploadForm["custom_fields"]), &fields); err != nil {
		t.Fatalf("custom_fields decode: %v", err)
	}
	if len(fields) != 1 || fields[0].Field != 9 || fields[0].Value != job.ID.String() {
		t.Errorf("provenance custom field = %+v, want field 9 carrying %s", fields, job.ID.String())
	}
	for _, header := range stub.authHeaders {
		if header != "Token sekrit" {
			t.Fatalf("authorization header = %q, want the static token scheme", header)
		}
	}
}

// TestArchiveFaxCachesMetadata pins the lazy ensure: two faxes run the
// metadata lookups ONCE (per kind) — the ids are cached after the first
// success, not re-resolved per offer.
func TestArchiveFaxCachesMetadata(t *testing.T) {
	stub := newPaperlessStub(t)
	server := stub.serve(t)
	archiver, err := paperless.NewArchiver(server.URL, "sekrit", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := archiver.ArchiveFax(context.Background(), archiveJob(t), archivePDF); err != nil {
		t.Fatal(err)
	}
	if err := archiver.ArchiveFax(context.Background(), archiveJob(t), archivePDF); err != nil {
		t.Fatal(err)
	}

	for path, count := range stub.lookups {
		if count != 1 {
			t.Errorf("%s looked up %d times across two faxes, want exactly 1 (cached)", path, count)
		}
	}
}

// TestArchiveFaxDuplicateIsInertSuccess pins the dedupe posture: a
// content-hash duplicate refusal is an honest outcome, not an error —
// the blob store keeps the original and Paperless keeps one copy.
func TestArchiveFaxDuplicateIsInertSuccess(t *testing.T) {
	stub := newPaperlessStub(t)
	stub.taskStatus = "failure"
	stub.taskResult = `{"duplicate_of":42,"duplicate_in_trash":false}`
	server := stub.serve(t)
	archiver, err := paperless.NewArchiver(server.URL, "sekrit", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := archiver.ArchiveFax(context.Background(), archiveJob(t), archivePDF); err != nil {
		t.Fatalf("duplicate refusal must be inert: %v", err)
	}
}

// TestArchiveFaxTaskFailureSurfaces pins the honest failure arm: a real
// consumption failure returns an error (the caller WARNs; the blob store
// keeps the fax either way).
func TestArchiveFaxTaskFailureSurfaces(t *testing.T) {
	stub := newPaperlessStub(t)
	stub.taskStatus = "failure"
	stub.taskResult = `{"error_type":"ocr","error_message":"OCR failed"}`
	server := stub.serve(t)
	archiver, err := paperless.NewArchiver(server.URL, "sekrit", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := archiver.ArchiveFax(context.Background(), archiveJob(t), archivePDF); err == nil {
		t.Fatal("task failure must surface as an error")
	}
}

// TestNewArchiverDisabled pins the off posture: an empty config pair
// builds NO archiver (nil, nil) — the composition root passes the nil
// straight to fax.New.
func TestNewArchiverDisabled(t *testing.T) {
	archiver, err := paperless.NewArchiver("", "", nil)
	if err != nil {
		t.Fatalf("disabled must not error: %v", err)
	}
	if archiver != nil {
		t.Fatal("disabled must build no archiver")
	}
}

// TestNewArchiverRejectsUnusableURL pins fail-fast construction: the SDK
// classifies an unusable base URL at origin, so a typoed paperless.url
// fails the boot instead of silently archiving nothing.
func TestNewArchiverRejectsUnusableURL(t *testing.T) {
	if _, err := paperless.NewArchiver("ftp://paperless.example.org", "sekrit", nil); err == nil {
		t.Fatal("non-http(s) base URL must fail construction")
	}
}
