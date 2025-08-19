# Frequently Asked Questions

## General Questions

### What is certwiz?

certwiz is a user-friendly command-line tool for certificate management. It simplifies common certificate operations like inspection, generation, conversion, and verification - making them as easy as HTTPie makes HTTP requests.

### Why not just use OpenSSL?

OpenSSL is powerful but complex. certwiz provides:
- Intuitive commands that are easy to remember
- Beautiful, readable output
- Smart defaults that just work
- No need to remember complex syntax
- Colored output with clear status indicators

### Is certwiz a replacement for OpenSSL?

No, certwiz complements OpenSSL. It handles the most common certificate operations with a simpler interface. For advanced operations, you may still need OpenSSL.

### What platforms does certwiz support?

certwiz runs on:
- macOS (Intel and Apple Silicon)
- Linux (x86_64, ARM64)
- Windows (x86_64)
- Any platform that Go supports

## Installation Issues

### "command not found" after installation

Make sure the binary is in your PATH:

```bash
# If installed with Go
export PATH=$PATH:$(go env GOPATH)/bin

# If installed manually
export PATH=$PATH:/usr/local/bin
```

Add the export line to your shell configuration (`~/.bashrc`, `~/.zshrc`, etc.)

### Permission denied during installation

Use sudo for system-wide installation:
```bash
sudo mv certwiz /usr/local/bin/
```

Or install to user directory:
```bash
mkdir -p ~/.local/bin
mv certwiz ~/.local/bin/
export PATH=$PATH:~/.local/bin
```

### Colors not showing in terminal

Ensure your terminal supports colors:
```bash
export TERM=xterm-256color
```

For Windows, use Windows Terminal or PowerShell 7+.

## Certificate Inspection

### How do I check if a certificate is expired?

```bash
certwiz inspect example.com
```

Look for the "Status" line - it will show:
- `Valid (X days remaining)` - Certificate is valid
- `EXPIRING SOON (X days remaining)` - Expires within 30 days
- `EXPIRED (X days ago)` - Certificate has expired

### Can I inspect certificates on non-standard ports?

Yes, multiple ways:

```bash
certwiz inspect example.com:8443
certwiz inspect example.com --port 8443
certwiz inspect https://example.com:8443
```

### What does the --full flag show?

The `--full` flag displays detailed certificate extensions:
- Key Usage (what the certificate can be used for)
- Extended Key Usage (specific purposes)
- Basic Constraints (CA status)
- Authority Info Access (OCSP, CA URLs)
- Certificate Policies
- All other extensions

### How do I see the certificate chain?

```bash
certwiz inspect example.com --chain
```

This shows the complete chain from server certificate to root CA.

### Why are some SANs truncated?

certwiz intelligently wraps SANs based on terminal width. All SANs are shown, but wrapped to fit your terminal. Resize your terminal for different formatting.

## Certificate Generation

### Can certwiz generate CA certificates?

Currently, certwiz generates self-signed certificates. CA certificate generation is planned for a future release.

### How do I add multiple domain names to a certificate?

Use multiple `--san` flags:

```bash
certwiz generate --cn example.com \
  --san example.com \
  --san www.example.com \
  --san "*.example.com" \
  --san api.example.com
```

### Can I generate certificates with IP addresses?

Yes, prefix IPs with `IP:`:

```bash
certwiz generate --cn server.local \
  --san server.local \
  --san IP:192.168.1.100 \
  --san IP:10.0.0.1 \
  --san IP:::1
```

### What key types are supported?

Currently RSA keys with configurable size (2048, 4096 bits). ECDSA support is planned.

### Where are generated files saved?

By default in the current directory. Use `--output` to specify:

```bash
certwiz generate --cn example.com --output /path/to/directory/
```

Files are named: `{cn}.crt` and `{cn}.key`

## Certificate Conversion

### What formats does certwiz support?

certwiz supports:
- PEM (Privacy Enhanced Mail) - Base64 encoded
- DER (Distinguished Encoding Rules) - Binary format

### How does format detection work?

certwiz automatically detects format by:
1. Checking for PEM headers (`-----BEGIN`)
2. Trying to parse as PEM, then DER
3. Using file extensions as hints

### Can I convert certificate chains?

Currently, certwiz converts single certificates. Chain conversion is planned.

### What about PKCS#12/PFX format?

PKCS#12/PFX support is planned for a future release.

## Certificate Verification

### What does verification check?

Basic verification checks:
- Certificate structure validity
- Expiration dates
- Basic constraints

With `--host`:
- Hostname matches CN or SANs
- Wildcard matching

With `--ca`:
- Certificate chain validation
- Signature verification

### How do I verify against a custom CA?

```bash
certwiz verify server.crt --ca /path/to/ca-bundle.crt
```

### Can I verify self-signed certificates?

Yes, but verification will note that it's self-signed. For full validation, provide the same certificate as the CA:

```bash
certwiz verify self-signed.crt --ca self-signed.crt
```

## Troubleshooting

### "failed to connect" errors

Check:
1. Network connectivity
2. Correct hostname/port
3. Firewall rules
4. TLS/SSL enabled on target

### "certificate signed by unknown authority"

This is informational - certwiz still shows certificate details. To verify against a CA:

```bash
certwiz verify cert.pem --ca ca-bundle.crt
```

### Garbled output or missing colors

Check terminal compatibility:
```bash
# Force colors
FORCE_COLOR=1 certwiz inspect example.com

# Disable colors
NO_COLOR=1 certwiz inspect example.com
```

### How do I debug connection issues?

Use combination of flags:

```bash
# Full details
certwiz inspect problematic.site --full

# Check chain
certwiz inspect problematic.site --chain

# Try different port
certwiz inspect problematic.site --port 8443
```

## Integration

### Can I use certwiz in scripts?

Yes! certwiz is script-friendly:
- Standard exit codes (0=success, 1=error)
- Parseable output
- Pipe-friendly

Example:
```bash
#!/bin/bash
if certwiz inspect example.com | grep -q "EXPIRED"; then
    echo "Certificate expired!"
    exit 1
fi
```

### Does certwiz support JSON output?

JSON output is planned for a future release. Currently, output is human-readable text.

### Can I use certwiz in CI/CD?

Yes! See [Examples](examples.md) for GitHub Actions and Jenkins integration examples.

### Is there a Docker image?

Docker image is planned. For now, you can create your own:

```dockerfile
FROM golang:alpine
RUN go install github.com/certwiz/certwiz@latest
ENTRYPOINT ["certwiz"]
```

## Security

### Is it safe to use certwiz with production certificates?

Yes, certwiz:
- Only reads certificates (never modifies without explicit command)
- Doesn't send data anywhere
- Doesn't store certificates
- Open source for audit

### Does certwiz validate certificate security?

certwiz shows security-relevant information:
- Key size
- Signature algorithms  
- Expiration status
- Certificate chain

For security scanning, combine with other tools.

### Can certwiz check for weak ciphers?

certwiz focuses on certificates, not cipher suites. Use tools like `nmap` or `sslyze` for cipher analysis.

## Future Features

### What features are planned?

Roadmap includes:
- ECDSA key support
- CA certificate generation
- PKCS#12/PFX support
- JSON output
- Certificate signing (CSR)
- ACME/Let's Encrypt integration
- Database of trusted CAs
- Certificate transparency log checking

### How can I request a feature?

Open an issue on GitHub with:
- Use case description
- Why it would be useful
- Examples of how it would work

### Can I contribute?

Yes! See [Contributing](contributing.md) guide.

## Getting Help

### Where can I get help?

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and discussions
- Documentation: This comprehensive guide
- Discord: Community chat (if available)

### How do I report a bug?

Open a GitHub issue with:
- certwiz version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Error messages

### Is there commercial support?

Currently certwiz is community-supported open source. Commercial support may be available in the future.