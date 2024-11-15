# Semver Release

A command-line tool to automatically manage semantic versioning tags in Git repositories.

## Features

- Automatically finds the latest semantic version tag
- Increments version numbers following semver rules
- Creates and pushes new Git tags
- Ensures clean working tree before creating releases
- Supports different release types (patch by default)

## Installation

```bash
go install github.com/urnetwork/semver-release@latest
````

## Usage

```bash
# Create a new patch release
semver-release

# Specify release type
semver-release --release-type patch
```

## Requirements

- Git repository
- Go 1.x or higher
- Write access to the repository for creating and pushing tags

