// Package paperless adapts the go-paperless SDK to webphone's optional
// downstream fax archive (fax.Archiver): every INBOUND fax lands in
// Paperless-ngx as a tagged, typed document carrying the webphone fax id
// as provenance. The blob store stays the only storage truth — a failed
// or dead Paperless never affects a fax; the offer is fire-and-forget
// (fax.Service owns the detachment, this package owns the wire).
package paperless

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"sync"
	"time"

	id "github.com/larsartmann/go-branded-id"
	sdk "github.com/larsartmann/go-paperless"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
)

// The lazily-ensured Paperless-ngx metadata (plan 2026-09-30): one tag,
// one document type, one provenance custom field. English names like
// every operator-greppable string in this repo. The three ids are
// branded so a transposed ensure/assign cannot file a fax under the
// wrong metadata (they are plain ints on the SDK wire).
type TagBrand struct{}

// Name implements the id.Brand interface.
func (TagBrand) Name() string { return "PaperlessTag" }

// TagID identifies one Paperless-ngx tag.
type TagID = id.ID[TagBrand, int]

// DocTypeBrand brands Paperless-ngx document-type identifiers.
type DocTypeBrand struct{}

// Name implements the id.Brand interface.
func (DocTypeBrand) Name() string { return "PaperlessDocType" }

// DocTypeID identifies one Paperless-ngx document type.
type DocTypeID = id.ID[DocTypeBrand, int]

// FieldBrand brands Paperless-ngx custom-field identifiers.
type FieldBrand struct{}

// Name implements the id.Brand interface.
func (FieldBrand) Name() string { return "PaperlessField" }

// FieldID identifies one Paperless-ngx custom field.
type FieldID = id.ID[FieldBrand, int]

const (
	tagName         = "fax"
	docTypeName     = "Fax"
	provenanceField = "webphone-fax-id"

	// taskPollInterval paces the consumption-task poll; the whole offer
	// is bounded by the caller's context (fax.Service's archiveTimeout).
	taskPollInterval = 2 * time.Second
)

// Compile-time seam conformance.
var _ fax.Archiver = (*Archiver)(nil)

// Archiver files inbound faxes into one Paperless-ngx instance.
type Archiver struct {
	client *sdk.Client
	log    *slog.Logger

	mu      sync.Mutex
	ensured bool
	tagID   TagID
	typeID  DocTypeID
	fieldID FieldID
}

// NewArchiver builds the adapter. An empty url+token pair means the
// integration is OFF: it returns a nil fax.Archiver (interface-typed on
// purpose — a nil *Archiver would smuggle a non-nil interface into
// fax.New) so the composition root can hand it straight to fax.New.
// Config validation (both-or-neither) rejects half a configuration
// before this runs; the SDK constructor classifies an unusable URL as a
// Rejection at origin.
func NewArchiver(baseURL, token string, log *slog.Logger) (fax.Archiver, error) {
	if baseURL == "" && token == "" {
		return nil, nil
	}
	client, err := sdk.New(baseURL, token)
	if err != nil {
		return nil, fmt.Errorf("paperless: %w", err) //nolint:erraudit // family-neutral propagation: the SDK classifies at origin (invalid_url Rejection)
	}
	if log == nil {
		log = slog.Default()
	}
	return &Archiver{client: client, log: log}, nil
}

// ArchiveFax files one inbound fax: ensure the metadata ids (cached
// after the first success), upload the PDF, and wait for the consumption
// verdict. A duplicate refusal (content-hash dedupe) is an INERT success
// — the document already exists and Paperless keeps exactly one copy.
func (a *Archiver) ArchiveFax(ctx context.Context, job domain.FaxJob, pdf []byte) error {
	tagID, typeID, fieldID, err := a.ensureMetadata(ctx)
	if err != nil {
		return err
	}

	taskID, err := a.client.Upload(ctx, sdk.UploadRequest{
		Filename:       path.Base(job.DocumentPath),
		Content:        pdf,
		Title:          faxTitle(job),
		Created:        job.CreatedAt,
		TagIDs:         []int{tagID.Get()},
		DocumentTypeID: typeID.Get(),
		CustomFields:   []sdk.CustomFieldValue{{Field: fieldID.Get(), Value: job.ID.String()}},
	})
	if err != nil {
		return fmt.Errorf("paperless: upload: %w", err) //nolint:erraudit // family-neutral propagation: the SDK classifies at origin
	}

	outcome, err := a.client.WaitForTask(ctx, taskID, taskPollInterval)
	if err != nil {
		return fmt.Errorf("paperless: task: %w", err) //nolint:erraudit // family-neutral propagation: the SDK classifies at origin
	}
	if documentID, _, refused := outcome.Duplicate(); refused { //nolint:erraudit // false positive: Duplicate() returns (int64, bool, bool) — the blank is a bool, not an error
		a.log.Info("paperless: archived (duplicate already on file, inert)",
			"fax_id", job.ID.String(), "document_id", documentID)
		return nil
	}
	a.log.Info("paperless: archived", "fax_id", job.ID.String(), "document_id", outcome.DocumentID)
	return nil
}

// ensureMetadata resolves the tag/document-type/provenance-field ids on
// first use — deliberately NOT at construction, so a boot never probes
// Paperless reachability. A failed ensure is not cached: the next fax
// retries.
func (a *Archiver) ensureMetadata(ctx context.Context) (TagID, DocTypeID, FieldID, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ensured {
		return a.tagID, a.typeID, a.fieldID, nil
	}

	tagID, err := a.client.EnsureTag(ctx, tagName)
	if err != nil {
		return TagID{}, DocTypeID{}, FieldID{}, fmt.Errorf("paperless: ensure tag: %w", err) //nolint:erraudit // family-neutral propagation: the SDK classifies at origin
	}
	typeID, err := a.client.EnsureDocumentType(ctx, docTypeName)
	if err != nil {
		return TagID{}, DocTypeID{}, FieldID{}, fmt.Errorf("paperless: ensure document type: %w", err) //nolint:erraudit // family-neutral propagation: the SDK classifies at origin
	}
	fieldID, err := a.client.EnsureCustomField(ctx, provenanceField)
	if err != nil {
		return TagID{}, DocTypeID{}, FieldID{}, fmt.Errorf("paperless: ensure custom field: %w", err) //nolint:erraudit // family-neutral propagation: the SDK classifies at origin
	}

	a.tagID, a.typeID, a.fieldID, a.ensured = id.NewID[TagBrand](tagID), id.NewID[DocTypeBrand](typeID), id.NewID[FieldBrand](fieldID), true
	return a.tagID, a.typeID, a.fieldID, nil
}

// faxTitle is the operator-facing document title: English, greppable,
// carrying the remote number verbatim (no per-number correspondent is
// minted — the plan rejected that as a correspondent per caller).
func faxTitle(job domain.FaxJob) string {
	return fmt.Sprintf("Fax from %s %s", job.Remote.String(), job.CreatedAt.Format("2006-01-02"))
}
