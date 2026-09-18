// Package blob stores attachment and fax document files on disk under the
// data directory. Files are named by fresh nanoids (never by user input);
// the database owns the mapping id → path.
package blob

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sixafter/nanoid"
)

// Store writes and reads files under one root directory.
type Store struct {
	root string
}

// New creates the store, making the root directory if needed.
func New(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create blob root %s: %w", root, err)
	}
	return &Store{root: root}, nil
}

// Save writes content under subDir with a generated name and the given
// extension (".pdf", ".jpg", ...). It returns the path relative to the
// store root — what the database persists.
func (s *Store) Save(subDir, ext string, content []byte) (string, error) {
	if subDir != "" {
		if err := os.MkdirAll(filepath.Join(s.root, subDir), 0o750); err != nil {
			return "", fmt.Errorf("create %s: %w", subDir, err)
		}
	}
	name, err := nanoid.New()
	if err != nil {
		return "", fmt.Errorf("generate file name: %w", err)
	}
	rel := filepath.Join(subDir, string(name)+ext)
	abs := filepath.Join(s.root, rel)

	if err := os.WriteFile(abs, content, 0o640); err != nil {
		return "", fmt.Errorf("write %s: %w", rel, err)
	}

	return rel, nil
}

// Open returns a reader for a relative path previously handed out by Save.
// It refuses paths that try to escape the root.
func (s *Store) Open(rel string) (io.ReadSeekCloser, error) {
	clean := filepath.Clean(rel)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("refusing path %q outside store", rel)
	}
	file, err := os.Open(filepath.Join(s.root, clean))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", clean, err)
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
