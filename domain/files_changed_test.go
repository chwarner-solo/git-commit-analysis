package domain_test

import (
	"testing"

	"github.com/chwarner-solo/git-code-analysis/domain"
)

func TestWorkTree_FilesChanged(t *testing.T) {
	repoDir := makeGitRepo(t)

	wt, err := domain.NewWorkTree(repoDir)
	if err != nil {
		t.Fatalf("NewWorkTree: %v", err)
	}

	commits, err := wt.CommitsAhead("main")
	if err != nil {
		t.Fatalf("CommitsAhead: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	tests := []struct {
		name         string
		commitIdx    int
		wantFile     string
		wantStatus   domain.FileStatus
		wantAdditions int
	}{
		{
			name:         "second commit adds test file",
			commitIdx:    0, // most recent
			wantFile:     "main_test.go",
			wantStatus:   domain.FileAdded,
			wantAdditions: 1,
		},
		{
			name:         "first commit adds main.go",
			commitIdx:    1,
			wantFile:     "main.go",
			wantStatus:   domain.FileAdded,
			wantAdditions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := wt.FilesChanged(commits[tt.commitIdx].SHA)
			if err != nil {
				t.Fatalf("FilesChanged: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("expected 1 file changed, got %d", len(files))
			}

			f := files[0]
			if f.Path != tt.wantFile {
				t.Errorf("Path = %q, want %q", f.Path, tt.wantFile)
			}
			if f.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", f.Status, tt.wantStatus)
			}
			if f.Additions != tt.wantAdditions {
				t.Errorf("Additions = %d, want %d", f.Additions, tt.wantAdditions)
			}
		})
	}
}

func TestWorkTree_FilesChanged_InvalidSHA(t *testing.T) {
	repoDir := makeGitRepo(t)

	wt, err := domain.NewWorkTree(repoDir)
	if err != nil {
		t.Fatalf("NewWorkTree: %v", err)
	}

	_, err = wt.FilesChanged("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err == nil {
		t.Error("expected error for invalid SHA, got nil")
	}
}
