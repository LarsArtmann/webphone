// Package fax is the application service for fax jobs: outbound PDF sends
// through the gateway, inbound documents from webhooks, and status updates.
package fax

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// Limits for fax uploads.
const (
	MaxPDFSize   = 20 << 20 // 20 MiB
	faxPageSize  = 100
	pdfSignature = "%PDF-"

	// archiveTimeout bounds one downstream archive offer (metadata ensure
	// + upload + consumption-task poll). The archive runs detached, so
	// this never delays a fax — it only caps the goroutine's life.
	archiveTimeout = 2 * time.Minute
)

// ErrInvalidFax describes a rejected fax with a user-facing message.
type ErrInvalidFax struct{ Reason string }

func (e *ErrInvalidFax) Error() string { return e.Reason }

// ErrorFamily marks every validation refusal as a Rejection: the user's
// own input, not retryable. The Reason string stays the rendered copy
// (English, runbook-greppable) — the family only drives behavior.
func (e *ErrInvalidFax) ErrorFamily() errorfamily.Family { return errorfamily.Rejection }

// Archiver is the optional downstream filing home for INBOUND faxes
// (Paperless-ngx today, internal/paperless). The blob store stays the
// only storage truth: an archive failure is logged, never surfaced, and
// a nil archiver means the integration is off.
type Archiver interface {
	ArchiveFax(ctx context.Context, job domain.FaxJob, pdf []byte) error
}

// Store is the fax-job persistence seam: the contract the SQL store
// held and the read-model engine now serves (same methods, same
// verdicts). AdvanceIdempotent is the webhook lane's in-process replay
// guard — see readmodels.Faxes for the exact semantics.
type Store interface {
	Create(ctx context.Context, job domain.FaxJob) error
	UpdateStatus(ctx context.Context, id domain.FaxID, status domain.FaxStatus, providerRef, errorText string, pages int) error
	AdvanceIdempotent(ctx context.Context, idemKey string, id domain.FaxID, status domain.FaxStatus, providerRef, errorText string, pages int) error
	List(ctx context.Context, owner domain.Extension, limit int) ([]domain.FaxJob, error)
	Get(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error)
	ByProviderRef(ctx context.Context, ref string) (domain.FaxJob, error)
}

// Service wires the fax store, blob store, and outbound gateway.
// branching-flow COMPOSITION_mixin (baselined): intentional symmetry
// with messaging.Service (the documented seam parallel); a mixin would
// couple the two lanes.
type Service struct {
	faxes      Store
	blobs      *blob.Store
	gateway    gateway.FaxGateway
	onChange   domain.ChangeFunc
	clock      func() time.Time
	identities map[string]string
	archiver   Archiver
}

// selfSendDetail is the local refusal text for a fax addressed to the
// extension's own DID (config identities) — the send-failure train C
// fast path. English like every service reason (operator-greppable).
const selfSendDetail = "faxes cannot be sent to your own number"

// New builds the fax service. onChange may be nil. identities maps an
// extension to its presented DID (config identities); a fax to the owner's
// own DID is refused locally before the gateway roundtrip (nil or missing
// entry = guard off). archiver is the optional Paperless-ngx filing home
// for inbound faxes (nil = integration off).
func New(
	faxes Store, blobs *blob.Store, gw gateway.FaxGateway, onChange domain.ChangeFunc, identities map[string]string, archiver Archiver,
) *Service {
	return &Service{
		faxes: faxes, blobs: blobs, gateway: gw, onChange: onChange, clock: time.Now,
		identities: identities, archiver: archiver,
	}
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
	if len(pdf) > MaxPDFSize {
		return domain.FaxJob{}, &ErrInvalidFax{
			Reason: fmt.Sprintf("PDF larger than %d MiB", MaxPDFSize>>20),
		}
	}

	now := s.clock()
	path, err := s.blobs.Save("faxes", ".pdf", pdf)
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "store.fax_spool", "spool pdf")
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
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "store.fax_create", "persist fax job")
	}
	s.emit(ctx, domain.FaxAdvanced{Owner: owner, Fax: job.ID})

	return s.drive(ctx, job)
}

// drive submits one persisted queued job to the gateway and records the
// verdict on the job row. Shared by Send (fresh upload) and Resend
// (re-submission of a stored document): the gateway path, the self-send
// refusal, and the status transitions have exactly one home.
func (s *Service) drive(ctx context.Context, job domain.FaxJob) (domain.FaxJob, error) {
	// Self-send fast path (train C): refuse locally instead of burning a
	// provider roundtrip the provider would refuse anyway; the failed job
	// below is the evidence-preserving half of the design.
	var receipt gateway.Receipt
	var err error
	if rejection := gateway.SelfSendRejection(s.identities, job.Owner, job.Remote, selfSendDetail); rejection != nil {
		err = rejection
	} else {
		receipt, err = s.gateway.SendFax(ctx, gateway.OutboundFax{
			Owner: job.Owner, To: job.Remote, PDFPath: s.blobs.Abs(job.DocumentPath),
		})
	}
	if err != nil {
		// Queued by construction (Send and Resend drive fresh jobs), so
		// the move is legal; the machine keeps the transition honest anyway.
		job.Status, _ = domain.FaxAdvance(job.Status, domain.FaxSubmitFailed())
		job.Error = err.Error()
		if updateErr := s.faxes.UpdateStatus(ctx, job.ID, domain.FaxFailed, "", job.Error, 0); updateErr != nil {
			slog.Warn("fax: mark failed", "error", updateErr)
		}
		s.emit(ctx, domain.FaxAdvanced{Owner: job.Owner, Fax: job.ID})
		return job, fmt.Errorf("gateway: %w", err) //nolint:erraudit // family-neutral propagation: the gateway owns the family (Rejection/Transient by provider answer)
	}

	// The receipt itself says whether this acceptance is already the
	// verdict (loopback resolves immediate) or the webhook still owes
	// the outcome (deferred). No gateway-type sniffing: the Receipt is
	// the contract.
	job.Status, _ = domain.FaxAdvance(job.Status, domain.FaxAccepted())
	job.ProviderRef = receipt.ProviderRef
	if updateErr := s.faxes.UpdateStatus(ctx, job.ID, domain.FaxSending, receipt.ProviderRef, "", 0); updateErr != nil {
		slog.Warn("fax: mark sending", "error", updateErr)
	}
	if receipt.Resolution == gateway.ResolutionImmediate {
		job.Status, _ = domain.FaxAdvance(job.Status, domain.FaxSettled())
		if updateErr := s.faxes.UpdateStatus(ctx, job.ID, domain.FaxTransmitted, "", "", 1); updateErr != nil {
			slog.Warn("fax: mark transmitted", "error", updateErr)
		}
	}
	s.emit(ctx, domain.FaxAdvanced{Owner: job.Owner, Fax: job.ID})

	return job, nil
}

// Resend re-submits a FAILED outbound job's stored document as a NEW
// job: no re-upload — the spooled PDF is reused — and the original row
// stays untouched as the failure evidence. Anything but a failed
// outbound fax is refused (ErrInvalidFax: re-sending an in-flight,
// delivered, or inbound row would duplicate or fabricate traffic).
func (s *Service) Resend(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error) {
	job, err := s.faxes.Get(ctx, owner, id)
	if err != nil {
		return domain.FaxJob{}, err
	}
	if job.Direction != domain.FaxOutbound || job.Status != domain.FaxFailed {
		return domain.FaxJob{}, &ErrInvalidFax{Reason: "only failed outbound faxes can be resent"}
	}
	// The document must still exist before a new job references it: a
	// loopback gateway would otherwise "transmit" a resend whose PDF is
	// gone (the probe is the honesty guard, drive never opens the file).
	probe, err := s.blobs.Open(job.DocumentPath)
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "store.fax_spool", "reopen fax document")
	}
	_ = probe.Close() //nolint:erraudit // existence probe: nothing to read

	now := s.clock()
	retry := domain.FaxJob{
		ID:           domain.GenerateFaxID(),
		Owner:        owner,
		Remote:       job.Remote,
		Direction:    domain.FaxOutbound,
		Status:       domain.FaxQueued,
		DocumentPath: job.DocumentPath,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.faxes.Create(ctx, retry); err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "store.fax_create", "persist fax retry")
	}
	s.emit(ctx, domain.FaxAdvanced{Owner: owner, Fax: retry.ID})

	return s.drive(ctx, retry)
}

// Receive ingests an inbound fax document from a webhook.
func (s *Service) Receive(ctx context.Context, inbound domain.InboundFax) (domain.FaxJob, error) {
	now := domain.OrClock(inbound.Received, s.clock)
	if len(inbound.PDFBytes) == 0 {
		return domain.FaxJob{}, &ErrInvalidFax{Reason: "inbound fax carries no document"}
	}
	path, err := s.blobs.Save("faxes", ".pdf", inbound.PDFBytes)
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "store.fax_spool", "spool inbound pdf")
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
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "store.fax_create", "persist inbound fax")
	}
	s.emit(ctx, domain.FaxAdvanced{Owner: inbound.Owner, Fax: job.ID})

	// Fire-and-forget archive offer AFTER persist + notify (the plan's
	// async posture): a slow or dead Paperless never delays or fails the
	// webhook ack, and the blob store keeps the original either way.
	if s.archiver != nil {
		go s.archiveInbound(job)
	}

	return job, nil
}

// archiveInbound offers one persisted inbound fax to the downstream
// archive on its own goroutine: the PDF is re-read from the blob store
// (never the webhook payload), bounded by archiveTimeout. Every failure
// is a WARN with the fax id — operator-greppable, never a user surface.
func (s *Service) archiveInbound(job domain.FaxJob) {
	ctx, cancel := context.WithTimeout(context.Background(), archiveTimeout)
	defer cancel()

	file, err := s.blobs.Open(job.DocumentPath)
	if err != nil {
		slog.Warn("fax: archive: open spooled pdf", "error", err, "fax_id", job.ID.String())
		return
	}
	defer func() { _ = file.Close() }()
	pdf, err := io.ReadAll(file)
	if err != nil {
		slog.Warn("fax: archive: read spooled pdf", "error", err, "fax_id", job.ID.String())
		return
	}
	if err := s.archiver.ArchiveFax(ctx, job, pdf); err != nil {
		slog.Warn("fax: archive", "error", err, "fax_id", job.ID.String())
	}
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
	next, ok := domain.FaxAdvance(job.Status, domain.FaxVerdict(status))
	if !ok {
		// No legal move — today only a verdict against an inbound
		// "received" fact (no lifecycle): no write, no nudge, inert
		// success. Replay dedup is the hooks seam's job (hooksIdem).
		return job, nil
	}
	if err := s.faxes.AdvanceIdempotent(ctx, "faxadv:"+providerRef+":"+string(status), job.ID, next, "", errMsg, pages); err != nil {
		return domain.FaxJob{}, err
	}
	s.emit(ctx, domain.FaxAdvanced{Owner: job.Owner, Fax: job.ID})

	job.Status = next
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

func (s *Service) emit(ctx context.Context, e domain.Event) {
	if s.onChange != nil {
		s.onChange(ctx, e)
	}
}
