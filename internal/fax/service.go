// Package fax is the application service for fax jobs: outbound PDF sends
// through the gateway, inbound documents from webhooks, and status updates.
package fax

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/store"
)

// Limits for fax uploads.
const (
	maxPDFSize = 20 << 20 // 20 MiB
	faxPageSize = 100
	pdfSignature = "%PDF-"
)

// ErrInvalidFax describes a rejected fax with a user-facing message.
type ErrInvalidFax struct{ Reason string }

func (e *ErrInvalidFax) Error() string { return e.Reason }

// ChangeFunc is called after any fax-job mutation (SSE fan-out).
type ChangeFunc func(jobID domain.FaxID)

// Service wires the fax store, blob store, and outbound gateway.
type Service struct {
	faxes    *store.Faxes
	blobs    *blob.Store
	gateway  gateway.FaxGateway
	onChange ChangeFunc
	clock    func() time.Time
}

// New builds the fax service. onChange may be nil.
func New(faxes *store.Faxes, blobs *blob.Store, gw gateway.FaxGateway, onChange ChangeFunc) *Service {
	return &Service{faxes: faxes, blobs: blobs, gateway: gw, onChange: onChange, clock: time.Now}
}

// Send validates the PDF, spools it, creates the job, and drives the
// gateway. Loopback gateways resolve to transmitted synchronously; webhook
// gateways leave the job "sending" until a status webhook lands.
func (s *Service) Send(
	ctx context.Context, owner domain.Extension, to domain.Phone, filename string, pdf []byte,
) (domain.FaxJob, error) {
	if len(pdf) == 0 {
		return domain.FaxJob{}, &ErrInvalidFax{Reason: "attach a PDF to send"}
	}
	if !bytes.HasPrefix(pdf, []byte(pdfSignature)) {
		return domain.FaxJob{}, &ErrInvalidFax{Reason: "only PDF documents can be faxed"}
	}
	if len(pdf) > maxPDFSize {
		return domain.FaxJob{}, &ErrInvalidFax{
			Reason: fmt.Sprintf("PDF larger than %d MiB", maxPDFSize>>20),
		}
	}

	now := s.clock()
	path, err := s.blobs.Save("faxes", ".pdf", pdf)
	if err != nil {
		return domain.FaxJob{}, fmt.Errorf("spool pdf: %w", err)
	}

	job := domain.FaxJob{
		ID:           domain.GenerateFaxID(),
		Owner:        owner,
		Remote:       to,
		Direction:    domain.FaxOutbound,
		Status:       domain.FaxQueued,
		DocumentPath: path,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.faxes.Create(ctx, job); err != nil {
		return domain.FaxJob{}, fmt.Errorf("persist fax job: %w", err)
	}
	s.notify(job.ID)

	receipt, err := s.gateway.SendFax(ctx, gateway.OutboundFax{
		Owner: owner, To: to, PDFPath: s.blobs.Abs(path),
	})
	if err != nil {
		job.Status = domain.FaxFailed
		job.Error = err.Error()
		if updateErr := s.faxes.UpdateStatus(ctx, job.ID, domain.FaxFailed, "", job.Error, 0); updateErr != nil {
			slog.Warn("fax: mark failed", "error", updateErr)
		}
		s.notify(job.ID)
		return job, fmt.Errorf("gateway: %w", err)
	}

	// Loopback resolves instantly (transmitted); webhook providers only
	// accepted the job — the verdict arrives on the status webhook.
	job.Status = domain.FaxSending
	job.ProviderRef = receipt.ProviderRef
	if updateErr := s.faxes.UpdateStatus(ctx, job.ID, domain.FaxSending, receipt.ProviderRef, "", 0); updateErr != nil {
		slog.Warn("fax: mark sending", "error", updateErr)
	}
	if _, isLoopback := s.gateway.(*gateway.Loopback); isLoopback {
		job.Status = domain.FaxTransmitted
		if updateErr := s.faxes.UpdateStatus(ctx, job.ID, domain.FaxTransmitted, "", "", 1); updateErr != nil {
			slog.Warn("fax: mark transmitted", "error", updateErr)
		}
	}
	s.notify(job.ID)

	return job, nil
}

// Receive ingests an inbound fax document from a webhook.
func (s *Service) Receive(ctx context.Context, inbound domain.InboundFax) (domain.FaxJob, error) {
	now := inbound.Received
	if now.IsZero() {
		now = s.clock()
	}
	if len(inbound.PDFBytes) == 0 {
		return domain.FaxJob{}, &ErrInvalidFax{Reason: "inbound fax carries no document"}
	}
	path, err := s.blobs.Save("faxes", ".pdf", inbound.PDFBytes)
	if err != nil {
		return domain.FaxJob{}, fmt.Errorf("spool inbound pdf: %w", err)
	}

	job := domain.FaxJob{
		ID:           domain.GenerateFaxID(),
		Owner:        inbound.Owner,
		Remote:       inbound.From,
		Direction:    domain.FaxInbound,
		Status:       domain.FaxReceived,
		Pages:        inbound.Pages,
		DocumentPath: path,
		ProviderRef:  inbound.ProviderRef,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.faxes.Create(ctx, job); err != nil {
		return domain.FaxJob{}, fmt.Errorf("persist inbound fax: %w", err)
	}
	s.notify(job.ID)

	return job, nil
}

// UpdateProviderStatus advances an outbound job located by provider ref.
func (s *Service) UpdateProviderStatus(
	ctx context.Context, providerRef string, status domain.FaxStatus, pages int, errMsg string,
) (domain.FaxJob, error) {
	job, err := s.faxes.ByProviderRef(ctx, providerRef)
	if err != nil {
		return domain.FaxJob{}, err
	}
	if status == domain.FaxReceived {
		status = domain.FaxTransmitted // an outbound job cannot become "received"
	}
	if err := s.faxes.UpdateStatus(ctx, job.ID, status, "", errMsg, pages); err != nil {
		return domain.FaxJob{}, err
	}
	s.notify(job.ID)

	job.Status = status
	job.Pages = pages
	job.Error = errMsg
	return job, nil
}

// List returns the owner's fax jobs, newest first.
func (s *Service) List(ctx context.Context, owner domain.Extension) ([]domain.FaxJob, error) {
	return s.faxes.List(ctx, owner, faxPageSize)
}

// Get fetches one job scoped to its owner.
func (s *Service) Get(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error) {
	return s.faxes.Get(ctx, owner, id)
}

// Document streams a job's PDF.
func (s *Service) Document(job domain.FaxJob) (io.ReadSeekCloser, error) {
	return s.blobs.Open(job.DocumentPath)
}

func (s *Service) notify(id domain.FaxID) {
	if s.onChange != nil {
		s.onChange(id)
	}
}

// ErrNotFound mirrors store.ErrNotFound for handlers.
var ErrNotFound = errors.New("fax job not found")
