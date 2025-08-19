# Changelog

All notable changes to certwiz will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[0.1.0]: https://github.com/certwiz/certwiz/releases/tag/v0.1.0