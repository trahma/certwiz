# Testing Guide

## Running Tests

The suite needs no network access and no external tools; every remote scenario
runs against an in-process TLS server.

```bash
# What CI runs
go test -v -race -coverprofile coverage.out ./...

# What you should run before pushing: three iterations catch state leaking
# between test cases, which has bitten this project before
go test -race -count=3 ./...

# One package
go test -race ./pkg/cert/
go test -race ./pkg/ui/
go test -race ./cmd/

# Coverage report
go test -coverprofile coverage.out ./... && go tool cover -func coverage.out
make test-coverage-html    # HTML report in coverage.html
```

`make test` regenerates the fixtures with `testdata/generate_test_certs.sh`
and then runs `go test -v ./...`. The fixtures are committed, so running
`go test` directly is enough for normal development.

### Go version check

CI tests Go 1.20 and 1.21 on Linux and Windows. Your local toolchain is
probably newer, so also build and test with the floor version before pushing:

```bash
GOTOOLCHAIN=go1.20.14 go build ./... && GOTOOLCHAIN=go1.20.14 go test ./...
```

Avoid `min`/`max` builtins, the `slices`, `maps`, and `cmp` packages,
`sync.OnceValue`, range-over-int, and `tls.VersionName` (use
`cert.TLSVersionName`).

### Lint

CI runs golangci-lint v2 with no config file. The same check without docker:

```bash
GOFLAGS=-mod=mod go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout=5m ./...
```

`errcheck` is strict in tests too: wrap deferred `Close` and `RemoveAll` calls
(`defer func() { _ = f.Close() }()`) and use `_ = os.Setenv(...)`.

## Test Layout

| Package | Files | What they cover |
|---|---|---|
| `pkg/cert` | `cert_test.go`, `features_test.go`, `inspect_url_test.go`, `csr_test.go`, `json_test.go`, `errors_test.go`, `tls_test.go`, `tls_retry_test.go`, `usage_test.go`, `refactor_test.go` | Parsing, generation, conversion, verification, remote inspection, TLS probing, JSON output, error paths |
| `pkg/ui` | `ui_test.go`, `display_test.go` | Table formatting, wrapping, every `Display*` function including `--full` extensions |
| `cmd` | `*_test.go` per command plus `report_error_test.go`, `json_test.go`, `helpers_test.go` | Commands run in-process via `RunE`, flag handling, JSON mode, stdout/stderr separation |
| `internal/config` | `config_test.go` | Config file loading and plain mode |
| `internal/environ` | `env_test.go` | CI and Unicode detection with an injected `getenv` |

### Fixtures

Shared helpers live in test files, not in a library:

- `pkg/cert/refactor_test.go`: `newTestKeyPair`, `startTLSServer(t, minVersion, maxVersion)`, `startTLSServerWithCert` for serving a chain.
- `cmd/helpers_test.go`: `captureStdout`, `captureStreams`, `withStdin`, `setOutputMode(t, json, plain)`, `resetCommandFlags`, `newTestCA`, `newTestLeaf`, `startTLSServer`, `startSilentListener`.
- `pkg/ui/ui_test.go`: `captureOutput`.
- `internal/testutil`: `TestdataPath("valid.pem")` builds a cross-platform path to `testdata/`.

`testdata/` holds committed certificates and keys used by file-based tests:

| File | Purpose |
|---|---|
| `valid.pem`, `valid.der`, `valid.key` | A valid certificate in both encodings with its key |
| `expired.pem`, `expired.key` | Expired certificate |
| `many-sans.pem`, `many-sans.key` | Certificate with many SANs (wrapping tests) |
| `strong.pem`, `strong.key` | 4096-bit RSA certificate |
| `invalid.pem` | Corrupted data for error tests |
| `ca.pem`, `ca.key`, `intermediate.pem`, `intermediate.key`, `chain-server.pem`, `chain-server.key` | A three-level chain |
| `fullchain.pem` | Bundle of the chain (server, intermediate, CA) |

Regenerate them with `make test-generate-certs`. The `.gitignore` excludes
`*.pem`, `*.key`, `*.crt`, and `*.der` everywhere except `testdata/`, so new
fixtures must be added there.

## Conventions

- **No network.** Use `startTLSServer` (or a plain `net.Listen` for a silent
  listener) on `127.0.0.1`. Tests that need an unresolvable host use
  `example.invalid`.
- **No leaked state.** Reset package-level flag variables, `jsonOutput`,
  `plainOutput`, and stdin with `t.Cleanup`. Call `resetCommandFlags` before
  running `rootCmd.Execute` again; cobra keeps parsed values between runs.
- **Deterministic output.** Use `setOutputMode` in cmd tests or `ui.SetConfig`
  with a plain config (`config.DefaultConfig()` then `ApplyPlainMode()`) in ui
  tests so assertions do not depend on ANSI codes, emojis, or the `CI`
  environment variable.
- **Assert on the right stream.** Errors go to stderr; use `captureStreams` or
  `cmd.SetErr` when checking them. JSON error payloads go to stdout.
- **Paths** go through `testutil.TestdataPath`, never `../../testdata`.
- **Table-driven tests** for classification and formatting helpers; one
  subtest per scenario.

## Coverage

Coverage is around 95% of statements overall: `cmd` about 84%, `pkg/cert`
about 95%, `pkg/ui` about 98%, `internal/environ` 100%. New code should keep
it there. The uncovered remainder is fault injection with no natural trigger
(a PEM encoder or `Close` failing mid-write, `crypto/rand` failing, the
`exec` path of `cert update`).

## Continuous Integration

`.github/workflows/test.yml` runs on every push and pull request:
- Tests with `-race` on Ubuntu and Windows with Go 1.20 and 1.21, and on macOS with the latest stable Go
- Lint with golangci-lint v2
- Build jobs that execute the binary on all three platforms
- Coverage upload to Codecov (needs the optional `CODECOV_TOKEN` secret to succeed)

`.github/workflows/goreleaser.yml` builds and publishes release archives when a
`v*` tag is pushed. See [Releasing](releasing.md).

## Writing Tests

When adding a feature:
1. Add unit tests in the package that owns the logic.
2. Add a command-level test that runs `xCmd.RunE` in plain and `--json` mode.
3. Add fixtures to `testdata/` only if an in-test generated certificate will not do.
4. Run `go test -race -count=3 ./...`, the lint command, and the Go 1.20 check.
5. If you added a command, update `expectedCommands` in `cmd/root_test.go`.
