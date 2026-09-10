# Release Process

## Overview

certwiz uses GitHub Actions and GoReleaser to build and publish releases. Pushing a tag matching `v*` triggers `.github/workflows/goreleaser.yml`, which runs `goreleaser release --clean` (goreleaser-action v7, GoReleaser v2) on a stable Go toolchain and publishes a GitHub Release.

## Published Artifacts

Archives are named `cert-<os>-<arch>.tar.gz` (`.zip` for Windows) and each contains the `cert` binary plus README.md, CHANGELOG.md, and LICENSE. The naming is defined in `.goreleaser.yml`: `amd64` is published as `x86_64` and `386` as `i386`.

| Archive | Platform |
|---------|----------|
| `cert-darwin-arm64.tar.gz` | macOS, Apple Silicon |
| `cert-darwin-x86_64.tar.gz` | macOS, Intel |
| `cert-linux-x86_64.tar.gz` | Linux, 64-bit x86 |
| `cert-linux-arm64.tar.gz` | Linux, 64-bit ARM |
| `cert-linux-armv7.tar.gz` | Linux, 32-bit ARM |
| `cert-linux-i386.tar.gz` | Linux, 32-bit x86 |
| `cert-freebsd-x86_64.tar.gz` | FreeBSD, 64-bit x86 |
| `cert-freebsd-arm64.tar.gz` | FreeBSD, 64-bit ARM |
| `cert-windows-x86_64.zip` | Windows, 64-bit x86 |
| `cert-windows-arm64.zip` | Windows, ARM64 |
| `cert-windows-i386.zip` | Windows, 32-bit x86 |
| `checksums.txt` | SHA-256 sums for every archive |

The version string is injected at build time with `-X certwiz/cmd.version={{.Version}}`, so the binary reports the tag's version without the `v` prefix.

## Creating a Release

### 1. Update the version

Edit `cmd/root.go`:
```go
var version = "0.4.1"
```

### 2. Update CHANGELOG.md

Move the entries under `## [Unreleased]` into a new dated section, leave `## [Unreleased]` empty above it, and add the release link at the bottom of the file:

```markdown
## [Unreleased]

## [0.4.1] - 2026-09-09

### Fixed
- ...
```

```markdown
[0.4.1]: https://github.com/trahma/certwiz/releases/tag/v0.4.1
```

### 3. Run the checks CI runs

```bash
gofmt -l .
go vet ./...
go test -race -count=3 ./...
GOFLAGS=-mod=mod go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout=5m ./...
```

### 4. Commit

```bash
git add cmd/root.go CHANGELOG.md
git commit -m "chore: bump version to 0.4.1"
```

### 5. Tag and push

Use an annotated tag and push the branch before the tag so the Test workflow runs on the commit as well:

```bash
git tag -a v0.4.1 -m "v0.4.1"
git push origin main
git push origin v0.4.1
```

### 6. Watch the build

```bash
gh run list --limit 2
gh run watch <run-id> --exit-status
```

Two runs start: `Test` on the push to `main` and `GoReleaser` on the tag. The GoReleaser run builds every platform, creates the archives and `checksums.txt`, and publishes the release with GitHub-generated notes grouped by commit prefix (`feat:`, `fix:`, `perf:`, `security:`; `docs:`, `test:`, and `chore:` commits are excluded).

### 7. Verify the release

```bash
gh release view v0.4.1

mkdir -p /tmp/cert-check && cd /tmp/cert-check
gh release download v0.4.1 --pattern 'cert-darwin-arm64.tar.gz'
tar xzf cert-darwin-arm64.tar.gz
./cert version
# cert version 0.4.1
```

Then follow the installation instructions for your platform to confirm they still work.

## Local Release Testing

```bash
# Cross-compile every platform into dist/ without GoReleaser
make build-all

# Dry-run the GoReleaser pipeline (requires goreleaser installed)
make release-test      # snapshot build, nothing published
make release-local     # snapshot build, skips publishing
```

`make build-all` uses its own file names and is only a build smoke test; the published names come from GoReleaser.

## Troubleshooting

### Build failures
- The release workflow uses `go-version: stable`; earlier toolchains omit `LC_UUID` from darwin binaries, which macOS 15.4+ refuses to load. Do not pin it lower.
- Run `go mod tidy` and confirm `go.mod` still declares `go 1.20`.
- Read the GoReleaser job log with `gh run view <run-id> --log`.

### Missing binaries
- Check the `ignore:` list in `.goreleaser.yml`; darwin/386, darwin/arm, windows/arm, freebsd/386, and freebsd/arm are deliberately skipped.

### Version mismatch
- The binary reports whatever the tag says. If `cert version` disagrees with `cmd/root.go`, the version bump was not committed before tagging.

## Security

- Binaries are built in GitHub Actions from the tagged commit with `CGO_ENABLED=0`.
- SHA-256 checksums are published in `checksums.txt`.
- Binaries are not signed.

## Future Enhancements

- Homebrew tap (the `brews:` section in `.goreleaser.yml` is commented out and needs a `HOMEBREW_TAP_GITHUB_TOKEN` secret)
- Snap package for Linux
- MSI installer for Windows
- Docker images
- Binary signing
