package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chwarner-solo/git-code-analysis/domain"
)

func TestNewWorkTree(t *testing.T) {
	// Build a real temp dir with a .git folder for the happy path
	validRepo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(validRepo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid git repo",
			path:    validRepo,
			wantErr: false,
		},
		{
			name:    "empty path rejected",
			path:    "",
			wantErr: true,
		},
		{
			name:    "non-existent path rejected",
			path:    "/this/does/not/exist",
			wantErr: true,
		},
		{
			name:    "directory exists but not a git repo",
			path:    t.TempDir(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wt, err := domain.NewWorkTree(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (WorkTree path: %q)", wt.Path())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if wt.Path() == "" {
				t.Error("WorkTree.Path() returned empty string on valid repo")
			}
		})
	}
}
