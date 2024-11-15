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

			status, err := wt.Status()
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			if !status.IsClean() {
				return fmt.Errorf("%s\nworking tree is not clean, please commit changes", status.String())
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

			// 7. Increment the version
			sort.Sort(semverTags)

			latestVersion := semverTags[len(semverTags)-1]
			fmt.Println("Latest version:", latestVersion)

			nextVersion := latestVersion.IncPatch()

			head, err := repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get head: %w", err)
			}

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
