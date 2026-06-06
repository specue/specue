package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Repository persists and retrieves the context Store. An interface so the model
// stays free of I/O and a test can substitute an in-memory store; FileRepository is
// the production implementation.
type Repository interface {
	// Load returns the stored registry, or an empty Store on first use (nothing
	// persisted yet).
	Load() (Store, error)
	// Save writes the registry.
	Save(Store) error
}

// FileRepository stores the registry as JSON at Path. The path is supplied by the
// caller — deciding it (an env override, a default home) is the application edge's
// job, not the domain's, so every source of that configuration lives in one place.
type FileRepository struct {
	Path string
}

// NewFileRepository returns a repository writing to path.
func NewFileRepository(path string) *FileRepository {
	return &FileRepository{Path: path}
}

// Load reads the JSON registry; a missing file is an empty Store (first run), a
// malformed file is an error rather than a silent reset.
func (r *FileRepository) Load() (Store, error) {
	raw, err := os.ReadFile(r.Path)
	if os.IsNotExist(err) {
		return Store{}, nil
	}
	if err != nil {
		return Store{}, err
	}
	var s Store
	if err := json.Unmarshal(raw, &s); err != nil {
		return Store{}, fmt.Errorf("parse %s: %w", r.Path, err)
	}
	return s, nil
}

// Save writes the registry, creating the parent directory if needed. Persisting
// to disk is what lets a context outlive the process that created it.
//
//specue:req:create-context#survives-across-invocations
func (r *FileRepository) Save(s Store) error {
	if err := os.MkdirAll(filepath.Dir(r.Path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.Path, raw, 0o644)
}
