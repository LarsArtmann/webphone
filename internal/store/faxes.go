package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Faxes persists fax jobs.
type Faxes struct {
	db *sql.DB
}

// NewFaxes builds the fax store.
func NewFaxes(db *sql.DB) *Faxes { return &Faxes{db: db} }

// Create inserts a new fax job.
func (s *Faxes) Create(ctx context.Context, job domain.FaxJob) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO fax_jobs (id, owner, remote, direction, status, pages, document_path, provider_ref, error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID.String(), job.Owner.String(), job.Remote.String(), string(job.Direction),
		string(job.Status), job.Pages, job.DocumentPath, job.ProviderRef, job.Error,
		job.CreatedAt.Unix(), job.UpdatedAt.Unix())
	if err != nil {
		return fmt.Errorf("insert fax job: %w", err)
	}
	return nil
}

// UpdateStatus advances a job's status; providerRef updates the correlation
// id when non-empty. errorText is stored verbatim on failure.
func (s *Faxes) UpdateStatus(
	ctx context.Context, id domain.FaxID, status domain.FaxStatus, providerRef, errorText string, pages int,
) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE fax_jobs
		SET status = ?, provider_ref = CASE WHEN ? != '' THEN ? ELSE provider_ref END,
		    error = ?, pages = ?, updated_at = ?
		WHERE id = ?
	`, string(status), providerRef, providerRef, errorText, pages, time.Now().Unix(), id.String())
	if err != nil {
		return fmt.Errorf("update fax job: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 { //nolint:erraudit // best-effort write; the response is already committed
		return ErrNotFound
	}
	return nil
}

// List returns the owner's fax jobs, newest first.
func (s *Faxes) List(ctx context.Context, owner domain.Extension, limit int) ([]domain.FaxJob, error) {
	return listRows(ctx, s.db, "list fax jobs", `
		SELECT id, owner, remote, direction, status, pages, document_path, provider_ref, error, created_at, updated_at
		FROM fax_jobs WHERE owner = ?
		ORDER BY created_at DESC, rowid DESC LIMIT ?
	`, []any{owner.String(), limit}, scanFax)
}

// Get fetches one job scoped to its owner.
func (s *Faxes) Get(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, owner, remote, direction, status, pages, document_path, provider_ref, error, created_at, updated_at
		FROM fax_jobs WHERE id = ? AND owner = ?
	`, id.String(), owner.String())
	job, err := scanFax(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.FaxJob{}, ErrNotFound
	}
	return job, err
}

// ByProviderRef resolves a job by its gateway correlation id — the status
// webhook's lookup path.
func (s *Faxes) ByProviderRef(ctx context.Context, ref string) (domain.FaxJob, error) {
	if ref == "" {
		return domain.FaxJob{}, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id, owner, remote, direction, status, pages, document_path, provider_ref, error, created_at, updated_at
		FROM fax_jobs WHERE provider_ref = ?
	`, ref)
	job, err := scanFax(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.FaxJob{}, ErrNotFound
	}
	return job, err
}

func scanFax(row rowScanner) (domain.FaxJob, error) {
	var (
		id, owner, remote, direction, status string
		documentPath, providerRef, errMsg    string
		pages                                int
		created, updated                     int64
	)
	if err := row.Scan(&id, &owner, &remote, &direction, &status, &pages,
		&documentPath, &providerRef, &errMsg, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FaxJob{}, err
		}
		return domain.FaxJob{}, fmt.Errorf("scan fax job row (document %s): %w", documentPath, err)
	}
	return domain.FaxJob{
		ID:           domain.MustFaxID(id),
		Owner:        domain.MustParseExtension(owner),
		Remote:       domain.MustParsePhone(remote),
		Direction:    domain.FaxDirection(direction),
		Status:       domain.FaxStatus(status),
		Pages:        pages,
		DocumentPath: documentPath,
		ProviderRef:  providerRef,
		Error:        errMsg,
		CreatedAt:    time.Unix(created, 0),
		UpdatedAt:    time.Unix(updated, 0),
	}, nil
}
