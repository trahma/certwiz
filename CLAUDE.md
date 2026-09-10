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
├── go.mod                 # Module: certwiz (go 1.20)
├── cmd/                   # CLI commands
│   ├── root.go           # Root command (Use: "cert"), Execute() prints unreported errors
│   ├── inspect.go        # cert inspect (looksLikeFilePath decides file vs host)
│   ├── generate.go       # cert generate
│   ├── convert.go        # cert convert
│   ├── verify.go         # cert verify
│   ├── ca.go             # cert ca
│   ├── csr.go            # cert csr
│   ├── sign.go           # cert sign
│   ├── tls.go            # cert tls
│   ├── update.go         # cert update
│   ├── helpers.go        # printJSON, printJSONError, reportError
│   └── helpers_test.go   # Test fixtures: captureStdout, setOutputMode, startTLSServer,
│                         #   startSilentListener, resetCommandFlags, newTestCA/newTestLeaf
├── pkg/                   # Core packages
│   ├── cert/             # Certificate operations
│   │   ├── cert.go      # Inspect/generate/convert/verify/sign, TLS probing, PEM writers
│   │   ├── json.go      # JSON output structures and ToJSON/MarshalJSON
│   │   ├── san.go       # SAN parsing utilities
│   │   ├── usage.go     # Key-usage name tables shared by JSON and UI
│   │   └── refactor_test.go  # Fixtures: newTestKeyPair, startTLSServer, startTLSServerWithCert
│   └── ui/               # Terminal UI with lipgloss
│       └── ui.go
├── internal/             # Internal packages
│   ├── config/           # YAML config loading (~/.config/certwiz/config.yaml)
│   │   └── config.go
│   ├── environ/          # Environment detection
│   │   └── env.go       # IsCI/SupportsUnicode (cached) over detectCI/detectUnicode
│   └── testutil/         # TestdataPath(filename) for files under testdata/
├── testdata/             # Test certificates (committed)
└── docs/                  # Documentation
    ├── installation.md
    ├── usage.md
    ├── commands.md
    ├── examples.md
    ├── contributing.md
    ├── releasing.md
    ├── testing.md
    └── faq.md
```

## Key Technical Details

### Language & Dependencies
- **Language**: Go 1.20+ (see Go 1.20 Compatibility below)
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **UI Library**: Lipgloss v0.9.1 (github.com/charmbracelet/lipgloss)
- **Config**: gopkg.in/yaml.v3
- **Certificate Handling**: Go standard library (crypto/x509, crypto/tls)

### Build Commands
```bash
make build        # Creates ./cert binary
make install      # Installs to $GOPATH/bin
make clean        # Removes build artifacts
make build-all    # Cross-platform builds into dist/
```

### Command Structure
All commands follow this pattern:
```bash
cert [command] [target] [flags]
```

Commands:
- `inspect` - View certificate details from files, bundles, stdin (`-`), or URLs
- `generate` - Create self-signed certificates
- `convert` - Convert between PEM/DER formats (bundle-aware)
- `verify` - Validate certificates (dates, hostname, CA chain, private key, expiry window)
- `ca` - Create Certificate Authority certificates
- `csr` - Generate Certificate Signing Requests
- `sign` - Sign CSRs with a CA certificate
- `tls` - Test which TLS versions a server supports
- `update` - Update cert to the latest version
- `version` - Show version information
- `completion` - Generate shell completion scripts

Global flags: `--json` (machine output on stdout), `--plain` (no borders, colors, or emojis). `--timeout` on `inspect` and `tls` is a `time.Duration` flag (`DurationVar`, default 5s).

## Code Style Guidelines

### Go Code
- Follow standard Go idioms
- Use meaningful variable names
- Keep functions small and focused
- Handle errors explicitly
- Add comments for complex logic

### UI/UX Principles
- **Colors**: Green (valid), Yellow (warning), Red (error), Blue (info)
- **Output**: Readable, terminal-width aware
- **Defaults**: Smart defaults that just work
- **Errors**: Clear, actionable error messages

### Terminal Output
- Bordered tables for certificate info; continuation lines are indented to the value column by `formatTable`
- Terminal width is read once per display from stdout (`terminalWidth()`), falling back to 80
- Color coding for status indicators
- Symbols: check mark (success), cross (failure), arrow (detail), link (URL); `ui.Emoji(emoji, ascii)` picks the ASCII form under `--plain`, config, or CI. There is no private emoji helper in `cmd`.

## Common Tasks for AI Assistants

### Adding a New Command
1. Create `cmd/newcommand.go` with a `RunE`
2. Report failures with `reportError(cmd, err)` and `return err`
3. Add to rootCmd (in the file's `init()` or `cmd/root.go`)
4. Add the name to `expectedCommands` in `cmd/root_test.go`
5. Update documentation and CHANGELOG.md (Unreleased)

### Adding Certificate Features
1. Extend `pkg/cert/cert.go` (and `json.go` for JSON fields)
2. Add UI support in `pkg/ui/ui.go`
3. Wire up in the appropriate command
4. Add tests (see Test Conventions)
5. Update docs

### Updating Documentation
- Main README uses `cert` command in examples
- Docs use `cert` command throughout
- Keep project name as `certwiz` in descriptions

## Important Patterns

### Certificate Inspection
```go
// From file (first cert) or every cert in a bundle
cert, err := cert.InspectFile(path)
certs, err := cert.InspectFileAll(path)
certs, err := cert.InspectData(data, "stdin")

// From URL; the older wrappers delegate to InspectURLWithOptions
cert, chain, err := cert.InspectURLWithOptions(url, port, connectHost, timeout, sigAlg)
```
`newCertificate(c, source, format)` wraps an `*x509.Certificate` and computes `IsExpired` and `DaysUntilExpiry` once; use it instead of building `Certificate` literals.

### Conversion, Writing Files, TLS Probing
```go
// Convert returns the detected input format ("PEM"/"DER"). PEM output keeps
// every certificate in a bundle; DER output errors on multi-cert input.
inputFormat, err := cert.Convert(inputPath, outputPath, "der")

// All generators write through these helpers. Keys are opened 0600 directly
// (and re-chmodded on non-Windows); certificates are 0644.
err = writePrivateKey(keyPath, privateKey)   // PKCS#8 PEM
err = writeCertificate(certPath, der)
err = writePEMFile(path, block, perm)

// CheckTLSVersions probes TLS 1.0-1.3 concurrently. If at least one version
// succeeded, each failure is re-probed once sequentially (some servers cap
// concurrent handshakes). A host that answers nothing is not retried, so it
// fails in about one timeout.
result, err := cert.CheckTLSVersions(host, port, timeout)
```
Key-usage names come from `cert.KeyUsageNames`, `cert.ExtKeyUsageName`, and `cert.ExtKeyUsageNames` in `pkg/cert/usage.go`; both JSON and the terminal view use them. TLS version names come from `cert.TLSVersionName`.

### UI Display
```go
ui.SetConfig(cfg)                       // invalidates cached text styles
ui.DisplayCertificate(cert, showFull)
ui.DisplayCertificateChain(chain)
ui.DisplayTLSVersionResults(result)
ui.ShowErrorTo(w, msg)                  // ShowError writes to os.Stderr
```
Text styles are cached per config. `getPanelStyle()` stays a builder because callers mutate it (see Go 1.20 Compatibility for why).

### Error Handling
Commands use `RunE`. Cobra's own error printing is disabled on the root command.
```go
RunE: func(cmd *cobra.Command, args []string) error {
    result, err := cert.DoThing(...)
    if err != nil {
        reportError(cmd, err) // JSON payload on stdout under --json, styled message on stderr otherwise
        return err
    }
    ...
}
```
`Execute()` in `cmd/root.go` prints `Error: <msg>` to stderr for any error a command returned without calling `reportError` (flag errors, `verification failed`). Stream contract: human-readable errors on stderr, JSON payloads (`{"success": false, "error": "..."}`) on stdout, exit code 1 on failure, and every failure is printed exactly once.

## CI/CD and Linting

### Local Testing Before Push
**IMPORTANT**: Run the same checks CI runs before pushing.

```bash
# Tests with race detector, repeated to catch state leaks between cases
go test -race -count=3 ./...

# Lint exactly as CI does (golangci-lint v2, no config file). No docker needed:
GOFLAGS=-mod=mod go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout=5m ./...

# Go 1.20 floor (downloads the toolchain once)
GOTOOLCHAIN=go1.20.14 go build ./... && GOTOOLCHAIN=go1.20.14 go test ./...

gofmt -l .   # must print nothing
go vet ./...
```
The docker variants (`golangci/golangci-lint:latest`, `golang:1.20`) also work on machines that have docker.

CI (`.github/workflows/test.yml`) runs Go 1.20 and 1.21 on ubuntu and windows, `stable` on macOS, lint via `golangci/golangci-lint-action@v9` with `version: latest`, and build jobs. All actions are on Node 24 majors (checkout v7, setup-go v7, cache v6, upload-artifact v7, codecov v7, goreleaser-action v7). `CODECOV_TOKEN` is an optional repository secret; without it the upload runs tokenless and never fails the build.

### Linting Requirements
- All error returns must be checked, in tests too: `defer func() { _ = f.Close() }()`, `defer func() { _ = os.RemoveAll(dir) }()`, `_ = os.Setenv(...)`
- Use embedded struct fields directly (`cert.Verify(opts)`, not `cert.Certificate.Verify(opts)`)
- No `//nolint` directives for linters that do not exist (golangci warns on unknown names)
- Follow golangci-lint default rules (no custom .golangci.yml)

### Go 1.20 Compatibility
CI still tests Go 1.20 and 1.21 on Linux and Windows; the local toolchain may be much newer, so code that compiles locally can break CI.
- No `min`/`max` builtins, no `slices`, `maps`, or `cmp` packages, no `sync.OnceValue`, no range-over-int
- No `tls.VersionName` (Go 1.21); use `cert.TLSVersionName`
- Pass loop variables into goroutines explicitly (`go func(i int, v T) {...}(i, v)`)
- Lipgloss v0.9.1 shares a `Style`'s internal rules map between copies, so never cache a `Style` that is later mutated with builder methods such as `BorderForeground`. Cache only render-only styles; keep `getPanelStyle()` a builder.

### Common CI Issues & Solutions
1. **"Previous case" error in switch statements**: Missing import (e.g., crypto/ecdsa)
2. **"Could remove embedded field from selector"**: Use `cert.FieldName` instead of `cert.Certificate.FieldName`
3. **Unchecked errors**: Add `_ =` for intentionally ignored errors
4. **Failures only with `-count>1`**: package-level flag variables or cobra flag state leaked from an earlier case; reset with `t.Cleanup` or `resetCommandFlags`
5. **Missing test files**: Ensure all testdata files are committed to git (update .gitignore if needed)
6. **IP addresses in SANs**: Use `--san IP:192.168.1.1` format, not just `--san 192.168.1.1`

## Testing Guidelines

### CRITICAL: Always Run Tests Before Pushing
**The most common CI failures come from not running tests locally first!**

```bash
# CI's exact command (note the space, not '=', for Windows):
go test -v -race -coverprofile coverage.out ./...

# If tests pass locally but fail in CI, try:
go clean -testcache
go mod tidy
go test -v -race -coverprofile coverage.out ./...
```

### Test Conventions
- **No network in tests.** Remote inspection and `cert tls` are tested against in-process TLS servers: `startTLSServer` / `startTLSServerWithCert` / `newTestKeyPair` in `pkg/cert/refactor_test.go`, and `startTLSServer` / `startSilentListener` / `newTestCA` / `newTestLeaf` in `cmd/helpers_test.go`.
- **The suite must pass `go test -race -count=3 ./...`.** State leaking between cases was a real problem. Reset package-level flag variables, `jsonOutput`/`plainOutput`, and stdin with `t.Cleanup`; use `resetCommandFlags` before repeated `rootCmd.Execute` calls.
- **Deterministic output.** Use `setOutputMode(t, json, plain)` in cmd tests, or `ui.SetConfig` with a plain config (`config.DefaultConfig()` then `ApplyPlainMode()`) in ui tests, so assertions do not depend on ANSI codes, emojis, or the CI environment variable.
- **Capture the right stream.** `captureStdout` (cmd) and `captureOutput` (ui) capture stdout; errors go to stderr, so use `captureStreams` or `cmd.SetErr` when asserting on them.
- **Testdata paths** go through `testutil.TestdataPath("valid.pem")`, never hard-coded relative paths.
- Coverage is around 95% of statements (cmd ~84%, pkg/cert ~95%, pkg/ui ~98%, internal/environ 100%). New code should keep it there; the uncovered remainder is fault injection (PEM encoder or Close failing, `crypto/rand` failing, the update exec path).

### Debugging CI Test Failures

When CI tests fail but local tests pass:

1. **Check the exact error message in CI.** Syntax errors usually mean file corruption; build failures in one package cascade to others.
2. **Verify file integrity**
   ```bash
   tail -c 10 cmd/update.go | xxd                 # single trailing newline?
   curl -s https://raw.githubusercontent.com/trahma/certwiz/main/cmd/update.go | diff - cmd/update.go
   ```
3. **Reproduce the Go version** with `GOTOOLCHAIN=go1.20.14 go test ./...` (see Go 1.20 Compatibility).
4. **Reproduce a platform** with docker if available: `docker run --rm -v $(pwd):/app -w /app golang:1.20-alpine go test ./...`

### Test File Management

Test data files must be committed to git. `.gitignore` ignores `*.pem`, `*.crt`, `*.key`, and `*.der` everywhere except `testdata/`:
```bash
git ls-files testdata/      # confirm they are tracked
```

### Path Issues in Tests
Never hard-code `../../testdata/x.pem`; it breaks on Windows. Use `testutil.TestdataPath("x.pem")`, which builds the path with `filepath.Join`.

### Manual Testing Commands
```bash
./cert inspect google.com
./cert inspect google.com --full --chain
./cert inspect fullchain.pem --chain
openssl s_client -connect example.com:443 </dev/null | ./cert inspect -
./cert inspect google.com --connect localhost:8080 --timeout 2s

./cert generate --cn test.local --san test.local --san IP:192.168.1.1
./cert convert test.pem test.der --format der
./cert verify test.crt --host test.local
./cert verify test.crt --key test.key
./cert verify test.crt --expires-in 30d

./cert tls google.com
./cert tls google.com --json

./cert ca --plain          # error on stderr once, exit 1
./cert update
./cert update --force
```

### Common Test Domains
- google.com (many SANs)
- github.com (standard setup)
- expired.badssl.com (expired cert)
- self-signed.badssl.com (self-signed)

### Running Tests Locally Before Push Checklist

1. Run full test suite: `go test -race -count=3 ./...`
2. Lint and Go 1.20 build (see Local Testing Before Push)
3. Check for any modified files: `git status`
4. Ensure all new files are added: `git add .`
5. Verify builds cleanly: `go build -o cert .`
6. Test the binary: `./cert version`
7. If adding new commands, update `expectedCommands` in `cmd/root_test.go`

## Debugging Tips

### Build Issues
- Ensure Go 1.20+ is installed
- Run `go mod tidy` for dependencies
- Check `go.mod` for module name (certwiz)

### Display Issues
- Test with different terminal widths
- Check `$TERM` environment variable
- Test with `--plain` and with `CI=1` (ASCII borders and symbols)

### Certificate Issues
- Use `--full` flag for complete details
- Use `--chain` to see trust path
- Check SANs match hostname

## Completed Features

- CA certificate generation (`cert ca`)
- Certificate signing requests (`cert csr`)
- Certificate signing with CA (`cert sign`), including EC CA keys and DER CA certs
- TLS version testing (`cert tls`) with concurrent probing and a sequential retry
- SHA-256 and SHA-1 fingerprints in terminal and JSON output
- Certificate bundle and stdin inspection (`cert inspect fullchain.pem --chain`, `cert inspect -`)
- Bundle-aware conversion (`cert convert`)
- JSON output format (all commands support `--json`)
- Network timeout configuration (`--timeout` is a Go duration flag, e.g. `2s`, `500ms`)
- Proxy/tunnel support (`--connect` flag)
- Signature algorithm selection (`--sig-alg` flag for inspect)
- Automatic update command (`cert update`)
- Certificate chain verification (`--ca` flag for verify)
- Private key matching (`--key`) and expiry threshold (`--expires-in`) for verify
- Secure key permissions (0600 on Unix systems)
- Error stream contract: human errors on stderr, JSON on stdout, exit 1, printed once
- Plain output mode and YAML config file

## Future Enhancements (Roadmap)

These are planned but not yet implemented:
- ECDSA key generation (generate, ca, and csr are RSA-only)
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

1. Bump `version` in `cmd/root.go`
2. In CHANGELOG.md, move the `[Unreleased]` entries into a `## [X.Y.Z] - YYYY-MM-DD` section and add `[X.Y.Z]: https://github.com/trahma/certwiz/releases/tag/vX.Y.Z` to the link list at the bottom
3. Run the Local Testing Before Push checks
4. Commit, then `git tag -a vX.Y.Z -m "vX.Y.Z"`
5. `git push origin main && git push origin vX.Y.Z`
6. The tag push triggers `.github/workflows/goreleaser.yml`, which publishes archives named `cert-<os>-<arch>` (`x86_64`, `i386`, `arm64`, `armv7`; `.tar.gz`, `.zip` on Windows; each contains a binary named `cert`) plus `checksums.txt`
7. Watch it: `gh run list --limit 3`, then `gh run watch <id> --exit-status`; confirm with `gh release view vX.Y.Z`

`make build-all` cross-compiles into `dist/` for local checks but is not part of the release.

## Common Issues & Solutions

### "command not found"
- Binary is named `cert`, not `certwiz`
- Check PATH includes install directory

### Colors not showing
- Terminal may not support colors
- Try `FORCE_COLOR=1 cert inspect ...`

### SANs wrapping incorrectly
- Width comes from `terminalWidth()` (stdout); when stdout is not a terminal it is 80
- `valueWidth` derives the wrap width from the panel margin, padding, and longest key

### `cert inspect <path>` says the file does not exist
- Targets that look like paths (leading `.`/`~`, absolute, Windows drive, certificate extension, or a separator after a non-host segment) are treated as files; see `looksLikeFilePath` in `cmd/inspect.go`. Everything else is tried as a host or URL.

### `promo/` directory
- Local video/promo generation experiments; git-ignored, not part of the project

## Integration Points

### CI/CD
- Use `cert` command in scripts; prefer `--json` and parse with `jq`
- Exit codes: 0 (success), 1 (error); `cert verify --expires-in 30d` is designed for cron and pipelines
- Errors are on stderr, so `2>/dev/null` hides them and `--json` keeps stdout parseable

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
4. Update relevant documentation and CHANGELOG.md
5. Follow existing code style

## Contact & Support

- GitHub Issues: Bug reports and features
- Documentation: /docs directory
- Examples: /docs/examples.md

---

*This file helps AI assistants understand the project structure and conventions. Keep it updated as the project evolves.*
