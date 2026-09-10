# certwiz

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> A user-friendly CLI tool for certificate management. Like HTTPie, but for certificates.

certwiz makes working with X.509 certificates as simple as possible. No more wrestling with OpenSSL's arcane syntax or trying to remember complex command flags. Just simple, intuitive commands that do what you expect.

## Features

- **Inspect** certificates from files, bundles, stdin, or live websites
- **Generate** self-signed certificates with custom SANs
- **Create CSRs** (Certificate Signing Requests) for CA signing
- **Create CAs** to sign certificates and build trust chains
- **Sign certificates** using your own Certificate Authority
- **Convert** between PEM and DER formats, bundles included
- **Verify** certificates against hostnames, CA chains, private keys, and expiry windows
- **Test TLS versions** a server supports and flag deprecated ones
- **View certificate chains** to understand trust paths
- **SHA-256 and SHA-1 fingerprints** for pinning and comparison
- **Detailed extension analysis** with human-readable output
- **Beautiful terminal output** with colors and formatting
- **JSON output** for scripting and automation
- **Self-update** with `cert update`
- **Smart defaults** that just work

## Quick Start

```bash
# Inspect a website's certificate
cert inspect google.com

# Generate a self-signed certificate
cert generate --cn myapp.local --san "*.myapp.local"

# Create a Certificate Signing Request
cert csr --cn server.example.com --org "My Company"

# Create a Certificate Authority
cert ca --cn "Company Root CA" --org "My Company"

# Sign a CSR with your CA
cert sign --csr server.csr --ca ca.crt --ca-key ca.key

# Convert certificate format
cert convert cert.pem cert.der --format der

# View the full certificate chain
cert inspect github.com --chain

# Inspect a bundle (fullchain.pem) or pipe a certificate in
cert inspect fullchain.pem --chain
openssl s_client -connect example.com:443 </dev/null | cert inspect -

# Verify a certificate: hostname, CA chain, matching key, and expiry window
cert verify server.crt --host example.com --ca ca.crt
cert verify server.crt --key server.key
cert verify server.crt --expires-in 30d   # exit 1 if it expires within 30 days

# Test which TLS versions a server supports
cert tls example.com

# Inspect through a proxy or tunnel
cert inspect api.example.com --connect localhost:8080
cert inspect internal.site --connect tunnel.local --port 443

# Force specific certificate type (for dual-cert servers)
cert inspect cloudflare.com --sig-alg ecdsa  # Get ECDSA certificate
cert inspect cloudflare.com --sig-alg rsa    # Get RSA certificate
```

## Installation

### Quick Install (Recommended)

Install the latest version with our installer script:

```bash
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash
```

Or install a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --version v0.4.1
```

### Updating

To update cert to the latest version:

```bash
cert update
```

The installer will automatically detect your existing installation and upgrade it in place.

### Manual Installation

Download pre-built binaries from the [releases page](https://github.com/trahma/certwiz/releases). Each archive contains a single binary named `cert`. Builds are available for macOS (Apple Silicon and Intel), Linux (x86_64, arm64, armv7, i386), FreeBSD, and Windows; a `checksums.txt` is published alongside them.

#### macOS
```bash
# Apple Silicon (M1/M2/M3/M4)
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-arm64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert

# Intel
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-x86_64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

#### Linux
```bash
# x86_64
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-x86_64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert

# arm64 (Raspberry Pi 4/5, AWS Graviton, etc.)
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-arm64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

#### Windows

Download `cert-windows-x86_64.zip` (or `arm64`) from the releases page, extract `cert.exe`, and place it somewhere on your `PATH`. Note that `cert update` is not available on Windows; download a new release to upgrade.

### From Source

Requires Go 1.20 or later.

```bash
git clone https://github.com/trahma/certwiz
cd certwiz
make build        # produces ./cert
make install      # installs to $GOPATH/bin
```

## Documentation

- [Installation Guide](docs/installation.md)
- [Usage Guide](docs/usage.md)
- [Command Reference](docs/commands.md)
- [Examples](docs/examples.md)
- [FAQ](docs/faq.md)
- [Contributing](docs/contributing.md)

## Why certwiz?

### Before certwiz (with OpenSSL)
```bash
# Inspecting a certificate - hard to remember!
openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -text -noout

# Generating a certificate with SANs - so complex!
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=example.com" -extensions v3_req -config <(echo "[req]"; echo "distinguished_name=req_distinguished_name"; echo "[v3_req]"; echo "subjectAltName=DNS:example.com,DNS:*.example.com")
```

### With cert - simple and intuitive!
```bash
# Inspecting a certificate
cert inspect example.com

# Generating a certificate with SANs
cert generate --cn example.com --san example.com --san "*.example.com"
```

## JSON Output

All commands support JSON output for easy scripting and automation:

```bash
# Inspect with JSON output
cert inspect google.com --json | jq '.subject.common_name'

# Generate and get file paths
cert generate --cn test.local --json | jq '.files[]'

# Verify and check status
cert verify cert.pem --json | jq '.is_valid'

# Parse certificate expiry
cert inspect cert.pem --json | jq '.days_until_expiry'

# Pin on a fingerprint
cert inspect example.com --json | jq -r '.fingerprint_sha256'

# Check TLS version support
cert tls example.com --json | jq '.min_supported'
```

Commands exit with status 1 on failure. Human-readable errors go to stderr; with `--json`, the error payload (`{"success": false, "error": "..."}`) goes to stdout so scripts can parse it. `cert verify` exits 1 when any check fails, which makes it usable directly in CI and cron jobs:

```bash
# Fail the pipeline if the certificate expires within 14 days
cert verify /etc/ssl/certs/site.pem --expires-in 14d
```

## Configuration

### Plain Output Mode

Use `--plain` flag to disable borders, colors, and emojis for easy copy/paste:

```bash
cert inspect google.com --plain
cert tls github.com --plain
```

### Config File

Create a config file to set default preferences:

**Locations** (checked in order):
1. `~/.config/certwiz/config.yaml`
2. `~/.certwiz.yaml`

**Example config:**
```yaml
output:
  plain: false      # Master switch for plain mode
  borders: true     # Show bordered panels
  colors: true      # Use colored output
  emojis: true      # Show emojis (check marks, etc.)
```

**Priority order:**
1. `--plain` flag (overrides everything)
2. Config file settings
3. CI environment detection (auto-disables emojis)
4. Built-in defaults

## Key Features in Detail

### Certificate Inspection
- View certificates from files (PEM/DER), bundles such as `fullchain.pem`, stdin (`cert inspect -`), or live websites
- Automatic format detection
- SHA-256 and SHA-1 fingerprints
- Negotiated TLS version and cipher suite for live connections
- Shows all SANs (DNS, IP, email, URI) with intelligent wrapping
- Highlights expiration status with color coding
- Displays full certificate chain with `--chain`
- Shows detailed extensions with `--full`
- Connect through proxies/tunnels with `--connect` flag
- Force ECDSA or RSA certificate selection with `--sig-alg` flag
- Configurable network timeout with `--timeout` (e.g. `2s`, `500ms`)

### Certificate Generation
- Create self-signed certificates, CAs, and CSRs instantly
- Support for multiple SANs (DNS names, IP addresses, emails, URIs)
- Customizable validity period and key size
- Random serial numbers per RFC 5280
- Private keys are written with `0600` permissions

### Certificate Verification
- Check certificate validity dates
- Verify hostname matches
- Validate against CA certificates (PEM or DER)
- Confirm a private key matches the certificate with `--key`
- Fail early on upcoming expiry with `--expires-in`
- Clear pass/fail status indicators and a non-zero exit code on failure

### TLS Version Testing
- Probe TLS 1.0 through 1.3 concurrently with `cert tls`
- Shows the negotiated cipher suite for each supported version
- Warns when deprecated TLS 1.0 or 1.1 are still enabled

### Beautiful Output
- Color-coded status indicators (green for valid, yellow for expiring soon, red for expired)
- Clean, bordered tables for certificate information
- Smart terminal width detection and text wrapping
- Icons and symbols for better readability

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/contributing.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [HTTPie](https://httpie.io/) for its user-friendly approach
- Built with [Cobra](https://github.com/spf13/cobra) for CLI management
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss) for beautiful output

