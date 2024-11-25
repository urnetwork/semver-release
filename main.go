package main

import (
	"github.com/urfave/cli/v2"
	"github.com/urnetwork/semver-release/latest"
	"github.com/urnetwork/semver-release/release"
	release_needed "github.com/urnetwork/semver-release/release-needed"
)

func main() {

	app := &cli.App{
		Name: "semver-release",

		Commands: []*cli.Command{
			latest.Command(),
			release.Command(),
			release_needed.Command(),
		},
	}
	app.RunAndExitOnError()
}
