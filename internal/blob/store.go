// Package blob stores attachment and fax document files on disk under the
// data directory. Files are named by fresh nanoids (never by user input);
// the database owns the mapping id → path.
package blob

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/go-error-family"
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
// store root — what the database persists.
func (s *Store) Save(subDir, ext string, content []byte) (string, error) {
	if subDir != "" {
		if err := os.MkdirAll(filepath.Join(s.root, subDir), 0o750); err != nil {
			return "", errorfamily.WrapInfrastructuref(err, "blob.subdir", "create %s", subDir)
		}
	}
	name, err := nanoid.New()
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "blob.name", "generate file name")
	}
	rel := filepath.Join(subDir, string(name)+ext)
	abs := filepath.Join(s.root, rel)

	if err := os.WriteFile(abs, content, 0o640); err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "blob.write", "write %s", rel)
	}

	return rel, nil
}

// Open returns a reader for a relative path previously handed out by Save.
// It refuses paths that try to escape the root.
func (s *Store) Open(rel string) (io.ReadSeekCloser, error) {
	clean := filepath.Clean(rel)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, errorfamily.Newf(errorfamily.Rejection, "blob.escape", "refusing path %q outside store", rel)
	}
	file, err := os.Open(filepath.Join(s.root, clean))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "blob.open", "open %s", clean)
	}
	return file, nil
}

// Abs resolves a relative path to its absolute location (for gateways that
// must read the file directly, e.g. the fax PDF upload).
func (s *Store) Abs(rel string) string {
	return filepath.Join(s.root, filepath.Clean(rel))
}

// Root returns the store root directory.
func (s *Store) Root() string { return s.root }

// Remove unlinks one stored blob. A missing file is NOT an error (the
// retention sweep may race a crashed earlier pass that already
// unlinked it); any other failure is.
func (s *Store) Remove(rel string) error {
	if err := os.Remove(s.Abs(rel)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errorfamily.WrapInfrastructuref(err, "blob.remove", "remove blob %s", rel)
	}
	return nil
}
