package domain_test

import (
	"testing"

	"github.com/chwarner-solo/git-code-analysis/domain"
)

func TestScoreRisk(t *testing.T) {
	tests := []struct {
		name         string
		commit       domain.Commit
		files        []domain.FileChange
		wantOverall  domain.RiskLevel
		wantFactors  []string // substrings expected in factor descriptions
	}{
		// --- LOW anchors ---
		{
			name:   "only test files changed",
			commit: commit("test: improve coverage"),
			files: []domain.FileChange{
				file("domain/risk_scorer_test.go", domain.FileModified, 20, 5),
				file("domain/worktree_test.go", domain.FileModified, 10, 2),
			},
			wantOverall: domain.RiskLow,
			wantFactors: []string{"only test files"},
		},

		// --- HIGH anchors ---
		{
			name:   "env file modified",
			commit: commit("chore: update config"),
			files: []domain.FileChange{
				file(".env", domain.FileModified, 1, 1),
			},
			wantOverall: domain.RiskHigh,
			wantFactors: []string{"env/config file"},
		},
		{
			name:   "secrets file modified",
			commit: commit("chore: rotate secrets"),
			files: []domain.FileChange{
				file("config/secrets.yml", domain.FileModified, 2, 2),
			},
			wantOverall: domain.RiskHigh,
			wantFactors: []string{"env/config file"},
		},
		{
			name:   "database migration added",
			commit: commit("feat: add users table"),
			files: []domain.FileChange{
				file("migrations/20240101_add_users.sql", domain.FileAdded, 15, 0),
			},
			wantOverall: domain.RiskHigh,
			wantFactors: []string{"database migration"},
		},
		{
			name:   "wip in subject",
			commit: commit("WIP: half-done refactor"),
			files: []domain.FileChange{
				file("domain/worktree.go", domain.FileModified, 5, 3),
			},
			wantOverall: domain.RiskHigh,
			wantFactors: []string{"WIP/hack/temp"},
		},
		{
			name:   "hack in subject",
			commit: commit("hack: quick fix for prod"),
			files: []domain.FileChange{
				file("main.go", domain.FileModified, 3, 1),
			},
			wantOverall: domain.RiskHigh,
			wantFactors: []string{"WIP/hack/temp"},
		},

		// --- MEDIUM: blast radius ---
		{
			name:   "many files changed",
			commit: commit("refactor: reorganise packages"),
			files:  manyFiles(12),
			wantOverall: domain.RiskMedium,
			wantFactors: []string{"blast radius"},
		},
		{
			name:   "high churn volume",
			commit: commit("feat: large new feature"),
			files: []domain.FileChange{
				file("service/handler.go", domain.FileAdded, 400, 0),
				file("service/handler_test.go", domain.FileAdded, 200, 0),
			},
			wantOverall: domain.RiskMedium,
			wantFactors: []string{"high churn"},
		},

		// --- Risk elevation: no tests alongside source ---
		{
			name:   "source changed without tests",
			commit: commit("fix: correct off-by-one"),
			files: []domain.FileChange{
				file("domain/worktree.go", domain.FileModified, 5, 3),
				file("domain/commit.go", domain.FileModified, 2, 1),
			},
			wantOverall: domain.RiskMedium,
			wantFactors: []string{"no test files"},
		},

		// --- Clean commit: source + tests, small diff ---
		{
			name:   "source and tests changed small diff",
			commit: commit("feat: add Path accessor"),
			files: []domain.FileChange{
				file("domain/worktree.go", domain.FileModified, 5, 0),
				file("domain/worktree_test.go", domain.FileModified, 10, 0),
			},
			wantOverall: domain.RiskLow,
			wantFactors: []string{"tests included"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ScoreRisk(tt.commit, tt.files)

			if got.Overall != tt.wantOverall {
				t.Errorf("Overall = %q, want %q", got.Overall, tt.wantOverall)
			}

			for _, want := range tt.wantFactors {
				if !anyFactorContains(got.Factors, want) {
					t.Errorf("expected a factor containing %q, got: %v", want, got.Factors)
				}
			}
		})
	}
}

// --- helpers ---

func commit(subject string) domain.Commit {
	return domain.Commit{Subject: subject}
}

func file(path string, status domain.FileStatus, add, del int) domain.FileChange {
	return domain.FileChange{Path: path, Status: status, Additions: add, Deletions: del}
}

func manyFiles(n int) []domain.FileChange {
	files := make([]domain.FileChange, n)
	for i := range files {
		files[i] = file("pkg/file.go", domain.FileModified, 5, 2)
	}
	return files
}

func anyFactorContains(factors []domain.RiskFactor, substr string) bool {
	for _, f := range factors {
		if contains(f.Description, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && indexOfSubstr(s, substr) >= 0)
}

func indexOfSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
