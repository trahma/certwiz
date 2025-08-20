# Changelog

All notable changes to certwiz will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.6] - 2025-08-20

### Added
- **PKI (Public Key Infrastructure) functionality**
  - New `cert ca` command to create Certificate Authority certificates
  - New `cert csr` command to generate Certificate Signing Requests  
  - New `cert sign` command to sign CSRs with a CA certificate
  - Complete PKI workflow support for internal certificate management
- **JSON output support**
  - Global `--json` flag for all commands
  - Structured JSON output for scripting and automation
  - JSON formatting for certificates, CSRs, and verification results
- **Enhanced certificate display**
  - Added CSR information display functionality
  - Improved certificate chain visualization
  - Better formatting for Subject Alternative Names (SANs)

### Fixed
- **CI/CD compatibility improvements**
  - Fixed Unicode character rendering issues in CI environments
  - Added ASCII fallback for terminals that don't support Unicode
  - Replaced emoji with ASCII equivalents in CI mode
- **Cross-platform compatibility**
  - Fixed Windows file permission issues in tests
  - Added OS-specific permission handling for private keys
  - Improved path handling for Windows systems
- **Test reliability**
  - Fixed TestCommandStructure to account for Cobra auto-generated commands
  - Added missing newlines in generated files for Go compatibility
  - Improved test coverage for new PKI features

### Changed
- Updated version to 0.1.6
- Improved code formatting with gofmt
- Enhanced error messages for better user experience

## [0.1.5] - 2025-08-19

### Added
- Self-update functionality with `cert update` command
- Version checking and automatic update notifications

### Fixed
- Various bug fixes and improvements

## [0.1.4] - 2025-08-19

### Fixed
- Installer script improvements
- Binary detection enhancements

## [0.1.3] - 2025-08-19

### Fixed
- Windows test path problems
- Cross-platform compatibility improvements

## [0.1.2] - 2025-08-19

### Added
- Improved installer script
- Better binary management

### Fixed
- Installation issues on various platforms

## [0.1.1] - 2024-08-19

### Fixed
- Minor bug fixes and improvements
- Documentation updates

## [0.1.0] - 2024-08-19

### Added
- Initial release of certwiz
- `cert inspect` command for viewing certificates from files or URLs
- `cert generate` command for creating self-signed certificates with SANs
- `cert convert` command for converting between PEM and DER formats
- `cert verify` command for validating certificates
- Beautiful terminal output with colors and formatting
- Smart text wrapping for long SANs lists
- Certificate chain viewing with `--chain` flag
- Detailed extension analysis with `--full` flag
- Support for custom ports when inspecting remote certificates
- Automatic format detection for certificate files
- Comprehensive documentation
- CLAUDE.md for AI assistant context

### Features
- Color-coded certificate status (valid, expiring, expired)
- Human-readable certificate extension display
- Support for wildcard certificates
- IP address SANs support
- Terminal width detection for responsive output
- Cross-platform support (macOS, Linux, Windows)

### Technical
- Built with Go 1.20+
- Uses Cobra for CLI framework
- Uses Lipgloss for beautiful terminal UI
- Binary named `cert` for ease of use
- Project name remains `certwiz`

[0.1.6]: https://github.com/trahma/certwiz/releases/tag/v0.1.6
[0.1.5]: https://github.com/trahma/certwiz/releases/tag/v0.1.5
[0.1.4]: https://github.com/trahma/certwiz/releases/tag/v0.1.4
[0.1.3]: https://github.com/trahma/certwiz/releases/tag/v0.1.3
[0.1.2]: https://github.com/trahma/certwiz/releases/tag/v0.1.2
[0.1.1]: https://github.com/trahma/certwiz/releases/tag/v0.1.1
[0.1.0]: https://github.com/trahma/certwiz/releases/tag/v0.1.0