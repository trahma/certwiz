# CLAUDE.md - AI Assistant Context

This file provides context for AI assistants (like Claude, ChatGPT, or GitHub Copilot) when working with the certwiz project.

## Project Overview

**Project Name**: certwiz  
**Binary Name**: cert  
**Purpose**: A user-friendly CLI tool for certificate management, similar to HTTPie but for certificates.

### Important Naming Convention
- The **project** is called `certwiz`
- The **binary/command** is called `cert`
- Documentation should refer to the command as `cert`
- The repository and package remain `certwiz`

## Project Structure

```
certwiz/                    # Project root (name: certwiz)
├── main.go                 # Entry point
├── Makefile               # Builds binary as 'cert'
├── go.mod                 # Module: certwiz
├── cmd/                   # CLI commands
│   ├── root.go           # Root command (Use: "cert")
│   ├── inspect.go        # cert inspect
│   ├── generate.go       # cert generate
│   ├── convert.go        # cert convert
│   ├── verify.go         # cert verify
│   ├── ca.go             # cert ca
│   ├── csr.go            # cert csr
│   ├── sign.go           # cert sign
│   ├── update.go         # cert update
│   └── helpers.go        # Shared helper functions
├── pkg/                   # Core packages
│   ├── cert/             # Certificate operations
│   │   ├── cert.go      # Main certificate functions
│   │   ├── json.go      # JSON output structures
│   │   └── san.go       # SAN parsing utilities
│   └── ui/               # Terminal UI with lipgloss
│       └── ui.go
├── internal/             # Internal packages
│   ├── environ/          # Environment detection
│   │   └── env.go       # CI and Unicode detection
│   └── testutil/         # Test utilities
└── docs/                  # Documentation
    ├── installation.md
    ├── usage.md
    ├── commands.md
    ├── examples.md
    ├── contributing.md
    └── faq.md
```

## Key Technical Details

### Language & Dependencies
- **Language**: Go 1.20+
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **UI Library**: Lipgloss (github.com/charmbracelet/lipgloss)
- **Certificate Handling**: Go standard library (crypto/x509, crypto/tls)

### Build Commands
```bash
make build        # Creates ./cert binary
make install      # Installs to $GOPATH/bin
make clean        # Removes build artifacts
make build-all    # Cross-platform builds
```

### Command Structure
All commands follow this pattern:
```bash
cert [command] [target] [flags]
```

Commands:
- `inspect` - View certificate details from files or URLs
- `generate` - Create self-signed certificates
- `convert` - Convert between PEM/DER formats
- `verify` - Validate certificates
- `ca` - Create Certificate Authority certificates
- `csr` - Generate Certificate Signing Requests
- `sign` - Sign CSRs with a CA certificate
- `update` - Update cert to the latest version
- `version` - Show version information
- `completion` - Generate shell completion scripts

## Code Style Guidelines

### Go Code
- Follow standard Go idioms
- Use meaningful variable names
- Keep functions small and focused
- Handle errors explicitly
- Add comments for complex logic

### UI/UX Principles
- **Colors**: Green (valid), Yellow (warning), Red (error), Blue (info)
- **Output**: Beautiful, readable, terminal-width aware
- **Defaults**: Smart defaults that just work
- **Errors**: Clear, actionable error messages

### Terminal Output
- Uses bordered tables for certificate info
- Smart text wrapping based on terminal width
- Color coding for status indicators
- Icons: ✓ (success), ✗ (failure), → (detail), 🔗 (URL)

## Common Tasks for AI Assistants

### Adding a New Command
1. Create `cmd/newcommand.go`
2. Define cobra.Command structure
3. Implement command logic
4. Add to rootCmd in init()
5. Update documentation

### Adding Certificate Features
1. Extend `pkg/cert/cert.go`
2. Add UI support in `pkg/ui/ui.go`
3. Wire up in appropriate command
4. Add tests
5. Update docs

### Updating Documentation
- Main README uses `cert` command in examples
- Docs use `cert` command throughout
- Keep project name as `certwiz` in descriptions

## Important Patterns

### Certificate Inspection
```go
// From file
cert, err := cert.InspectFile(filepath)

// From URL with basic options
cert, chain, err := cert.InspectURLWithChain(url, port)

// With proxy/tunnel support
cert, chain, err := cert.InspectURLWithConnect(url, port, connectHost)

// With timeout
cert, chain, err := cert.InspectURLWithConnectTimeout(url, port, connectHost, timeout)

// With signature algorithm preference (ECDSA/RSA)
cert, chain, err := cert.InspectURLWithOptions(url, port, connectHost, timeout, sigAlg)
```

### UI Display
```go
// Display certificate
ui.DisplayCertificate(cert, showFull)

// Display chain
ui.DisplayCertificateChain(chain)
```

### Error Handling
```go
if err != nil {
    ui.ShowError(err.Error())
    os.Exit(1)
}
```

## CI/CD and Linting

### Local Testing Before Push
**IMPORTANT**: Always test locally with the same tools CI uses before pushing changes.

```bash
# Run tests with race detector (as CI does)
go test -v -race ./...

# Run linting with latest golangci-lint
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run --timeout=5m

# Check specific Go version compatibility
docker run --rm -v $(pwd):/app -w /app golang:1.20 go test ./...
```

### Linting Requirements
- All error returns must be checked (use `_ =` if intentionally ignoring)
- Deferred Close() calls should handle errors: `defer func() { _ = conn.Close() }()`
- Use embedded struct fields directly when possible (avoid redundant field access)
- Follow golangci-lint default rules (no custom .golangci.yml needed)

### Common CI Issues & Solutions
1. **"Previous case" error in switch statements**: Missing import (e.g., crypto/ecdsa)
2. **"Could remove embedded field from selector"**: Use `cert.FieldName` instead of `cert.Certificate.FieldName`
3. **Unchecked errors**: Add `_ =` for intentionally ignored errors
4. **Test failures with -race**: Ensure no concurrent access to shared resources
5. **Missing test files**: Ensure all testdata files are committed to git (update .gitignore if needed)
6. **IP addresses in SANs**: Use `--san IP:192.168.1.1` format, not just `--san 192.168.1.1`

## Testing Guidelines

### CRITICAL: Always Run Tests Before Pushing
**The most common CI failures come from not running tests locally first!**

```bash
# ALWAYS run this exact command before pushing (same as CI):
go test -v -race -coverprofile coverage.out ./...

# If tests pass locally but fail in CI, try:
go clean -testcache  # Clear test cache
go mod tidy          # Ensure dependencies are correct
go test -v -race -coverprofile coverage.out ./...
```

### Debugging CI Test Failures

When CI tests fail but local tests pass:

1. **Check the exact error message in CI**
   - Syntax errors often indicate file corruption or encoding issues
   - "unexpected var after top level declaration" = likely extra characters at EOF
   - Build failures in one package can cascade to others

2. **Verify file integrity**
   ```bash
   # Check for hidden characters at end of file
   tail -5 cmd/update.go | od -c
   
   # Ensure file ends with single newline
   tail -c 10 cmd/update.go | xxd
   
   # Compare local vs GitHub version
   curl -s https://raw.githubusercontent.com/trahma/certwiz/main/cmd/update.go | diff - cmd/update.go
   ```

3. **Force a fresh CI build**
   ```bash
   # Add and remove a comment to trigger fresh build
   echo "// CI refresh" >> cmd/update.go
   git add cmd/update.go && git commit -m "Trigger CI rebuild"
   git push origin main
   
   # Then clean up
   git revert HEAD && git push origin main
   ```

4. **Check for platform-specific issues**
   ```bash
   # Test on different OS if possible
   docker run --rm -v $(pwd):/app -w /app golang:1.20-alpine go test ./...
   docker run --rm -v $(pwd):/app -w /app golang:1.20-bullseye go test ./...
   ```

### Test File Management

**Important**: Test data files must be committed to git!

```bash
# Check that testdata files are tracked
git ls-files testdata/

# If missing, ensure .gitignore allows them
# .gitignore should have:
!testdata/*.pem
!testdata/*.der
!testdata/*.crt
!testdata/*.key

# Add test files to git
git add testdata/*.pem testdata/*.der
git commit -m "Add test certificates"
```

### Path Issues in Tests

**Cross-platform path handling is critical!**

```go
// BAD - Will fail on Windows
file: "../../testdata/valid.pem"

// GOOD - Works everywhere
import "path/filepath"
file: filepath.Join("..", "..", "testdata", "valid.pem")

// BETTER - Use a helper function
func testdataPath(filename string) string {
    return filepath.Join("..", "..", "testdata", filename)
}
```

### Windows-Specific Test Issues

```bash
# Windows command line parsing issues
# BAD:
go test -coverprofile=coverage.out  # Windows may parse this incorrectly

# GOOD:
go test -coverprofile coverage.out  # Space instead of =
```

### Manual Testing Commands
```bash
# Basic inspection
./cert inspect google.com

# Full details with chain
./cert inspect google.com --full --chain

# Generate certificate (note: IP addresses need IP: prefix)
./cert generate --cn test.local --san test.local --san IP:192.168.1.1

# Convert format
./cert convert test.pem test.der --format der

# Verify certificate
./cert verify test.crt --host test.local

# Test update functionality
./cert update
./cert update --force
```

### Common Test Domains
- google.com (many SANs)
- github.com (standard setup)
- expired.badssl.com (expired cert)
- self-signed.badssl.com (self-signed)

### Running Tests Locally Before Push Checklist

1. ✅ Run full test suite: `go test -v -race -coverprofile coverage.out ./...`
2. ✅ Check for any modified files: `git status`
3. ✅ Ensure all new files are added: `git add .`
4. ✅ Verify builds cleanly: `go build -o cert .`
5. ✅ Test the binary: `./cert version`
6. ✅ If adding new commands, update test count in `cmd/root_test.go`

## Debugging Tips

### Build Issues
- Ensure Go 1.20+ is installed
- Run `go mod tidy` for dependencies
- Check `go.mod` for module name (certwiz)

### Display Issues
- Test with different terminal widths
- Check `$TERM` environment variable
- Test with `NO_COLOR=1` for color issues

### Certificate Issues
- Use `--full` flag for complete details
- Use `--chain` to see trust path
- Check SANs match hostname

## Completed Features

These features have been successfully implemented:
- ✅ CA certificate generation (`cert ca` command)
- ✅ Certificate signing requests (`cert csr` command)
- ✅ Certificate signing with CA (`cert sign` command)
- ✅ JSON output format (all commands support `--json` flag)
- ✅ Network timeout configuration (`--timeout` flag)
- ✅ Proxy/tunnel support (`--connect` flag)
- ✅ Signature algorithm selection (`--sig-alg` flag for inspect)
- ✅ Automatic update command (`cert update`)
- ✅ Certificate chain verification (`--ca` flag for verify)
- ✅ Secure key permissions (0600 on Unix systems)

## Future Enhancements (Roadmap)

These are planned but not yet implemented:
- ECDSA key generation (for generate command)
- PKCS#12/PFX support
- ACME/Let's Encrypt integration
- Certificate transparency logs
- Web UI dashboard
- OCSP stapling verification
- Certificate pinning validation
- Automatic certificate renewal
- Integration with HashiCorp Vault
- Kubernetes cert-manager integration

## Release Process

1. Update version in code
2. Run tests: `go test ./...`
3. Build all platforms: `make build-all`
4. Update CHANGELOG.md
5. Create git tag: `git tag vX.Y.Z`
6. Push tag: `git push origin vX.Y.Z`

## Common Issues & Solutions

### "command not found"
- Binary is named `cert`, not `certwiz`
- Check PATH includes install directory

### Colors not showing
- Terminal may not support colors
- Try `FORCE_COLOR=1 cert inspect ...`

### SANs wrapping incorrectly
- Check terminal width detection
- Verify lipgloss terminal detection

## Integration Points

### CI/CD
- Use `cert` command in scripts
- Exit codes: 0 (success), 1 (error)
- Parse output with grep/awk

### Docker
```dockerfile
FROM golang:alpine
WORKDIR /app
COPY . .
RUN go build -o cert .
ENTRYPOINT ["./cert"]
```

## Contributing

When contributing:
1. Binary must be named `cert`
2. Help text should show `cert` examples
3. Maintain backward compatibility
4. Update relevant documentation
5. Follow existing code style

## Contact & Support

- GitHub Issues: Bug reports and features
- Documentation: /docs directory
- Examples: /docs/examples.md

---

*This file helps AI assistants understand the project structure and conventions. Keep it updated as the project evolves.*