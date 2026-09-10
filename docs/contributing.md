# Contributing to certwiz

Thank you for your interest in contributing to certwiz! We welcome contributions from the community and are grateful for any help you can provide.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on what is best for the community
- Show empathy towards other community members

## How to Contribute

### Reporting Issues

Found a bug or have a feature request? Please open an issue:

1. Check if the issue already exists
2. Use a clear and descriptive title
3. Provide as much information as possible:
   - certwiz version (`cert version`)
   - Operating system and version
   - Steps to reproduce the issue
   - Expected vs actual behavior
   - Any error messages or logs

### Suggesting Enhancements

We love feature suggestions! Please:

1. Check if the feature has already been suggested
2. Explain the use case and why it would be useful
3. Provide examples of how it would work
4. Consider if it aligns with certwiz's goal of simplicity

### Pull Requests

We actively welcome pull requests! Here's how:

#### Setup Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/certwiz
   cd certwiz
   ```
3. Install Go 1.20 or higher (if not already installed)
4. Install development tools:
   ```bash
   # Install the latest version to test against
   curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash
   
   # Build from source for development
   make build
   ```

5. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/trahma/certwiz
   ```

6. Create a branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. Make your changes
2. Add tests if applicable
3. Ensure all tests pass, repeatedly and with the race detector (state leaking between test cases has broken CI before):
   ```bash
   go test -race -count=3 ./...
   ```
   Tests must not touch the network. Remote behaviour is tested against in-process TLS servers; see the fixtures in `pkg/cert/refactor_test.go` and `cmd/helpers_test.go`.

4. Format and vet your code:
   ```bash
   gofmt -l .        # must print nothing
   go vet ./...
   ```

5. Lint with the same linter CI uses (golangci-lint v2, no config file, strict `errcheck` including in tests):
   ```bash
   GOFLAGS=-mod=mod go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout=5m ./...
   ```

6. Confirm the code still builds on the minimum Go version (CI tests Go 1.20 and 1.21; avoid `min`/`max` builtins, the `slices`/`maps`/`cmp` packages, `sync.OnceValue`, and range-over-int):
   ```bash
   GOTOOLCHAIN=go1.20.14 go build ./... && GOTOOLCHAIN=go1.20.14 go test ./...
   ```

7. Build and test locally:
   ```bash
   make build
   ./cert inspect google.com
   ```

#### Commit Guidelines

We follow conventional commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Test additions or changes
- `chore:` Maintenance tasks

Examples:
```bash
git commit -m "feat: add support for EC certificates"
git commit -m "fix: correct SAN parsing for wildcard domains"
git commit -m "docs: update installation instructions for Windows"
```

#### Submitting Pull Request

1. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a Pull Request with:
   - Clear title and description
   - Link to any related issues
   - Screenshots if UI changes
   - Test results

3. Address review feedback
4. Ensure CI passes

## Development Guidelines

### Code Style

- Follow Go idioms and best practices
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused
- Error messages should be helpful and actionable

### Testing

- Write unit tests for new functionality
- Update existing tests when modifying code
- Keep coverage where it is (around 95% of statements)
- Test edge cases and error conditions
- Never use the real network in tests; use the in-process TLS server fixtures
- Reset package-level flag variables and output mode with `t.Cleanup` so cases do not leak into each other

Example test:
```go
func TestInspectFile(t *testing.T) {
    cert, err := InspectFile(testutil.TestdataPath("valid.pem"))
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if cert.Subject.CommonName != "example.com" {
        t.Errorf("expected CN=example.com, got %s", cert.Subject.CommonName)
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Update command help text
- Add/update documentation in docs/
- Include examples for new features

### Dependencies

- Minimize external dependencies
- Justify any new dependencies in PR
- Keep dependencies up to date
- Use go.mod for version management

## Project Structure

```
certwiz/
├── main.go              # Entry point
├── cmd/                 # CLI commands (one file per command, helpers.go for shared error/JSON output)
├── pkg/                # Core packages
│   ├── cert/          # Certificate operations, JSON output, usage name tables
│   └── ui/            # Terminal UI
├── internal/           # config (YAML settings), environ (CI/Unicode detection), testutil
├── docs/              # Documentation
└── testdata/          # Test fixtures
```

## Adding New Features

### Adding a New Command

1. Create new file in `cmd/`:
   ```go
   // cmd/newcmd.go
   package cmd
   
   import "github.com/spf13/cobra"
   
   var newCmd = &cobra.Command{
       Use:   "newcmd",
       Short: "Brief description",
       Long:  `Detailed description`,
       RunE: func(cmd *cobra.Command, args []string) error {
           if err := doWork(); err != nil {
               reportError(cmd, err) // JSON on stdout under --json, styled message on stderr otherwise
               return err
           }
           return nil
       },
   }
   
   func init() {
       rootCmd.AddCommand(newCmd)
   }
   ```

2. Add tests in `cmd/newcmd_test.go` and add the command name to `expectedCommands` in `cmd/root_test.go`

3. Update documentation and the `## [Unreleased]` section of CHANGELOG.md

### Adding Certificate Support

1. Extend `pkg/cert/cert.go`:
   ```go
   func NewCertificateOperation() error {
       // Implementation
   }
   ```

2. Add UI support in `pkg/ui/ui.go`

3. Wire up in appropriate command

### Improving UI

1. Use lipgloss styles consistently
2. Maintain color scheme
3. Ensure terminal width compatibility
4. Test on different terminal emulators

## Release Process

Maintainers handle releases. See [releasing.md](releasing.md). In short: bump `version` in `cmd/root.go`, move the Unreleased changelog entries into a dated section, commit, push an annotated `vX.Y.Z` tag, and GoReleaser publishes the release.

## Getting Help

- Check [existing issues](https://github.com/trahma/certwiz/issues)
- Read the [documentation](https://github.com/trahma/certwiz/tree/main/docs)
- Ask in [Discussions](https://github.com/trahma/certwiz/discussions)

## Recognition

Contributors are recognized in release notes.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
