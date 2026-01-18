# Releasing Milhouse

## Overview

Milhouse uses an automated release process powered by GitHub Actions and GoReleaser.

## Release Process

### 1. Prepare the Release

1. Update version in `Makefile`:
   ```makefile
   VERSION ?= X.Y.Z
   ```

2. Make your changes and commit:
   ```bash
   git add -A
   git commit -m "feat: description of changes

   Bump version to X.Y.Z"
   git push origin main
   ```

### 2. Create and Push Tag

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

Or use the Makefile helper:
```bash
make tag VERSION=X.Y.Z
```

### 3. Automated Release (GitHub Actions)

When a `v*` tag is pushed, GitHub Actions automatically:

1. **Runs tests** - `go test -v ./...`
2. **Builds binaries** via GoReleaser:
   - macOS (darwin): amd64, arm64
   - Linux: amd64, arm64
3. **Creates GitHub release** with:
   - Source archives (tar.gz, zip)
   - Binary archives for each platform
   - SHA256 checksums
4. **Updates Homebrew tap** (`daydemir/homebrew-tap`)

### 4. Verify Release

```bash
# Check GitHub Actions status
# https://github.com/daydemir/milhouse/actions

# After workflow completes:
brew update
brew upgrade milhouse
mil version
```

## Version Numbering

Follow semantic versioning (SemVer):
- **MAJOR** (X): Breaking changes
- **MINOR** (Y): New features, backwards compatible
- **PATCH** (Z): Bug fixes, backwards compatible

## Configuration Files

- `.goreleaser.yaml` - GoReleaser build configuration
- `.github/workflows/release.yml` - GitHub Actions workflow
- `Makefile` - Version definition and build commands

## Secrets Required

The release workflow requires these GitHub secrets:
- `GITHUB_TOKEN` - Auto-provided by GitHub Actions
- `HOMEBREW_TAP_GITHUB_TOKEN` - PAT with write access to `daydemir/homebrew-tap`
