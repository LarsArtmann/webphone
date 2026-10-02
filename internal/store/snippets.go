package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// Snippets persists per-extension reply snippets (M21): canned texts
// the composer picks from; quick ones double as chips.
type Snippets struct {
	db *sql.DB
}

// SnippetsMaxPerExtension bounds one extension's snippet list (the
// contacts-cap posture: the bound is the durability story, the UI
// renders the whole list).
const SnippetsMaxPerExtension = 100

// ErrSnippetListFull is returned when a NEW snippet would exceed the
// per-owner cap; replacing an existing id stays open.
var ErrSnippetListFull = errors.New("snippet list full") //nolint:erraudit // sentinel: identity, not an error family (plan guardrail #4); classified via init registration

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrSnippetListFull: errorfamily.Rejection,
	})
}

// NewSnippets builds the snippet store.
func NewSnippets(db *sql.DB) *Snippets { return &Snippets{db: db} }

// List returns the owner's snippets, oldest first — a stable order for
// the picker and the chips.
func (s *Snippets) List(ctx context.Context, owner domain.Extension) ([]domain.Snippet, error) {
	return listRows(ctx, s.db, "list snippets", `
		SELECT id, owner, body, quick, created_at
		FROM snippets WHERE owner = ?
		ORDER BY created_at, rowid
	`, []any{owner.String()}, scanSnippet)
}

func scanSnippet(row rowScanner) (domain.Snippet, error) {
	var (
		id, owner, body string
		quick           int
		createdAt       int64
	)
	if err := row.Scan(&id, &owner, &body, &quick, &createdAt); err != nil {
		return domain.Snippet{}, errorfamily.WrapInfrastructuref(err, "store.snippet_scan", "scan snippet %s", id)
	}
	return domain.Snippet{
		ID:        domain.MustSnippetID(id),
		Owner:     domain.MustParseExtension(owner),
		Body:      body,
		Quick:     quick == 1,
		CreatedAt: time.Unix(createdAt, 0),
	}, nil
}

// Save inserts a snippet, or replaces the owner's snippet with the same
// id (the edit path). A NEW id past SnippetsMaxPerExtension inserts
// nothing and returns ErrSnippetListFull — the whole check lives in the
// INSERT's WHERE clause, so the cap holds under concurrent saves.
func (s *Snippets) Save(ctx context.Context, snippet domain.Snippet) error {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO snippets (id, owner, body, quick, created_at)
		SELECT ?, ?, ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM snippets WHERE id = ? AND owner = ?)
		   OR (SELECT COUNT(*) FROM snippets WHERE owner = ?) < ?
		ON CONFLICT(id) DO UPDATE SET
			body = excluded.body,
			quick = excluded.quick
	`, snippet.ID.String(), snippet.Owner.String(), snippet.Body, boolInt(snippet.Quick),
		snippet.CreatedAt.Unix(),
		snippet.ID.String(), snippet.Owner.String(),
		snippet.Owner.String(), SnippetsMaxPerExtension)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.snippet_save", "save snippet %s", snippet.ID)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return ErrSnippetListFull
	}
	return nil
}

// Delete removes one snippet scoped to the owner.
func (s *Snippets) Delete(ctx context.Context, owner domain.Extension, id domain.SnippetID) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM snippets WHERE id = ? AND owner = ?
	`, id.String(), owner.String())
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.snippet_delete", "delete snippet %s", id)
	}
	return updatedOrNotFound(res)
}
