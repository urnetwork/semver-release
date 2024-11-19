package latest

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {

	cfg := struct {
		skipNewline bool
	}{}

	return &cli.Command{
		Name:      "latest",
		Args:      true,
		ArgsUsage: "<repo-path>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "skip-newline",
				Aliases:     []string{"n"},
				Usage:       "skip newline",
				EnvVars:     []string{"SKIP_NEWLINE"},
				Destination: &cfg.skipNewline,
			},
		},
		Description: "Shows latest version",
		Action: func(c *cli.Context) error {

			repoPath := c.Args().First()
			if repoPath == "" {
				return fmt.Errorf("repo path is required")
			}

			dir, err := filepath.Abs(".")
			if err != nil {
				return fmt.Errorf("failed to get absolute path of the current dir: %w", err)
			}

			repoRoot, err := findRepositoryRoot(dir)
			if err != nil {
				return fmt.Errorf("failed to find repository root: %w", err)
			}

			repo, err := git.PlainOpen(repoRoot)
			if err != nil {
				return fmt.Errorf("failed to open git repo: %w", err)
			}

			wt, err := repo.Worktree()
			if err != nil {
				return fmt.Errorf("failed to get worktree: %w", err)
			}

			_, err = wt.Status()
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			tags, err := repo.Tags()
			if err != nil {
				return fmt.Errorf("failed to get tags: %w", err)
			}

			semverTags := semver.Collection{}
			err = tags.ForEach(func(r *plumbing.Reference) error {
				v, err := semver.NewVersion(r.Name().Short())
				if err == nil {
					semverTags = append(semverTags, v)
					return nil
				}
				log.Println("skipping tag", "tag", r.Name().Short(), "error", err)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to iterate tags: %w", err)
			}

			if len(semverTags) == 0 {
				semverTags = append(semverTags, semver.MustParse("v0.0.0"))
			}

			sort.Sort(semverTags)

			latestVersion := semverTags[len(semverTags)-1]

			fmt.Print(latestVersion)

			if !cfg.skipNewline {
				fmt.Println()
			}

			return nil

		},
	}
}

func findRepositoryRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no git repository found")
		}
		dir = parent
	}
}
