package domain_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/chwarner-solo/git-code-analysis/domain"
)

// makeGitRepo creates a real git repo with commits for testing.
// Returns the path to the repo root.
func makeGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	// Initial commit on main
	writeFile(t, dir, "README.md", "# test repo")
	run("add", ".")
	run("commit", "-m", "chore: initial commit")

	// Create feature branch
	run("checkout", "-b", "feature/test")

	// Commit 1 on feature branch
	writeFile(t, dir, "main.go", "package main")
	run("add", ".")
	run("commit", "-m", "feat: add main package")

	// Commit 2 on feature branch
	writeFile(t, dir, "main_test.go", "package main_test")
	run("add", ".")
	run("commit", "-m", "test: add test file")

	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

func TestWorkTree_CommitsAhead(t *testing.T) {
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
		t.Fatalf("expected 2 commits ahead of main, got %d", len(commits))
	}

	// Most recent commit first
	if commits[0].Subject != "test: add test file" {
		t.Errorf("commits[0].Subject = %q, want %q", commits[0].Subject, "test: add test file")
	}
	if commits[1].Subject != "feat: add main package" {
		t.Errorf("commits[1].Subject = %q, want %q", commits[1].Subject, "feat: add main package")
	}

	// SHAs must be non-empty
	for i, c := range commits {
		if c.SHA == "" {
			t.Errorf("commits[%d].SHA is empty", i)
		}
		if c.Author.Name == "" {
			t.Errorf("commits[%d].Author.Name is empty", i)
		}
	}
}

func TestWorkTree_CommitsAhead_NoneAhead(t *testing.T) {
	repoDir := makeGitRepo(t)

	wt, err := domain.NewWorkTree(repoDir)
	if err != nil {
		t.Fatalf("NewWorkTree: %v", err)
	}

	// HEAD is ahead of feature/test by 0 commits relative to itself
	commits, err := wt.CommitsAhead("feature/test")
	if err != nil {
		t.Fatalf("CommitsAhead: %v", err)
	}

	if len(commits) != 0 {
		t.Errorf("expected 0 commits, got %d", len(commits))
	}
}
