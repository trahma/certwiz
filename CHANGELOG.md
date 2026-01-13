# Changelog

All notable changes to certwiz will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.3] - 2026-01-13

### Added
- **Plain output mode** with `--plain` flag
  - Disables borders, colors, and emojis for easy copy/paste
  - Works with all commands
- **Config file support**
  - Locations: `~/.config/certwiz/config.yaml` or `~/.certwiz.yaml`
  - Configure default output preferences (borders, colors, emojis)
  - Priority: `--plain` flag > config file > CI detection > defaults

### Documentation
- Added configuration section to README
- Documented config file format and locations

## [0.2.2] - 2026-01-13

### Added
- **New `cert tls` command** for TLS version testing
  - Tests which TLS versions a server supports (1.0, 1.1, 1.2, 1.3)
  - Shows supported versions with color-coded status
  - Includes `--json` output for automation
  - Supports `--timeout` flag for connection timeout

## [0.2.1] - 2025-09-02

### Added
- **Signature algorithm selection for `inspect` command**
  - New `--sig-alg` flag to force ECDSA or RSA certificate selection
  - Supports values: `auto` (default), `ecdsa`, `rsa`
  - Works by controlling cipher suite advertisement in TLS ClientHello
  - Useful for testing servers with dual-certificate configurations
  - Only affects TLS 1.2 and below (TLS 1.3 handles signatures differently)

### Documentation
- Added examples for `--sig-alg` flag usage
- Updated command reference with new flag details
- Added explanation of signature algorithm selection mechanism

## [0.2.0] - 2025-09-02

### Added
- **Centralized environment helpers** (`internal/environ`)
  - `IsCI()` for detecting CI environments
  - `SupportsUnicode()` for terminal capability detection
- **CA chain verification** in `verify` command
  - New `--ca` flag accepts PEM or DER format CA certificates
  - Full chain validation using `x509.VerifyOptions`
- **Network timeout for `inspect` command**
  - Default 5-second connection timeout to prevent hangs
  - New `--timeout` flag for custom timeout values
- **DRY SAN parsing** via `pkg/cert/san.go`
  - Supports DNS, IP, email, and URI SANs
  - Consistent parsing across generate, CSR, and CA commands

### Changed
- **Secure key permissions**: Generated private keys now have `0600` permissions on Unix
- **Standardized JSON output** with unified `printJSON` and `printJSONError` helpers
- **Consistent error handling**: All commands now use `RunE` with proper error returns
- **Improved code organization**: Reduced duplication across commands
- **Enhanced documentation**: Updated commands.md, usage.md, and FAQ

### Security
- Private keys are now created with restrictive permissions (0600) by default
- Proper CA certificate validation in verify command

## [0.1.10] - 2025-08-20

### Fixed
- **Critical fix for `cert update` command**
  - Fixed syscall.Exec argument issue that caused installer to drop to shell
  - Properly pass "bash" as argv[0] for syscall.Exec
  - Update command now correctly executes the installer script

## [0.1.9] - 2025-08-20

### Added
- **New `--connect` flag for inspect command**
  - Allows connecting to a different host while validating the certificate for the target hostname
  - Useful for testing certificates through proxies, tunnels, or local services
  - Supports port specification in the connect host (e.g., `--connect localhost:8080`)
  - Example: `cert inspect api.example.com --connect localhost:8080`

### Changed
- Enhanced inspect command to support split hostname/connection scenarios
- Improved examples in help text to show proxy/tunnel use cases

## [0.1.8] - 2025-08-20

### Fixed
- **Critical fix for macOS auto-update functionality**
  - Fixed SIGKILL issue when running `cert update` on macOS
  - Installer now uses atomic move operations for self-updates
  - Update command uses syscall.Exec to break process inheritance chain
  - Clears macOS extended attributes (com.apple.provenance) that prevent execution
- **Improved update reliability**
  - Downloads installer to temp file instead of piping directly
  - Better handling of running binary replacement
  - Fallback mechanisms for compatibility

### Changed
- Removed duplicate release workflow (now using only goreleaser)
- Enhanced installer script with self-update detection
- Simplified default command output (run without arguments) - examples now only shown with `cert help`

## [0.1.7] - 2025-08-20

### Fixed
- Fixed goreleaser configuration for v2 compatibility
- Added version: 2 declaration to .goreleaser.yml
- Updated release and brews sections for v2 format

### Note
This is a hotfix release to address the goreleaser build issue in v0.1.6.
All features from v0.1.6 are included in this release.

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

[0.2.3]: https://github.com/trahma/certwiz/releases/tag/v0.2.3
[0.2.2]: https://github.com/trahma/certwiz/releases/tag/v0.2.2
[0.2.1]: https://github.com/trahma/certwiz/releases/tag/v0.2.1
[0.2.0]: https://github.com/trahma/certwiz/releases/tag/v0.2.0
[0.1.10]: https://github.com/trahma/certwiz/releases/tag/v0.1.10
[0.1.9]: https://github.com/trahma/certwiz/releases/tag/v0.1.9
[0.1.8]: https://github.com/trahma/certwiz/releases/tag/v0.1.8
[0.1.7]: https://github.com/trahma/certwiz/releases/tag/v0.1.7
[0.1.6]: https://github.com/trahma/certwiz/releases/tag/v0.1.6
[0.1.5]: https://github.com/trahma/certwiz/releases/tag/v0.1.5
[0.1.4]: https://github.com/trahma/certwiz/releases/tag/v0.1.4
[0.1.3]: https://github.com/trahma/certwiz/releases/tag/v0.1.3
[0.1.2]: https://github.com/trahma/certwiz/releases/tag/v0.1.2
[0.1.1]: https://github.com/trahma/certwiz/releases/tag/v0.1.1
[0.1.0]: https://github.com/trahma/certwiz/releases/tag/v0.1.0