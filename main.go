package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/urfave/cli/v2"
)

func main() {
	cfg := struct {
		releaseType string
	}{}
	app := &cli.App{
		Name: "semver-release",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "release-type",
				Usage:       "release type",
				EnvVars:     []string{"RELEASE_TYPE"},
				Destination: &cfg.releaseType,
				Value:       "patch",
			},
		},

		Action: func(c *cli.Context) error {

			// 1. Open the repo
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

			// if !status.IsClean() {
			// 	return fmt.Errorf("%s\nworking tree is not clean, please commit changes", status.String())
			// }

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
			fmt.Println("Latest version:", latestVersion)

			tagRef, err := repo.Tag("v" + latestVersion.String())
			if err != nil {
				return fmt.Errorf("failed to get tag: %w", err)
			}

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

			nextVersion := latestVersion.IncPatch()

			head, err := repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get head: %w", err)
			}

			tagCommit, err := repo.CommitObject(tagCommitHash)
			if err != nil {
				return fmt.Errorf("failed to get tag commit: %w", err)
			}

			headCommit, err := repo.CommitObject(head.Hash())
			if err != nil {
				return fmt.Errorf("failed to get head commit: %w", err)
			}

			if tagCommit.Hash == headCommit.Hash {
				fmt.Println("No changes since last release, nothing to tag")
				return nil
			}

			if len(tagCommit.ParentHashes) == 1 && tagCommit.ParentHashes[0] == headCommit.Hash {
				fmt.Println("No changes since last release, nothing to tag")
				return nil
			}

			// tagTree, err := tagCommit.Tree()
			// if err != nil {
			// 	return fmt.Errorf("failed to get tag tree: %w", err)
			// }

			// headTree, err := headCommit.Tree()
			// if err != nil {
			// 	return fmt.Errorf("failed to get head tree: %w", err)
			// }

			// diff, err := tagTree.Diff(headTree)

			// if err != nil {
			// 	return fmt.Errorf("failed to get diff: %w", err)
			// }

			// fmt.Println("head:", head.Hash())
			// fmt.Println("tag:", tagRef.Hash())
			// if diff.Len() == 0 {
			// 	return fmt.Errorf("no changes since last release")
			// }

			// for _, c := range diff {
			// 	fmt.Println(c.String())
			// }

			return nil

			_, err = repo.CreateTag("v"+nextVersion.String(), head.Hash(), &git.CreateTagOptions{
				Tagger: &object.Signature{
					Name:  "semver-release",
					Email: "noreply@bringyour.com",
					When:  time.Now(),
				},
				Message: "Release " + nextVersion.String(),
			})

			if err != nil {
				return fmt.Errorf("failed to create tag: %w", err)
			}

			err = repo.Push(&git.PushOptions{
				FollowTags: true,
			})

			if err != nil {
				return fmt.Errorf("failed to push tag: %w", err)
			}

			return nil

		},
	}
	app.RunAndExitOnError()
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
