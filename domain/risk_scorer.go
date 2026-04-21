package domain

import (
	"path/filepath"
	"strconv"
	"strings"
)

// ScoreRisk assesses the risk of a commit based on its subject and file changes.
// Rules are evaluated independently; the overall level is the maximum of all factors.
func ScoreRisk(c Commit, files []FileChange) RiskAssessment {
	var factors []RiskFactor

	factors = append(factors, scoreSubject(c)...)
	factors = append(factors, scoreFiles(files)...)
	factors = append(factors, scoreTestCoverage(files)...)
	factors = append(factors, scoreBlastRadius(files)...)
	factors = append(factors, scoreChurn(files)...)

	return RiskAssessment{
		Overall: maxLevel(factors),
		Factors: factors,
	}
}

// scoreSubject flags dangerous keywords in the commit subject line.
func scoreSubject(c Commit) []RiskFactor {
	var factors []RiskFactor
	lower := strings.ToLower(c.Subject)

	wipKeywords := []string{"wip", "hack", "temp", "fixup", "hotfix", "quick fix"}
	for _, kw := range wipKeywords {
		if strings.Contains(lower, kw) {
			factors = append(factors, RiskFactor{
				Description: "WIP/hack/temp keyword in commit subject",
				Level:       RiskHigh,
			})
			break
		}
	}

	return factors
}

// scoreFiles inspects individual files for high-risk patterns.
func scoreFiles(files []FileChange) []RiskFactor {
	var factors []RiskFactor

	for _, f := range files {
		if isConfigOrSecrets(f.Path) {
			factors = append(factors, RiskFactor{
				Description: "env/config file modified: " + f.Path,
				Level:       RiskHigh,
			})
		}
		if isMigration(f.Path) {
			factors = append(factors, RiskFactor{
				Description: "database migration file: " + f.Path,
				Level:       RiskHigh,
			})
		}
	}

	return factors
}

// scoreTestCoverage checks whether source changes are accompanied by tests.
func scoreTestCoverage(files []FileChange) []RiskFactor {
	if len(files) == 0 {
		return nil
	}

	hasSource := false
	hasTests := false

	for _, f := range files {
		if isTestFile(f.Path) {
			hasTests = true
		} else {
			hasSource = true
		}
	}

	if !hasSource && hasTests {
		return []RiskFactor{{
			Description: "only test files changed",
			Level:       RiskLow,
		}}
	}

	if hasSource && hasTests {
		return []RiskFactor{{
			Description: "tests included alongside source changes",
			Level:       RiskLow,
		}}
	}

	if hasSource && !hasTests {
		return []RiskFactor{{
			Description: "no test files changed alongside source",
			Level:       RiskMedium,
		}}
	}

	return nil
}

// scoreBlastRadius flags commits touching many files.
func scoreBlastRadius(files []FileChange) []RiskFactor {
	const threshold = 10
	if len(files) > threshold {
		return []RiskFactor{{
			Description: "blast radius: " + itoa(len(files)) + " files changed",
			Level:       RiskMedium,
		}}
	}
	return nil
}

// scoreChurn flags commits with large total line changes.
func scoreChurn(files []FileChange) []RiskFactor {
	const threshold = 500
	total := 0
	for _, f := range files {
		total += f.Additions + f.Deletions
	}
	if total > threshold {
		return []RiskFactor{{
			Description: "high churn: " + itoa(total) + " lines changed",
			Level:       RiskMedium,
		}}
	}
	return nil
}

// --- path classifiers ---

func isTestFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return strings.HasSuffix(strings.TrimSuffix(base, ext), "_test")
	case ".ts", ".js":
		return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.")
	}
	return strings.Contains(path, "/test/") || strings.Contains(path, "/tests/")
}

func isConfigOrSecrets(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	sensitiveNames := []string{".env", "secrets", "credentials", "auth"}
	sensitiveExts := []string{".env", ".pem", ".key", ".p12", ".pfx"}

	for _, name := range sensitiveNames {
		if strings.Contains(base, name) {
			return true
		}
	}
	for _, ext := range sensitiveExts {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}

	configExts := []string{".yml", ".yaml", ".toml", ".ini", ".json"}
	configDirs := []string{"config/", "configs/", "settings/"}
	for _, ext := range configExts {
		if strings.HasSuffix(base, ext) {
			for _, dir := range configDirs {
				if strings.Contains(strings.ToLower(path), dir) {
					return true
				}
			}
		}
	}
	return false
}

func isMigration(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "migration") ||
		strings.Contains(lower, "migrate") ||
		(strings.Contains(lower, "/db/") && filepath.Ext(path) == ".sql")
}

// --- utilities ---

func maxLevel(factors []RiskFactor) RiskLevel {
	level := RiskLow
	for _, f := range factors {
		if riskWeight(f.Level) > riskWeight(level) {
			level = f.Level
		}
	}
	return level
}

func riskWeight(r RiskLevel) int {
	switch r {
	case RiskHigh:
		return 3
	case RiskMedium:
		return 2
	default:
		return 1
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
