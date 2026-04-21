package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/chwarner-solo/git-code-analysis/domain"
)

func main() {
	base := flag.String("base", "main", "base branch to compare against")
	repo := flag.String("repo", ".", "path to git repository")
	flag.Parse()

	wt, err := domain.NewWorkTree(*repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	commits, err := wt.CommitsAhead(*base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting commits: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Printf("No commits ahead of %q\n", *base)
		return
	}

	fmt.Printf("Commits ahead of %q: %d\n\n", *base, len(commits))

	for _, c := range commits {
		files, err := wt.FilesChanged(c.SHA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not get files for %s: %v\n", c.SHA[:7], err)
			continue
		}

		risk := domain.ScoreRisk(c, files)
		fmt.Printf("[%s] %s  %s\n", risk.Overall, c.SHA[:7], c.Subject)

		for _, f := range files {
			fmt.Printf("        %-12s +%-4d -%-4d  %s\n",
				f.Status, f.Additions, f.Deletions, f.Path)
		}

		for _, factor := range risk.Factors {
			fmt.Printf("        ! %s\n", factor.Description)
		}
		fmt.Println()
	}
}
