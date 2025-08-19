# Release Process

## Overview

CertWiz uses GitHub Actions to automatically build and publish releases when a version tag is pushed.

## Supported Platforms

The following platforms are automatically built for each release:

### macOS
- `darwin-amd64` - Intel Macs
- `darwin-arm64` - Apple Silicon (M1/M2/M3)

### Linux
- `linux-amd64` - 64-bit x86
- `linux-arm64` - 64-bit ARM
- `linux-386` - 32-bit x86

### Windows
- `windows-amd64` - 64-bit x86
- `windows-arm64` - ARM64
- `windows-386` - 32-bit x86

### FreeBSD
- `freebsd-amd64` - 64-bit x86
- `freebsd-arm64` - 64-bit ARM

## Creating a Release

### 1. Update Version

Update the version in `cmd/root.go`:
```go
var version = "0.2.0"  // Update this
```

### 2. Update CHANGELOG

Update `CHANGELOG.md` with the new version and changes:
```markdown
## [0.2.0] - 2025-08-19
### Added
- New features...
### Fixed
- Bug fixes...
```

### 3. Commit Changes

```bash
git add -A
git commit -m "Release v0.2.0"
git push
```

### 4. Create and Push Tag

```bash
git tag v0.2.0
git push origin v0.2.0
```

This will trigger the GitHub Actions workflow that:
1. Builds binaries for all platforms
2. Creates archives (tar.gz for Unix, zip for Windows)
3. Generates SHA256 checksums
4. Creates a GitHub Release with all artifacts

## Manual Release Testing

### Test Multi-Platform Build Locally

```bash
# Build all platforms
make build-all

# Check the dist/ directory
ls -la dist/
```

### Test with GoReleaser (Optional)

If you have GoReleaser installed:

```bash
# Test release process without publishing
make release-test

# Create local release artifacts
make release-local
```

## Release Workflows

We have two release workflows available:

### 1. Standard Release (`release.yml`)
- Triggered on version tags (`v*`)
- Builds all platform binaries
- Creates GitHub release with artifacts
- Simple and straightforward

### 2. GoReleaser (`goreleaser.yml`)
- Uses GoReleaser for more advanced features
- Better changelog generation
- Homebrew formula support (optional)
- More customization options

## Verifying Releases

After a release is published:

1. Check the GitHub Releases page
2. Verify all platform binaries are present
3. Test installation instructions for your platform
4. Verify checksums match

```bash
# Download and verify on macOS
curl -L https://github.com/yourusername/certwiz/releases/download/v0.2.0/cert-darwin-arm64.tar.gz | tar xz
./cert-darwin-arm64 version
# Should output: cert version v0.2.0
```

## Troubleshooting

### Build Failures
- Check Go version compatibility (requires 1.20+)
- Verify all dependencies with `go mod tidy`
- Check GitHub Actions logs for detailed errors

### Missing Binaries
- Ensure all GOOS/GOARCH combinations are valid
- Check for platform-specific build constraints

### Version Mismatch
- Ensure version in code matches git tag
- The release workflow injects version at build time

## Security

- Binaries are built in GitHub's secure environment
- SHA256 checksums provided for verification
- Consider signing binaries with GPG (future enhancement)

## Future Enhancements

- [ ] Homebrew formula auto-update
- [ ] Snap package for Linux
- [ ] MSI installer for Windows
- [ ] Docker images
- [ ] Binary signing with GPG
- [ ] Automatic changelog generation from commits