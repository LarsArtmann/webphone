// The reply-snippets seam's GATE + VIEW surface: the interface the
// server consumes and the projection row the journal backend folds.
// The engine-backed store that used to implement the gate here retired
// with the readmodels engine (2026-11-11 storage-ladder train C) —
// internal/cqrs/snippets.go is the one implementation now.
package readmodels

import (
	"context"
	"errors"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// SnippetsMaxPerExtension bounds one extension's snippet list (the
// contacts-cap posture: the bound is the durability story, the UI
// renders the whole list).
const SnippetsMaxPerExtension = 100

// ErrSnippetListFull is returned when a NEW snippet would exceed the
// per-owner cap; replacing an existing id stays open. The sentinel
// homed here with the read-model swap (the SQL snippets store left
// with its file, the contract stays).
var ErrSnippetListFull = errors.New("snippet list full")

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrSnippetListFull: errorfamily.Rejection,
	})
}

// SnippetView is the projection row ONE journal fold produces for a
// reply snippet. Unix-nanos created_at keeps list ordering EXACTLY
// numeric in sqlite (the SQL store compared unix seconds with a rowid
// tiebreak; nanos make ties virtually impossible and the order
// stable).
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the retired SQL row verbatim (parity port).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type SnippetView struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Body      string `json:"body"`
	Quick     bool   `json:"quick"`
	CreatedAt int64  `json:"created_at"`
}

// SnippetGate is the reply-snippet seam as the server surfaces see it
// (upsert by id keeping the creation time, owner-scoped oldest-first
// list, owner-scoped delete). The journal-backed service
// (internal/cqrs) is the implementation.
type SnippetGate interface {
	Save(ctx context.Context, snippet domain.Snippet) error
	List(ctx context.Context, owner domain.Extension) ([]domain.Snippet, error)
	Delete(ctx context.Context, owner domain.Extension, id domain.SnippetID) error
}
