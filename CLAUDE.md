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
│   └── verify.go         # cert verify
├── pkg/                   # Core packages
│   ├── cert/             # Certificate operations
│   │   └── cert.go
│   └── ui/               # Terminal UI with lipgloss
│       └── ui.go
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
- `inspect` - View certificate details
- `generate` - Create self-signed certificates
- `convert` - Convert between PEM/DER formats
- `verify` - Validate certificates

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

// From URL
cert, chain, err := cert.InspectURLWithChain(url, port)
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

## Testing Guidelines

### Manual Testing Commands
```bash
# Basic inspection
./cert inspect google.com

# Full details with chain
./cert inspect google.com --full --chain

# Generate certificate
./cert generate --cn test.local --san test.local

# Convert format
./cert convert test.pem test.der --format der

# Verify certificate
./cert verify test.crt --host test.local
```

### Common Test Domains
- google.com (many SANs)
- github.com (standard setup)
- expired.badssl.com (expired cert)
- self-signed.badssl.com (self-signed)

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

## Future Enhancements (Roadmap)

These are planned but not yet implemented:
- ECDSA key generation
- CA certificate generation
- PKCS#12/PFX support
- JSON output format
- Certificate signing requests (CSR)
- ACME/Let's Encrypt integration
- Certificate transparency logs
- Web UI dashboard

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