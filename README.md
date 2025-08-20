# certwiz 🔐

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> A user-friendly CLI tool for certificate management. Like HTTPie, but for certificates.

certwiz makes working with X.509 certificates as simple as possible. No more wrestling with OpenSSL's arcane syntax or trying to remember complex command flags. Just simple, intuitive commands that do what you expect.

## ✨ Features

- 🔍 **Inspect** certificates from files or live websites
- 🔐 **Generate** self-signed certificates with custom SANs
- 📝 **Create CSRs** (Certificate Signing Requests) for CA signing
- 🏛️ **Create CAs** to sign certificates and build trust chains
- ✍️ **Sign certificates** using your own Certificate Authority
- 🔄 **Convert** between PEM and DER formats effortlessly
- ✅ **Verify** certificates against hostnames
- 🔗 **View certificate chains** to understand trust paths
- 📊 **Detailed extension analysis** with human-readable output
- 🎨 **Beautiful terminal output** with colors and formatting
- 📄 **JSON output** for scripting and automation
- 💡 **Smart defaults** that just work

## 🚀 Quick Start

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

# Inspect through a proxy or tunnel
cert inspect api.example.com --connect localhost:8080
cert inspect internal.site --connect tunnel.local --port 443
```

## 📦 Installation

### Quick Install (Recommended)

Install the latest version with our installer script:

```bash
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash
```

Or install a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --version v0.1.0
```

### Updating

To update cert to the latest version:

```bash
cert update
```

The installer will automatically detect your existing installation and upgrade it in place.

### Manual Installation

Download pre-built binaries from the [releases page](https://github.com/trahma/certwiz/releases).

#### macOS
```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-arm64.tar.gz | tar xz
sudo mv cert-darwin-arm64 /usr/local/bin/cert

# Intel
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-amd64.tar.gz | tar xz
sudo mv cert-darwin-amd64 /usr/local/bin/cert
```

#### Linux
```bash
# 64-bit
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-amd64.tar.gz | tar xz
sudo mv cert-linux-amd64 /usr/local/bin/cert
```

### From Source

```bash
go install github.com/trahma/certwiz@latest
# or
git clone https://github.com/trahma/certwiz
cd certwiz
make build
```

### Download Binary

Download pre-built binaries from the [releases page](https://github.com/certwiz/certwiz/releases).

The binary will be named `cert` for ease of use.

## 📖 Documentation

- [Installation Guide](docs/installation.md)
- [Usage Guide](docs/usage.md)
- [Command Reference](docs/commands.md)
- [Examples](docs/examples.md)
- [FAQ](docs/faq.md)
- [Contributing](docs/contributing.md)

## 🎯 Why certwiz?

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

## 📄 JSON Output

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
```

## 🔥 Key Features in Detail

### Certificate Inspection
- View certificates from files (PEM/DER) or live websites
- Automatic format detection
- Shows all SANs with intelligent wrapping
- Highlights expiration status with color coding
- Displays full certificate chain with `--chain`
- Shows detailed extensions with `--full`
- Connect through proxies/tunnels with `--connect` flag

### Certificate Generation
- Create self-signed certificates instantly
- Support for multiple SANs (DNS names and IP addresses)
- Customizable validity period and key size
- Generates both certificate and private key files

### Certificate Verification
- Check certificate validity dates
- Verify hostname matches
- Validate against CA certificates
- Clear pass/fail status indicators

### Beautiful Output
- Color-coded status indicators (🟢 valid, 🟡 expiring soon, 🔴 expired)
- Clean, bordered tables for certificate information
- Smart terminal width detection and text wrapping
- Icons and symbols for better readability

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/contributing.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [HTTPie](https://httpie.io/) for its user-friendly approach
- Built with [Cobra](https://github.com/spf13/cobra) for CLI management
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss) for beautiful output

---

Made with ❤️ by the certwiz team