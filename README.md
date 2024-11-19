# semver-release

A command-line tool for managing semantic versioning releases.

## Installation

```bash
go install github.com/urnetwork/semver-release@latest
```

## Usage

`semver-release` provides two main commands:

### Latest Version

```bash
semver-release latest <repo-path>
```

This command will output the latest semantic version tag in the repository.
You can instruct the command to skip printing the end of line character by using the `--skip-newline` / `-n` flag.

### Create Release
```bash
semver-release release
```

The `release` command creates a new semantic version tag based on the latest version. You can specify the type of version increment (`patch`, `minor`, `major`) using the `--type` flag:


## Requirements

- Git repository
- Go 1.x or higher
- Write access to the repository for creating and pushing tags
