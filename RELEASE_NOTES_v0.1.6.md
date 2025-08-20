# certwiz v0.1.6

## 🎉 Major Features

### 🔐 PKI (Public Key Infrastructure) Support
certwiz now includes complete PKI functionality for managing your own certificate infrastructure:

- **`cert ca`** - Create Certificate Authority certificates
- **`cert csr`** - Generate Certificate Signing Requests
- **`cert sign`** - Sign CSRs with your CA certificate

Example workflow:
```bash
# Create a CA
cert ca --cn "My Company CA" --org "My Company" --country US

# Generate a CSR
cert csr --cn server.example.com --san server.example.com --san www.example.com

# Sign the CSR with your CA
cert sign --csr server.csr --ca ca.crt --ca-key ca.key --days 365
```

### 📊 JSON Output Support
All commands now support JSON output for scripting and automation:

```bash
# Get certificate info as JSON
cert inspect server.crt --json

# Generate and output as JSON
cert generate --cn test.local --json

# Verify and get JSON results
cert verify server.crt --json
```

## 🐛 Bug Fixes & Improvements

### CI/CD Compatibility
- Fixed Unicode rendering issues in CI environments
- Added ASCII fallback for terminals without Unicode support
- Emoji are replaced with text labels in CI mode

### Cross-Platform Support
- Fixed Windows file permission handling
- Improved compatibility across Linux, macOS, and Windows
- Better path handling for all operating systems

### Test Improvements
- Enhanced test coverage for new PKI features
- Fixed test compatibility issues
- Added platform-specific test handling

## 📝 Full Changelog

See [CHANGELOG.md](https://github.com/trahma/certwiz/blob/main/CHANGELOG.md) for complete details.

## 🚀 Installation

### macOS/Linux
```bash
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash
```

### From Source
```bash
go install github.com/trahma/certwiz@v0.1.6
```

## 📖 Documentation

- [README](https://github.com/trahma/certwiz/blob/main/README.md)
- [CHANGELOG](https://github.com/trahma/certwiz/blob/main/CHANGELOG.md)

## 🙏 Acknowledgments

Thanks to all contributors and users who helped make this release possible!

---

**Full Changelog**: https://github.com/trahma/certwiz/compare/v0.1.5...v0.1.6