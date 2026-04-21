package domain

import "time"

// Commit represents a single git commit in the branch diff.
type Commit struct {
	SHA       string
	Subject   string
	Body      string
	Author    Identity
	AuthoredAt time.Time
	ParentSHAs []string
}

// IsMerge returns true if this commit has more than one parent.
func (c Commit) IsMerge() bool {
	return len(c.ParentSHAs) > 1
}

// Identity is a git author or committer.
type Identity struct {
	Name  string
	Email string
}

// FileChange describes a single file affected by a commit.
type FileChange struct {
	Path         string
	PreviousPath string // non-empty on renames
	Status       FileStatus
	Additions    int
	Deletions    int
}

// FileStatus mirrors git's single-letter status codes.
type FileStatus string

const (
	FileAdded    FileStatus = "added"
	FileModified FileStatus = "modified"
	FileDeleted  FileStatus = "deleted"
	FileRenamed  FileStatus = "renamed"
	FileCopied   FileStatus = "copied"
)
