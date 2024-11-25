package release_needed

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

	return &cli.Command{
		Flags: []cli.Flag{},
		Name:  "release-needed",
		Action: func(c *cli.Context) error {

			repoPath := c.Args().First()
			if repoPath == "" {
				repoPath = "."
			}

			dir, err := filepath.Abs(repoPath)
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

			st, err := wt.Status()
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			if !st.IsClean() {
				return fmt.Errorf("working directory is not clean:\n%v", st)
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

			head, err := repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get head: %w", err)
			}

			headCommit, err := repo.CommitObject(head.Hash())
			if err != nil {
				return fmt.Errorf("failed to get head commit: %w", err)
			}

			tagRef, err := repo.Tag("v" + latestVersion.String())
			switch err {
			case git.ErrTagNotFound:
				// this is an initial release
				fmt.Println("true")
				return nil
			case nil:
				tagCommitHash := tagRef.Hash()

				tagObject, err := repo.TagObject(tagRef.Hash())
				switch err {
				case plumbing.ErrObjectNotFound:
					// tagObject is not a tag, it's a commit
				case nil:
					tagCommitHash = tagObject.Target
				default:
					return fmt.Errorf("failed to get tag object: %w", err)
				}

				tagCommit, err := repo.CommitObject(tagCommitHash)
				if err != nil {
					return fmt.Errorf("failed to get tag commit: %w", err)
				}

				if tagCommit.Hash == headCommit.Hash {
					// no changes since last release
					fmt.Println("false")
					return nil
				}

				if len(tagCommit.ParentHashes) == 1 && tagCommit.ParentHashes[0] == headCommit.Hash {
					// no changes since last release, release added a new commit
					fmt.Println("false")
					return nil
				}

				fmt.Println("true")
				return nil
			default:
				return fmt.Errorf("failed to get tag: %w", err)
			}

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
