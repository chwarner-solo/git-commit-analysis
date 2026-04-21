package domain

import (
	"fmt"
	"os"
	"path/filepath"
)

// WorkTree is a validated, opaque handle to a git repository.
// If you hold one, it is real. There is no other way to construct one.
type WorkTree struct {
	path string
}

// NewWorkTree is the only constructor. Returns an error if the path
// is empty, does not exist, or is not a git repository.
func NewWorkTree(path string) (WorkTree, error) {
	if path == "" {
		return WorkTree{}, fmt.Errorf("path cannot be empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return WorkTree{}, fmt.Errorf("invalid path: %w", err)
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return WorkTree{}, fmt.Errorf("path does not exist: %s", abs)
	}

	if _, err := os.Stat(filepath.Join(abs, ".git")); os.IsNotExist(err) {
		return WorkTree{}, fmt.Errorf("not a git repository: %s", abs)
	}

	return WorkTree{path: abs}, nil
}

// Path returns the absolute path to the repository root.
// It is the only accessor — the field itself is unexported.
func (w WorkTree) Path() string {
	return w.path
}
