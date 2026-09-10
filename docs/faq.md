# Frequently Asked Questions

## General Questions

### What is certwiz?

certwiz is a user-friendly command-line tool for certificate management. It simplifies common certificate operations like inspection, generation, conversion, verification, and TLS version testing, making them as easy as HTTPie makes HTTP requests. The binary is called `cert`.

### Why not just use OpenSSL?

OpenSSL is powerful but complex. certwiz provides:
- Intuitive commands that are easy to remember
- Readable output with clear status indicators
- Smart defaults that just work
- JSON output for scripts and exit codes you can rely on

### Is certwiz a replacement for OpenSSL?

No, certwiz complements OpenSSL. It handles the most common certificate operations with a simpler interface. For advanced operations, you may still need OpenSSL.

### What platforms does certwiz support?

Release binaries are published for:
- macOS (Apple Silicon and Intel)
- Linux (x86_64, arm64, armv7, i386)
- FreeBSD (x86_64, arm64)
- Windows (x86_64, arm64, i386)

Anything else that Go 1.20+ targets can be built from source.

## Installation Issues

### "command not found" after installation

Make sure the binary is in your PATH:

```bash
# If installed with `make install`
export PATH=$PATH:$(go env GOPATH)/bin

# If installed manually
export PATH=$PATH:/usr/local/bin
```

Add the export line to your shell configuration (`~/.bashrc`, `~/.zshrc`, etc.)

### Permission denied during installation

Use sudo for system-wide installation:
```bash
sudo mv cert /usr/local/bin/
```

Or install to a user directory:
```bash
mkdir -p ~/.local/bin
mv cert ~/.local/bin/
export PATH=$PATH:~/.local/bin
```

### Can I use `go install`?

Not from the GitHub path. The module is declared as `certwiz`, so `go install github.com/trahma/certwiz@latest` fails with a module path mismatch. Use the installer script, a release archive, or clone the repository and run `make build` or `make install`.

### How do I update?

```bash
cert update
```

This downloads the installer script and replaces the binary in place. It is not available on Windows; download a new release from GitHub instead.

### Colors not showing in terminal

Ensure your terminal supports colors:
```bash
export TERM=xterm-256color
```

For Windows, use Windows Terminal or PowerShell 7+.

## Certificate Inspection

### How do I check if a certificate is expired?

```bash
cert inspect example.com
```

Look for the "Status" line - it will show:
- `Valid (X days remaining)` - Certificate is valid
- `EXPIRING SOON (X days remaining)` - Expires within 30 days
- `EXPIRED (X days ago)` - Certificate has expired

For scripts, use `cert inspect example.com --json | jq '.days_until_expiry'`.

### Can I inspect certificates on non-standard ports?

Yes, multiple ways:

```bash
cert inspect example.com:8443
cert inspect example.com --port 8443
cert inspect https://example.com:8443
```

### Can I inspect a certificate from stdin?

Yes. Use `-` as the target:

```bash
openssl s_client -connect example.com:443 -showcerts </dev/null 2>/dev/null | cert inspect - --chain
kubectl get secret tls -o jsonpath='{.data.tls\.crt}' | base64 -d | cert inspect -
```

### What happens with a bundle like fullchain.pem?

Every certificate in the file is parsed. The first one is displayed and a hint tells you how many more there are; add `--chain` to see all of them.

### Why does `cert inspect` say my file does not exist?

Anything that is not an existing file is normally treated as a hostname. Arguments that look like a path (they contain a directory separator, start with `.` or `~`, or end in a certificate extension such as `.pem`, `.crt`, `.der`, `.key`, `.csr`) are the exception: if the file is missing you get `certificate file does not exist: <path>` instead of a failed connection. Check the path and the current directory. Bare hostnames such as `example.com` are never treated as files.

### What does the --full flag show?

The `--full` flag displays detailed certificate extensions:
- Key Usage (what the certificate can be used for)
- Extended Key Usage (specific purposes)
- Basic Constraints (CA status)
- Subject Alternative Name summary
- Authority Info Access (OCSP, CA URLs)
- CRL Distribution Points
- Certificate Policies
- All other extensions, with critical ones marked

### How do I see the certificate chain?

```bash
cert inspect example.com --chain
```

This shows every certificate the server presented after the leaf.

### Where are the fingerprints?

SHA-256 and SHA-1 fingerprints are in the main table and in the JSON output as `fingerprint_sha256` and `fingerprint_sha1`.

### Why are some SANs on multiple lines?

certwiz wraps SANs based on terminal width. All SANs are shown, wrapped to fit your terminal. Use `--json` when you need them as a list.

### Does `inspect` have a connection timeout?

Yes. Remote inspection uses a 5-second timeout by default. Change it with `--timeout`, which accepts Go durations such as `2s` or `500ms`.

## Certificate Generation

### Can certwiz generate CA certificates?

Yes. Use `cert ca` to create a self-signed Certificate Authority, `cert csr` to create a request, and `cert sign` to sign the request with the CA.

### How do I add multiple domain names to a certificate?

Use multiple `--san` flags:

```bash
cert generate --cn example.com \
  --san example.com \
  --san www.example.com \
  --san "*.example.com" \
  --san api.example.com
```

### Can I generate certificates with IP addresses?

Yes, prefix IPs with `IP:`:

```bash
cert generate --cn server.local \
  --san server.local \
  --san IP:192.168.1.100 \
  --san IP:10.0.0.1 \
  --san IP:::1
```

`cert csr` and `cert sign` also accept `email:` and `uri:` prefixes.

### What key types are supported?

Generation currently produces RSA keys with configurable size (2048, 4096 bits). ECDSA generation is planned. Inspection, verification, and signing already handle ECDSA certificates and keys.

### Where are generated files saved?

By default in the current directory. Use `--output` to specify another directory:

```bash
cert generate --cn example.com --output /path/to/directory/
```

Files are named `{cn}.crt` and `{cn}.key` (`{cn}-ca.crt` and `{cn}-ca.key` for a CA, `{cn}.csr` for a request). For `ca` and `csr`, spaces and other awkward characters in the CN become underscores. Private keys are written with `0600` permissions.

## Certificate Conversion

### What formats does certwiz support?

- PEM (Privacy Enhanced Mail) - Base64 encoded
- DER (Distinguished Encoding Rules) - Binary format

### How does format detection work?

The input is parsed as PEM first; if it contains no PEM blocks it is parsed as DER. File extensions are not used.

### Can I convert certificate chains?

Converting to PEM keeps every certificate in the bundle. DER can hold only one certificate, so converting a bundle to DER is an error.

### What about PKCS#12/PFX format?

PKCS#12/PFX support is planned for a future release.

## Certificate Verification

### What does verification check?

Always:
- Certificate parses correctly
- Not before / not after dates

With `--host`: the hostname matches the CN or SANs, including wildcards.

With `--ca`: the certificate chains to the given CA file (PEM bundle or a single DER certificate) and its signature verifies.

With `--key`: the private key (PKCS#8, PKCS#1, or EC) matches the certificate's public key.

With `--expires-in`: the certificate is not within the given window of expiring (`30d`, `720h`, or a plain number of days).

Any failed check gives exit code 1.

### How do I verify against a custom CA?

```bash
cert verify server.crt --ca /path/to/ca-bundle.crt
```

### Can I verify self-signed certificates?

Without `--ca`, only the dates (and `--host`, `--key`, `--expires-in` if given) are checked, so a self-signed certificate passes. To check the signature too, provide the certificate itself as the CA:

```bash
cert verify self-signed.crt --ca self-signed.crt
```

### How do I check expiry in CI or cron?

```bash
cert verify /etc/ssl/certs/site.pem --expires-in 30d
```

The command exits 1 and prints `Certificate expires in N days (within the 30-day threshold)` on stderr when the certificate is inside the window. For remote hosts, use `cert inspect host --json | jq '.days_until_expiry'`. See [Examples](examples.md) for complete scripts.

## TLS Testing

### How do I see which TLS versions a server supports?

```bash
cert tls example.com
cert tls example.com --json
```

All four versions (1.0 to 1.3) are probed concurrently and the negotiated cipher suite is shown for each supported version. A version that fails while others succeed is retried once, so a server that limits concurrent handshakes is not misreported.

### Can certwiz check for weak ciphers?

`cert tls` reports the cipher suite negotiated for each TLS version and warns when TLS 1.0 or 1.1 are still enabled. It does not enumerate every cipher the server accepts; use `nmap --script ssl-enum-ciphers` or `sslyze` for a full cipher audit.

## Troubleshooting

### "failed to connect" errors

Check:
1. Network connectivity
2. Correct hostname/port
3. Firewall rules
4. TLS enabled on the target
5. The timeout (`--timeout 10s` for slow hosts)

If the target was meant to be a file, see "Why does `cert inspect` say my file does not exist?" above.

### "Chain verification failed"

`cert verify --ca` could not build a path from the certificate to the CA you supplied. Make sure the CA file contains the issuing intermediate as well as the root, or pass a bundle such as `fullchain.pem`.

### Garbled output or missing colors

Disable borders, colors, and emojis with the `--plain` flag, set `CI=1`, or turn them off permanently in the config file:

```bash
cert inspect example.com --plain
CI=1 cert inspect example.com
```

```yaml
# ~/.config/certwiz/config.yaml
output:
  colors: false
  emojis: false
```

### Where do errors go?

Human-readable errors are printed to stderr exactly once, and the exit code is 1. With `--json`, the error is a JSON payload on stdout (`{"success": false, "error": "..."}`) so scripts can parse it.

### How do I debug connection issues?

```bash
cert inspect problematic.site --full
cert inspect problematic.site --chain
cert inspect problematic.site --port 8443 --timeout 10s
cert tls problematic.site
```

## Integration

### Can I use certwiz in scripts?

Yes:
- Exit codes: 0 on success, 1 on any error or failed check
- `--json` for machine-readable output on every command
- `--plain` for stable text output
- Errors on stderr, results on stdout

Example:
```bash
#!/bin/bash
if [[ $(cert inspect example.com --json | jq '.is_expired') == "true" ]]; then
    echo "Certificate expired!"
    exit 1
fi
```

### Does certwiz support JSON output?

Yes. Add `--json` to any command.

### Can I use certwiz in CI/CD?

Yes. See [Examples](examples.md) for GitHub Actions and Jenkins integrations.

### Is there a Docker image?

Not an official one. A minimal image that uses the release binary:

```dockerfile
FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl \
 && curl -sL https://github.com/trahma/certwiz/releases/latest/download/cert-linux-x86_64.tar.gz \
    | tar xz -C /usr/local/bin cert
ENTRYPOINT ["cert"]
```

## Security

### Is it safe to use certwiz with production certificates?

Yes, certwiz:
- Only reads certificates unless you run a command that writes files (`generate`, `ca`, `csr`, `sign`, `convert`)
- Does not send data anywhere except to the host you ask it to connect to (and GitHub for `cert update`)
- Does not store certificates
- Writes private keys with `0600` permissions
- Is open source

### Does certwiz validate certificate security?

certwiz shows security-relevant information:
- Key type and size
- Signature algorithm
- Expiration status
- Certificate chain
- Fingerprints
- Negotiated TLS version and cipher suite for live connections

For security scanning, combine with other tools.

## Future Features

### What features are planned?

- ECDSA key generation
- PKCS#12/PFX support
- ACME/Let's Encrypt integration
- Certificate transparency log checking
- OCSP checks

### How can I request a feature?

Open an issue on GitHub with:
- Use case description
- Why it would be useful
- Examples of how it would work

### Can I contribute?

Yes. See the [Contributing](contributing.md) guide.

## Getting Help

### Where can I get help?

- GitHub Issues: bug reports and feature requests
- GitHub Discussions: questions
- This documentation

### How do I report a bug?

Open a GitHub issue with:
- `cert version` output
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Error messages (they are on stderr)

### Is there commercial support?

certwiz is community-supported open source.
