package domain

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FilesChanged returns the files affected by a single commit SHA.
// Uses git diff-tree which is stable, plumbing-level, and safe to parse.
func (w WorkTree) FilesChanged(sha string) ([]FileChange, error) {
	// --no-commit-id: suppress the SHA header line
	// -r:             recurse into subtrees
	// --numstat:      machine-readable additions/deletions counts
	// --diff-filter:  all statuses
	// --name-status would give us rename info; numstat doesn't — use both passes
	cmd := exec.Command(
		"git", "diff-tree",
		"--no-commit-id", "-r",
		"--numstat",
		"--diff-filter=ACDMRT",
		sha,
	)
	cmd.Dir = w.path

	numstatOut, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff-tree numstat failed for %s: %w", sha, err)
	}

	// Second pass: get status letters (A/M/D/R/C) and rename paths
	cmd2 := exec.Command(
		"git", "diff-tree",
		"--no-commit-id", "-r",
		"--name-status",
		"--diff-filter=ACDMRT",
		sha,
	)
	cmd2.Dir = w.path

	namestatOut, err := cmd2.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff-tree name-status failed for %s: %w", sha, err)
	}

	return mergeFileChanges(
		strings.TrimSpace(string(numstatOut)),
		strings.TrimSpace(string(namestatOut)),
	)
}

// mergeFileChanges combines numstat (additions/deletions) with
// name-status (status letter + rename paths) into FileChange values.
func mergeFileChanges(numstat, namestatus string) ([]FileChange, error) {
	statMap, err := parseNumstat(numstat)
	if err != nil {
		return nil, err
	}
	return parseNameStatus(namestatus, statMap)
}

type fileStat struct{ additions, deletions int }

func parseNumstat(raw string) (map[string]fileStat, error) {
	result := map[string]fileStat{}
	if raw == "" {
		return result, nil
	}
	for _, line := range strings.Split(raw, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		add, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // binary files show "-" — skip for now
		}
		del, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		result[parts[2]] = fileStat{additions: add, deletions: del}
	}
	return result, nil
}

func parseNameStatus(raw string, stats map[string]fileStat) ([]FileChange, error) {
	if raw == "" {
		return nil, nil
	}
	var changes []FileChange
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		statusLetter := string(parts[0][0]) // R100 → R, M → M etc
		fc := FileChange{Status: gitStatusToFileStatus(statusLetter)}

		switch statusLetter {
		case "R", "C":
			if len(parts) < 3 {
				continue
			}
			fc.PreviousPath = parts[1]
			fc.Path = parts[2]
			if s, ok := stats[fc.Path]; ok {
				fc.Additions = s.additions
				fc.Deletions = s.deletions
			}
		default:
			fc.Path = parts[1]
			if s, ok := stats[fc.Path]; ok {
				fc.Additions = s.additions
				fc.Deletions = s.deletions
			}
		}

		changes = append(changes, fc)
	}
	return changes, nil
}

func gitStatusToFileStatus(letter string) FileStatus {
	switch letter {
	case "A":
		return FileAdded
	case "M":
		return FileModified
	case "D":
		return FileDeleted
	case "R":
		return FileRenamed
	case "C":
		return FileCopied
	default:
		return FileModified
	}
}
