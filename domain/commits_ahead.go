package domain

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// commitDelimiter separates commits in git log output.
const commitDelimiter = "---COMMIT---"

// CommitsAhead returns all commits reachable from HEAD that are not
// reachable from base, most recent first.
func (w WorkTree) CommitsAhead(base string) ([]Commit, error) {
	// Format: SHA | author name | author email | authored date (unix) | subject | body
	// %x00 is null byte — safe delimiter within a field
	format := strings.Join([]string{
		"%H",  // full SHA
		"%an", // author name
		"%ae", // author email
		"%at", // author date unix timestamp
		"%s",  // subject
		"%b",  // body
	}, "%x1F") // unit separator between fields

	// Each commit block is terminated by our delimiter on its own line
	fullFormat := fmt.Sprintf("--format=%s%%n%s", format, commitDelimiter)

	cmd := exec.Command("git", "log", fmt.Sprintf("%s..HEAD", base), fullFormat)
	cmd.Dir = w.path

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseCommits(strings.TrimSpace(string(out)))
}

// parseCommits splits raw git log output into Commit values.
func parseCommits(raw string) ([]Commit, error) {
	if raw == "" {
		return nil, nil
	}

	var commits []Commit
	blocks := strings.Split(raw, commitDelimiter)

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		c, err := parseCommitBlock(block)
		if err != nil {
			return nil, err
		}
		commits = append(commits, c)
	}

	return commits, nil
}

// parseCommitBlock parses a single commit block from git log output.
func parseCommitBlock(block string) (Commit, error) {
	// The first line is our formatted fields; remaining lines are overflow
	// from the body (git appends body lines after the format line)
	lines := strings.SplitN(block, "\n", 2)
	fields := strings.Split(lines[0], "\x1F")

	if len(fields) < 6 {
		return Commit{}, fmt.Errorf("unexpected commit format: %q", lines[0])
	}

	ts, err := parseUnixTimestamp(fields[3])
	if err != nil {
		return Commit{}, fmt.Errorf("parsing timestamp: %w", err)
	}

	body := strings.TrimSpace(fields[5])
	// Append any extra body lines git emitted after the format line
	if len(lines) > 1 {
		extra := strings.TrimSpace(lines[1])
		if extra != "" {
			if body != "" {
				body += "\n" + extra
			} else {
				body = extra
			}
		}
	}

	return Commit{
		SHA:        fields[0],
		Author:     Identity{Name: fields[1], Email: fields[2]},
		AuthoredAt: ts,
		Subject:    fields[4],
		Body:       body,
	}, nil
}

func parseUnixTimestamp(s string) (time.Time, error) {
	var unix int64
	_, err := fmt.Sscanf(s, "%d", &unix)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid unix timestamp %q: %w", s, err)
	}
	return time.Unix(unix, 0), nil
}
