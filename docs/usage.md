# Usage Guide

This guide covers the basic usage of certwiz commands with practical examples.
For every flag and default, see the [Command Reference](commands.md).

## Commands

- `inspect` - View certificate information from a file, stdin, or a live server
- `generate` - Create a self-signed certificate
- `csr` - Create a Certificate Signing Request and key
- `ca` - Create a Certificate Authority
- `sign` - Sign a CSR with your CA
- `convert` - Convert between PEM and DER
- `verify` - Validate certificates
- `tls` - Test TLS version support
- `update` - Update cert to the latest release

Every command accepts `--json` for machine-readable output and `--plain` for
output without borders, colors, or emojis.

## Inspecting Certificates

### Inspect a Website

The simplest use case - check a website's certificate:

```bash
cert inspect google.com
```

Output shows:
- Subject and Issuer
- Validity dates and expiration status
- Public key type and signature algorithm
- SHA-256 and SHA-1 fingerprints
- Negotiated TLS version and cipher suite
- Subject Alternative Names (SANs)

### Inspect a Certificate File

```bash
# PEM format
cert inspect server.crt

# DER format
cert inspect certificate.der

# Format is detected from the content, not the extension
cert inspect mycert.pem
```

If a path-like argument does not exist (for example `./sever.crt` with a
typo), cert reports a missing file instead of trying it as a hostname.

### Bundles and stdin

A file may contain several certificates, such as the `fullchain.pem` that
ACME clients produce. The first certificate is shown; add `--chain` to see
the rest:

```bash
cert inspect fullchain.pem --chain
```

Use `-` to read from stdin, which pairs well with openssl or curl:

```bash
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | cert inspect -
curl -s https://example.com/ca.pem | cert inspect -
```

### Custom Ports

```bash
# Specify port explicitly
cert inspect example.com:8443

# Or use the --port flag
cert inspect example.com --port 8443
```

### View Certificate Chain

See the complete trust chain as presented by the server:

```bash
cert inspect github.com --chain
```

This shows:
- Server certificate
- Intermediate certificates
- The root, if the server sends it

### Detailed Extension Information

View all certificate extensions with human-readable formatting:

```bash
cert inspect google.com --full
```

Shows:
- Key Usage flags
- Extended Key Usage
- Basic Constraints
- Authority Info Access URLs
- CRL Distribution Points
- Certificate Policies
- Any other extensions, with `[CRITICAL]` markers

### Combine Options

```bash
# View everything
cert inspect example.com --full --chain

# Check a specific port with full details
cert inspect api.example.com:8443 --full

# Give a slow server more time
cert inspect slow.example.com --timeout 15s
```

### Advanced Inspection Options

#### Inspect Through Proxy or Tunnel

Connect to a different host while requesting the certificate for the target
hostname (SNI):

```bash
# Through SSH tunnel
cert inspect api.internal.com --connect localhost:8080

# Through proxy
cert inspect prod.service --connect proxy.local:443

# Different port
cert inspect backend.local --connect 127.0.0.1:3000 --port 443
```

#### Force Certificate Type Selection

For servers with both ECDSA and RSA certificates, force selection:

```bash
# Force ECDSA certificate (if available)
cert inspect cloudflare.com --sig-alg ecdsa

# Force RSA certificate (if available)
cert inspect cloudflare.com --sig-alg rsa

# Auto selection (default - server chooses)
cert inspect cloudflare.com --sig-alg auto
```

Note: `ecdsa` and `rsa` cap the connection at TLS 1.2, because TLS 1.3 does
not select certificates by cipher suite.

## Generating Certificates

### Basic Self-Signed Certificate

```bash
cert generate --cn myapp.local
```

Creates:
- `myapp.local.crt` - Certificate file
- `myapp.local.key` - Private key file, written with `0600` permissions

### With Subject Alternative Names (SANs)

```bash
cert generate --cn myapp.local \
  --san myapp.local \
  --san "*.myapp.local" \
  --san localhost \
  --san IP:127.0.0.1 \
  --san IP:192.168.1.100
```

IP addresses need the `IP:` prefix; without it the value is treated as a DNS
name.

### Custom Validity Period

```bash
# Valid for 2 years
cert generate --cn myapp.local --days 730

# Valid for 90 days (Let's Encrypt style)
cert generate --cn myapp.local --days 90
```

### Custom Key Size

```bash
# 4096-bit RSA key
cert generate --cn myapp.local --key-size 4096

# Default is 2048-bit
cert generate --cn myapp.local --key-size 2048
```

### Specify Output Directory

```bash
cert generate --cn myapp.local --output /etc/ssl/certs/
```

## Running Your Own CA

Create a CA once, then sign CSRs with it. This is the usual path for internal
services and development environments where browsers and clients can be told
to trust your CA.

```bash
# 1. Create the CA (key size defaults to 4096, validity to 10 years)
cert ca --cn "Example Internal CA" --org "Example Corp"

# 2. Create a CSR and key for a service
cert csr --cn api.internal.example --san api.internal.example --san IP:10.0.0.5

# 3. Sign it
cert sign --csr api.internal.example.csr \
  --ca Example_Internal_CA-ca.crt \
  --ca-key Example_Internal_CA-ca.key \
  --days 365

# 4. Check the result against the CA
cert verify api.internal.example.crt --ca Example_Internal_CA-ca.crt --host api.internal.example
```

Spaces and other unsafe characters in the CA name become `_` in the file
names. `csr` and `sign` accept `email:` and `uri:` SANs as well as DNS and
`IP:` entries.

## Converting Certificates

### PEM to DER

```bash
cert convert certificate.pem certificate.der --format der
```

### DER to PEM

```bash
cert convert certificate.der certificate.pem --format pem
```

### Input Format Detection and Bundles

The input format is detected from the content, so any extension works:

```bash
cert convert input.crt output.der --format der
cert convert input.der output.pem --format pem
```

PEM output keeps every certificate in a bundle. DER holds a single
certificate, so converting a bundle to DER is an error; split the bundle
first if you need DER.

## Verifying Certificates

### Basic Verification

```bash
cert verify server.crt
```

Checks that the file parses and that the certificate is within its validity
period.

### Verify Against Hostname

```bash
cert verify server.crt --host example.com
```

Verifies that the hostname matches the certificate's SANs (or Common Name
for legacy certificates), including wildcards.

### Verify Against CA

```bash
cert verify server.crt --ca ca-bundle.crt
```

Validates that a trust chain can be built to a certificate in the CA file.
The CA file may be PEM (including a bundle) or DER.

### Check the Private Key Matches

```bash
cert verify server.crt --key server.key
```

Confirms the key belongs to the certificate. PKCS#8, PKCS#1 (RSA), and SEC1
(EC) keys are accepted, PEM or DER encoded.

### Fail Before Expiry

```bash
# Exit 1 if the certificate expires within 30 days
cert verify server.crt --expires-in 30d
```

Accepts days (`30d` or `30`) or any Go duration (`720h`). This is the
building block for renewal alerts in CI and cron.

### Combined Verification

```bash
cert verify server.crt \
  --host api.example.com \
  --ca /etc/ssl/certs/ca-bundle.crt \
  --key server.key \
  --expires-in 14d
```

Any failing check makes the command exit 1.

## Testing TLS Versions

### Check Supported TLS Versions

```bash
cert tls google.com
cert tls https://example.com
```

This shows:
- Which of TLS 1.0, 1.1, 1.2, and 1.3 are supported, with the cipher suite negotiated for each
- Minimum and maximum supported versions
- A security warning if TLS 1.0 or 1.1 is still enabled

The four versions are probed concurrently, and a version that fails while
others succeed is retried once, so a busy server is not misreported.

### Custom Port and Timeout

```bash
# Test non-standard port
cert tls api.example.com:8443

# With custom timeout (applies to each handshake)
cert tls slow-server.example.com --timeout 10s
```

### JSON Output for Scripts

```bash
# Get structured data
cert tls example.com --json | jq .

# Print the newest supported version
cert tls example.com --json | jq -r '.max_supported'

# Fail if deprecated versions are enabled
cert tls example.com --json \
  | jq -e '[.versions[] | select(.name == "TLS 1.0" or .name == "TLS 1.1") | .supported] | any' >/dev/null \
  && echo "FAIL: deprecated TLS enabled"
```

## Understanding the Output

### Color Coding

- Green: valid, healthy, good
- Yellow: warning, expiring soon (less than 30 days)
- Red: error, expired, critical issue
- Blue: informational, neutral

Borders are colored by the certificate's status, so an expired certificate's
panel is red.

### Status Messages

```
Valid (365 days remaining)         # Healthy certificate
EXPIRING SOON (15 days remaining)  # Needs renewal soon
EXPIRED (10 days ago)              # Certificate has expired
```

### Symbols

In a normal terminal, checks are shown with a check mark, failures with a
cross, and details with an arrow. In CI environments and with `--plain`
these become `[OK]`, `[X]`, and `->` so the output is safe to copy and grep.
Extensions that must be understood by a client are marked `[CRITICAL]`.

### Plain Mode and Configuration

Use `--plain` for output without borders, colors, or emojis:

```bash
cert inspect google.com --plain
cert tls github.com --plain
```

To make that the default, or to turn off only some decoration, create
`~/.config/certwiz/config.yaml` (or `~/.certwiz.yaml`):

```yaml
output:
  plain: false
  borders: true
  colors: true
  emojis: true
```

`--plain` overrides the config file, which overrides CI detection.

### Errors and Exit Codes

Every command exits 0 on success and 1 on failure. Human-readable errors go
to stderr; with `--json` the error is a JSON object on stdout:

```bash
cert inspect missing.pem
# stderr: Error: certificate file does not exist: missing.pem

cert inspect missing.pem --json
# stdout: {"success": false, "error": "certificate file does not exist: missing.pem"}
```

## JSON Output

Use `--json` to integrate with scripts and tools:

```bash
# Inspect with JSON
cert inspect google.com --json | jq '.subject.common_name'

# Days until expiry
cert inspect cert.pem --json | jq '.days_until_expiry'

# Fingerprint for pinning
cert inspect example.com --json | jq -r '.fingerprint_sha256'

# Verify with JSON
cert verify server.crt --host example.com --json | jq '.is_valid'

# Generate and consume file paths
cert generate --cn test.local --json | jq -r '.files[]'
```

## Tips and Tricks

### Quick Domain Check

```bash
# Check multiple domains quickly
for domain in google.com github.com cloudflare.com; do
  echo "=== $domain ==="
  cert inspect "$domain" --plain | grep -E "Status|Valid"
done
```

### Save a Report

```bash
cert inspect example.com --full --chain --plain > example.com.info.txt
```

### Check Internal Services

```bash
# Check internal service with self-signed cert
cert inspect internal.service.local:8443

# Verify against internal CA
cert verify internal.crt --ca /path/to/internal-ca.crt
```

### Batch Certificate Generation

```bash
# Generate certificates for multiple domains
for domain in app1.local app2.local app3.local; do
  cert generate --cn "$domain" --san "$domain" --san "*.$domain"
done
```

### Certificate Monitoring

```bash
#!/bin/bash
# Alert when any certificate expires within 14 days
domains=("example.com" "api.example.com" "www.example.com")

for domain in "${domains[@]}"; do
  days=$(cert inspect "$domain" --json | jq '.days_until_expiry')
  if [[ -z "$days" || "$days" -lt 14 ]]; then
    echo "ALERT: $domain expires in ${days:-?} days"
  fi
done
```

For certificates on disk, `cert verify --expires-in 14d` does the same with
an exit code and no parsing.

## Common Workflows

### Setting Up Local Development

```bash
# 1. Generate a certificate for local development
cert generate --cn myapp.local \
  --san myapp.local \
  --san "*.myapp.local" \
  --san localhost \
  --san IP:127.0.0.1

# 2. Move to appropriate directory
sudo mv myapp.local.crt /usr/local/etc/ssl/certs/
sudo mv myapp.local.key /usr/local/etc/ssl/private/

# 3. Verify the certificate
cert verify /usr/local/etc/ssl/certs/myapp.local.crt --host myapp.local
```

### Debugging SSL Issues

```bash
# 1. Check the problematic certificate
cert inspect problematic-site.com --full

# 2. View the certificate chain
cert inspect problematic-site.com --chain

# 3. Check which TLS versions are offered
cert tls problematic-site.com

# 4. Check specific port if non-standard
cert inspect problematic-site.com:8443

# 5. Save details for analysis
cert inspect problematic-site.com --full --chain --plain > cert-analysis.txt
```

### Certificate Renewal Process

```bash
# 1. Check current certificate
cert inspect mysite.com

# 2. Generate new certificate
cert generate --cn mysite.com --san mysite.com --san www.mysite.com

# 3. Verify new certificate and key
cert verify mysite.com.crt --host mysite.com --key mysite.com.key

# 4. Convert if needed
cert convert mysite.com.crt mysite.com.der --format der
```

## Next Steps

- Explore [Command Reference](commands.md) for all options
- See [Examples](examples.md) for real-world scenarios
- Read [FAQ](faq.md) for common questions
