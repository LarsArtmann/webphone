// Package blob stores attachment and fax document files on disk under the
// data directory. Files are named by fresh nanoids (never by user input);
// the database owns the mapping id → path.
package blob

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/sixafter/nanoid"
)

// Store writes and reads files under one root directory.
type Store struct {
	root string
}

// New creates the store, making the root directory if needed.
func New(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "blob.root", "create blob root %s", root)
	}
	return &Store{root: root}, nil
}

// Save writes content under subDir with a generated name and the given
// extension (".pdf", ".jpg", ...). It returns the path relative to the
// store root — what the database persists — branded through
// ParseRelPath, so a malformed extension fails here instead of minting
// a path no later Open would accept.
func (s *Store) Save(subDir, ext string, content []byte) (domain.RelPath, error) { //nolint:branching-flow // literal args at both call sites; a swap fails immediately and visibly (triage 2026-10-11)
	if subDir != "" {
		if err := os.MkdirAll(filepath.Join(s.root, subDir), 0o750); err != nil {
			return domain.RelPath{}, errorfamily.WrapInfrastructuref(err, "blob.subdir", "create %s", subDir)
		}
	}
	name, err := nanoid.New()
	if err != nil {
		return domain.RelPath{}, errorfamily.WrapInfrastructuref(err, "blob.name", "generate file name")
	}
	rel := filepath.Join(subDir, string(name)+ext)
	abs := filepath.Join(s.root, rel)

	if err := os.WriteFile(abs, content, 0o640); err != nil {
		return domain.RelPath{}, errorfamily.WrapInfrastructuref(err, "blob.write", "write %s", rel)
	}

	branded, err := domain.ParseRelPath(filepath.ToSlash(rel))
	if err != nil {
		return domain.RelPath{}, errorfamily.WrapInfrastructuref(err, "blob.name", "brand stored path")
	}
	return branded, nil
}

// Open returns a reader for a relative path previously handed out by
// Save. The RelPath brand IS the escape guard: only ParseRelPath mints
// one and it refuses absolute and ".."-escaping shapes, so the join
// below cannot leave the root by construction.
func (s *Store) Open(rel domain.RelPath) (io.ReadSeekCloser, error) {
	file, err := os.Open(filepath.Join(s.root, rel.String()))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "blob.open", "open %s", rel)
	}
	return file, nil
}

// Abs resolves a relative path to its absolute location (for gateways that
// must read the file directly, e.g. the fax PDF upload).
func (s *Store) Abs(rel domain.RelPath) string {
	return filepath.Join(s.root, rel.String())
}

// Root returns the store root directory.
func (s *Store) Root() string { return s.root }

// Remove unlinks one stored blob. A missing file is NOT an error (the
// retention sweep may race a crashed earlier pass that already
// unlinked it); any other failure is.
func (s *Store) Remove(rel domain.RelPath) error {
	if err := os.Remove(s.Abs(rel)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errorfamily.WrapInfrastructuref(err, "blob.remove", "remove blob %s", rel)
	}
	return nil
}

// ProbeWrite proves the root accepts writes: a temp file is created,
// closed and removed again. A full disk or a lost/read-only mount
// fails here — the failure mode that would otherwise silently eat
// attachments. This is the ONE home of the blob write probe: the
// server's /healthz check and the container health check both call it.
func ProbeWrite(root string) error {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return errorfamily.WrapInfrastructuref(err, "blob.probe", "create blob root %s", root)
	}
	file, err := os.CreateTemp(root, ".healthz-*")
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "blob.probe", "write probe in %s", root)
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		return errorfamily.WrapInfrastructuref(err, "blob.probe", "close probe file")
	}
	if err := os.Remove(name); err != nil {
		return errorfamily.WrapInfrastructuref(err, "blob.probe", "remove probe file")
	}
	return nil
}

// HealthCheck answers the container's health question with the write
// probe: the store is stateless (no held file descriptors), so
// writability IS its health. Structural conformance to the container's
// HealthcheckerWithContext interface is asserted in internal/app.
func (s *Store) HealthCheck(_ context.Context) error { return ProbeWrite(s.root) }
