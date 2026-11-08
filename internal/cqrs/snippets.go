// The snippets aggregate (SUPERB Phase 3, A21): the reply-snippet list
// as a per-owner stream. The replace-keeps-creation-time semantics and
// the 100-per-owner cap for NEW ids live in the DECISION now (the old
// store pre-read the row under a mutex; the decider simply knows its
// state), and the ErrSnippetListFull contract survives the swap
// verbatim.
package cqrs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// SnippetsStreamType keys per-owner snippet streams.
const SnippetsStreamType = "snippets"

// snippetsMaxSlots mirrors readmodels.SnippetsMaxPerExtension at the
// aggregate.
const snippetsMaxSlots = readmodels.SnippetsMaxPerExtension

// SaveSnippet upserts one snippet by id (replace keeps the slot's
// creation time — the decider emits the PRESERVED stamp).
type SaveSnippet struct {
	*command.BasicCommand
	Snippet domain.Snippet
}

// DeleteSnippet removes one snippet scoped to its owner (the state IS
// the owner's list, so a miss is the not-found rejection).
type DeleteSnippet struct {
	*command.BasicCommand
	// SnippetID (the field name ID would shadow the embedded
	// command's ID() method and break the Command interface).
	SnippetID domain.SnippetID
}

// snippetSlot is one snippet's aggregate state.
type snippetSlot struct {
	Body      string
	Quick     bool
	CreatedAt int64 // unix nanos — the list's exact numeric order
}

// snippetState is the per-owner snippet list.
type snippetState struct {
	Slots map[string]snippetSlot
}

func (snippetState) initial() snippetState { return snippetState{Slots: map[string]snippetSlot{}} }

func applySnippetState(s snippetState, evt event.Event) (snippetState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.SnippetSaved{})):
		fact, err := event.DecodePayloadAuto[domain.SnippetSaved](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.snippet_fold", "decode SnippetSaved")
		}
		s.Slots[fact.Key.ID] = snippetSlot{Body: fact.Body, Quick: fact.Quick, CreatedAt: fact.CreatedAt.UnixNano()}
	case event.Type(metaengine.EventTypeName(domain.SnippetDeleted{})):
		fact, err := event.DecodePayloadAuto[domain.SnippetDeleted](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.snippet_fold", "decode SnippetDeleted")
		}
		delete(s.Slots, fact.Key.ID)
	}
	return s, nil
}

func decideSaveSnippet(s snippetState, ver event.Version, cmd SaveSnippet) ([]event.Event, error) {
	owner := cmd.Snippet.Owner.String()
	id := cmd.Snippet.ID.String()
	createdAt := cmd.Snippet.CreatedAt
	if prev, ok := s.Slots[id]; ok {
		createdAt = time.Unix(0, prev.CreatedAt)
	} else if len(s.Slots) >= snippetsMaxSlots {
		return nil, errorfamily.WrapRejectionf(readmodels.ErrSnippetListFull, "cqrs.snippet_list_full",
			"%s at %d snippets", owner, snippetsMaxSlots)
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.SnippetSaved{})), cmd.StreamID(), SnippetsStreamType, ver+1, domain.SnippetSaved{
		Key:       domain.SnippetKey{ID: id},
		Owner:     owner,
		Body:      cmd.Snippet.Body,
		Quick:     cmd.Snippet.Quick,
		CreatedAt: createdAt,
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.snippet_save", "build SnippetSaved")
	}
	return []event.Event{evt}, nil
}

func decideDeleteSnippet(s snippetState, ver event.Version, cmd DeleteSnippet) ([]event.Event, error) {
	if _, ok := s.Slots[cmd.SnippetID.String()]; !ok {
		return nil, errorfamily.NewRejection("cqrs.snippet_missing", fmt.Sprintf("snippet %s", cmd.SnippetID.String()))
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.SnippetDeleted{})), cmd.StreamID(), SnippetsStreamType, ver+1, domain.SnippetDeleted{
		Key: domain.SnippetKey{ID: cmd.SnippetID.String()},
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.snippet_delete", "build SnippetDeleted")
	}
	return []event.Event{evt}, nil
}

// RegisterSnippets wires the snippets aggregate into a System.
func RegisterSnippets(sys *system.System) error {
	if err := system.RegisterDecider[snippetState](sys, SnippetsStreamType, decider.Decider[snippetState]{
		Initial: snippetState{}.initial(),
		Apply:   applySnippetState,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.snippet_register", "register snippets decider")
	}
	if err := system.RegisterCommand[*SaveSnippet, snippetState](sys, "snippet.save",
		func(_ context.Context, cmd *SaveSnippet) system.Op[snippetState] {
			return system.Execute[snippetState](context.Background(), cmd.StreamID(), SnippetsStreamType,
				func(state snippetState, ver event.Version) ([]event.Event, error) {
					return decideSaveSnippet(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.snippet_register", "register snippet.save")
	}
	if err := system.RegisterCommand[*DeleteSnippet, snippetState](sys, "snippet.delete",
		func(_ context.Context, cmd *DeleteSnippet) system.Op[snippetState] {
			return system.Execute[snippetState](context.Background(), cmd.StreamID(), SnippetsStreamType,
				func(state snippetState, ver event.Version) ([]event.Event, error) {
					return decideDeleteSnippet(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.snippet_register", "register snippet.delete")
	}
	return nil
}

// snippetsProjection folds SnippetSaved/SnippetDeleted facts into the
// SAME view shape the server reads today (readmodels.SnippetView).
func snippetsProjection() metaengine.QueryDecl[snippetKeyInput, readmodels.SnippetView] {
	return metaengine.Query[snippetKeyInput, readmodels.SnippetView]("snippets",
		metaengine.OnRecord(domain.SnippetSaved{},
			func(_ record.Record, e domain.SnippetSaved) (domain.SnippetKey, readmodels.SnippetView) {
				return e.Key, readmodels.SnippetView{
					ID:        e.Key.ID,
					Owner:     e.Owner,
					Body:      e.Body,
					Quick:     e.Quick,
					CreatedAt: e.CreatedAt.UnixNano(),
				}
			}),
		metaengine.OnRecord(domain.SnippetDeleted{}, metaengine.Remove[readmodels.SnippetView]()),
		metaengine.FilterOnField[readmodels.SnippetView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.SnippetView]("created_at", false),
	)
}

// snippetKeyInput is the projection's key-extraction input (the same
// shape the readmodels engine declares).
type snippetKeyInput struct{ ID string }

// SnippetsDomainDecl is the snippets aggregate's share of the system's
// DomainConfig (a registration failure is a boot-time miswire and
// panics loudly).
func SnippetsDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterSnippets(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register snippets aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(snippetsProjection())},
		Events: []event.Type{
			event.Type(metaengine.EventTypeName(domain.SnippetSaved{})),
			event.Type(metaengine.EventTypeName(domain.SnippetDeleted{})),
		},
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// Snippets is the dispatch-backed reply-snippet service: the same
// contract the engine-backed store exposed (Save/List/Delete; the cap
// and replace-keeps-stamp semantics answer from the decision).
type Snippets struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.SnippetView]
	vers   streamVersions
	mu     sync.Mutex
}

// NewSnippets builds the backend over a System.
func NewSnippets(sys *system.System) *Snippets {
	return &Snippets{sys: sys, reader: metaengine.NewReader[readmodels.SnippetView](sys.MetaEngine(), "snippets"), vers: streamVersions{}}
}

// awaitOwnWrite waits until the snippet projection has folded the
// just-dispatched batch (D7): the Settings partial re-renders from the
// projection in the SAME request. Best-effort — a tail-read failure
// WARNs (the write committed) and never fails the request.
func (s *Snippets) awaitOwnWrite(ctx context.Context, owner string) {
	awaitOwnWrite(ctx, s.sys, SnippetsStreamType, owner, "snippets", s.vers)
}

// Save upserts a snippet by id (the decision preserves the slot's
// creation time and enforces the per-owner cap for NEW ids).
func (s *Snippets) Save(ctx context.Context, snippet domain.Snippet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := dispatchOwnerCommand(ctx, s.sys, "snippet.save", snippet.Owner.String(),
		func(base *command.BasicCommand) *SaveSnippet {
			return &SaveSnippet{BasicCommand: base, Snippet: snippet}
		})
	if err != nil {
		if errors.Is(err, readmodels.ErrSnippetListFull) {
			return err
		}
		return errorfamily.WrapInfrastructuref(err, "cqrs.snippet_save", "dispatch snippet.save")
	}
	s.awaitOwnWrite(ctx, snippet.Owner.String())
	return nil
}

// List returns the owner's snippets, oldest first.
func (s *Snippets) List(ctx context.Context, owner domain.Extension) ([]domain.Snippet, error) {
	// WithLimit is load-bearing (the planner's default ceiling is 100):
	// the whole list renders and the cap check needs every row.
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", false),
		metaengine.WithLimit(snippetsMaxSlots+1),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.snippet_list", "list snippets for %s", owner.String())
	}
	out := make([]domain.Snippet, 0, len(views))
	for _, v := range views {
		out = append(out, domain.Snippet{
			ID:        domain.MustSnippetID(v.ID),
			Owner:     domain.MustParseExtension(v.Owner),
			Body:      v.Body,
			Quick:     v.Quick,
			CreatedAt: time.Unix(0, v.CreatedAt),
		})
	}
	return out, nil
}

// Delete removes one snippet scoped to its owner (a miss answers the
// rejection; the handler's 404 mapping is unchanged).
func (s *Snippets) Delete(ctx context.Context, owner domain.Extension, snippetID domain.SnippetID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "snippet.delete", owner.String(),
		func(base *command.BasicCommand) *DeleteSnippet {
			return &DeleteSnippet{BasicCommand: base, SnippetID: snippetID}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, owner.String())
	return nil
}
