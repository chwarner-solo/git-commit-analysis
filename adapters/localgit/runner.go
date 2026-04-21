package localgit

import "github.com/chwarner-solo/git-code-analysis/domain"

// Runner executes git commands against a WorkTree.
// This is the adapter that bridges the domain to the local git CLI.
type Runner struct {
	workTree domain.WorkTree
}

// New creates a Runner bound to the given WorkTree.
func New(wt domain.WorkTree) *Runner {
	return &Runner{workTree: wt}
}
