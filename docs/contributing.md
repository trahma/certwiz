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
   - certwiz version (`certwiz --version`)
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

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/certwiz
   cd certwiz
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/certwiz/certwiz
   ```

4. Create a branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. Make your changes
2. Add tests if applicable
3. Ensure all tests pass:
   ```bash
   go test ./...
   ```

4. Format your code:
   ```bash
   go fmt ./...
   ```

5. Lint your code:
   ```bash
   golangci-lint run
   ```

6. Build and test locally:
   ```bash
   make build
   ./certwiz inspect google.com
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
- Aim for good test coverage
- Test edge cases and error conditions

Example test:
```go
func TestInspectFile(t *testing.T) {
    cert, err := InspectFile("testdata/valid.pem")
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
├── cmd/                 # CLI commands
│   ├── root.go         # Root command
│   ├── inspect.go      # Inspect command
│   ├── generate.go     # Generate command
│   ├── convert.go      # Convert command
│   └── verify.go       # Verify command
├── pkg/                # Core packages
│   ├── cert/          # Certificate operations
│   └── ui/            # Terminal UI
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
       Run: func(cmd *cobra.Command, args []string) {
           // Implementation
       },
   }
   
   func init() {
       rootCmd.AddCommand(newCmd)
   }
   ```

2. Add tests in `cmd/newcmd_test.go`

3. Update documentation

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

Maintainers handle releases:

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag
4. GitHub Actions builds releases
5. Update documentation

## Getting Help

- Join our [Discord server](https://discord.gg/certwiz)
- Check [existing issues](https://github.com/certwiz/certwiz/issues)
- Read the [documentation](https://github.com/certwiz/certwiz/docs)
- Ask in [Discussions](https://github.com/certwiz/certwiz/discussions)

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project README

## License

By contributing, you agree that your contributions will be licensed under the MIT License.