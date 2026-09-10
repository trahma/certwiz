# Command Reference

Complete reference for all certwiz commands and options. The `cert --help` and
`cert <command> --help` output is the source of truth; this page mirrors it and
adds detail.

## Global Options

These options work with every command:

```
    --json      Output machine-readable JSON
    --plain     Output in plain format (no borders, colors, or emojis)
-h, --help      Show help for any command
```

The root command also accepts `-v, --version` to print the version.

## inspect

Inspect a certificate from a file, URL, or stdin.

### Synopsis

```bash
cert inspect [file|url|-] [flags]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--full` | Show full certificate details including extensions | `false` |
| `--chain` | Show certificate chain (URLs and multi-certificate files) | `false` |
| `--port` | Port for remote inspection | `443` |
| `--connect` | Connect to a different host (e.g. `localhost:8080`) while validating the cert for the target hostname | |
| `--timeout` | Network timeout for remote inspection, as a duration (e.g. `5s`, `500ms`) | `5s` |
| `--sig-alg` | Preferred signature algorithm: `auto`, `ecdsa`, or `rsa` (TLS 1.2 only) | `auto` |

### Arguments

- `target` - Certificate file path, URL/domain name, or `-` for stdin (required)

How the target is classified:

- An existing file is read and parsed (PEM or DER). Files may contain several
  certificates, such as `fullchain.pem`; the first is shown and `--chain`
  shows the rest.
- `-` reads from stdin.
- Anything that looks like a path but does not exist (contains a directory
  separator, starts with `.` or `~`, or ends in a certificate extension such as
  `.pem`, `.crt`, `.cer`, `.der`, `.key`, `.csr`, `.p7b`, `.pfx`, `.p12`) is
  reported as a missing file rather than attempted as a hostname.
- Everything else is treated as a hostname or URL. A port may be given as
  `host:port`, in the URL, or with `--port`.

### Examples

```bash
# Inspect local certificate file
cert inspect server.crt
cert inspect /path/to/certificate.pem

# Inspect a bundle (fullchain.pem) - use --chain to see all certificates
cert inspect fullchain.pem --chain

# Read from stdin
openssl s_client -connect example.com:443 </dev/null | cert inspect -

# Inspect remote certificate
cert inspect google.com
cert inspect https://example.com
cert inspect api.example.com:8443

# With options
cert inspect google.com --full
cert inspect github.com --chain
cert inspect example.com --full --chain
cert inspect internal.service --port 8443
cert inspect slow.example.com --timeout 10s

# Through proxy or tunnel
cert inspect api.example.com --connect localhost:8080
cert inspect prod.internal --connect tunnel.local --port 443
cert inspect backend.local --connect 127.0.0.1:3000

# Force specific signature algorithm (for servers with both ECDSA and RSA certs)
cert inspect cloudflare.com --sig-alg ecdsa  # Forces ECDSA certificate
cert inspect cloudflare.com --sig-alg rsa    # Forces RSA certificate
cert inspect cloudflare.com --sig-alg auto   # Let server choose (default)
```

### Signature Algorithm Flag Usage

The `--sig-alg` flag controls which cipher suites are advertised in the TLS ClientHello:
- `auto` (default): Uses all available cipher suites, server chooses based on preference
- `ecdsa`: Only advertises ECDSA-compatible cipher suites, forcing ECDSA certificate if available
- `rsa`: Only advertises RSA-compatible cipher suites, forcing RSA certificate if available

Notes:
- `ecdsa` and `rsa` cap the connection at TLS 1.2, because TLS 1.3 does not select certificates by cipher suite
- Servers must have both ECDSA and RSA certificates configured for this to have an effect
- Useful for testing dual-certificate configurations and debugging certificate selection issues

### Connect Flag Usage

The `--connect` flag is useful for:
- Testing certificates through SSH tunnels
- Inspecting certificates behind proxies
- Validating certificates in local development environments
- Checking certificates on different servers with the same hostname

When using `--connect`:
- The connection is made to the host specified in `--connect`
- The certificate is requested for the original target hostname (SNI)
- Port can be specified in the connect host (e.g. `localhost:8080`) or via `--port`
- If port is in both, the one in `--connect` takes precedence

### Output Details

The inspect command shows:
- **Subject**: Certificate subject DN
- **Issuer**: Certificate issuer DN
- **Serial Number**: Unique certificate identifier (hex)
- **Valid From/To**: Certificate validity period
- **Status**: Current validity status with days remaining
- **Public Key**: Key type and size
- **Signature Algorithm**: Algorithm used to sign the certificate
- **SHA-256 / SHA-1 Fingerprint**: Certificate fingerprints, colon-separated uppercase hex
- **TLS Version / Cipher Suite**: The negotiated protocol and cipher (remote inspection only)
- **SANs**: All Subject Alternative Names (DNS, IP, email, URI). When there are more than ten, the row starts with a `(N total)` line.

Long values such as fingerprints and SAN lists wrap onto continuation lines
aligned with the value column. Example (`--plain`):

```
Subject            : CN=test.example.com, O=Test, C=US
Issuer             : CN=test.example.com, O=Test, C=US
Serial Number      : 52533095fd01b8b34b9c811c3b1cfc646e951600
Valid From         : 2026-01-07 04:17:14 UTC
Valid To           : 2027-01-07 04:17:14 UTC
Status             : Valid (119 days remaining)
Public Key         : RSA 2048 bits
Signature Algorithm: SHA256-RSA
SHA-256 Fingerprint: 1B:57:44:F5:A5:BB:71:90:02:A3:C9:67:98:57:82:3D:57
                     38:FD:63:88:18:DE:13:9D:1D:40:EF:B3:71:28:84
SHA-1 Fingerprint  : D5:2E:1B:A9:2B:F2:DD:26:8A:0F:DB:2A:3A:02:56:42:9B
                     3E:7A:A1
SANs               : test.example.com, *.test.example.com, 127.0.0.1
```

With `--full`, a "Certificate Extensions" section follows:
- **Key Usage**: Permitted key usage flags (for example "Digital Signature", "Certificate Sign")
- **Extended Key Usage**: Extended usage purposes (for example "Server Authentication", "Client Authentication"); unknown OIDs are listed as-is
- **Basic Constraints**: CA status and path length
- **Subject Alternative Name**: A count of SANs by type
- **Authority Info Access**: OCSP and CA issuer URLs
- **CRL Distribution Points**: Certificate revocation list URLs
- **Certificate Policies**: Policy OIDs with well-known names (for example "Domain Validated")
- **Other Extensions**: Any additional extensions by name or OID

Extensions marked critical carry a `[CRITICAL]` label. The usage labels are
the same strings used in JSON output.

With `--chain`:
- Each additional certificate, from the server certificate's issuer upwards
- Subject, issuer, validity dates and status for each

### JSON Output

`cert inspect --json` emits one object. Fields:

| Field | Description |
|-------|-------------|
| `subject`, `issuer` | Objects with `common_name`, `organization`, `organizational_unit`, `country`, `province`, `locality`, `street_address`, `postal_code` (arrays except `common_name`) |
| `serial_number` | Hex string |
| `not_before`, `not_after` | RFC 3339 timestamps |
| `is_ca`, `is_expired` | Booleans |
| `days_until_expiry` | Integer (negative when expired) |
| `signature_algorithm`, `public_key_algorithm`, `public_key_size` | Strings and key size in bits |
| `fingerprint_sha256`, `fingerprint_sha1` | Colon-separated uppercase hex |
| `dns_names`, `ip_addresses`, `email_addresses`, `uris` | SAN arrays (omitted when empty) |
| `key_usage`, `ext_key_usage` | Arrays of usage names (omitted when empty) |
| `source`, `format` | Where the certificate came from and `PEM` or `DER` |
| `tls_version`, `cipher_suite` | Negotiated values (remote inspection only) |
| `chain` | With `--chain`: array of `{subject, issuer, not_before, not_after, is_expired, serial_number}` |

## generate

Generate a self-signed certificate.

### Synopsis

```bash
cert generate [flags]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--cn` | Common Name for the certificate (required) | |
| `--san` | Subject Alternative Name: a DNS name or `IP:<address>` (repeatable) | |
| `--days` | Validity period in days | `365` |
| `--key-size` | RSA key size in bits | `2048` |
| `--output` | Output directory | `.` |

### SAN Format

SANs can be specified as:
- DNS names: `--san example.com`
- Wildcards: `--san "*.example.com"`
- IP addresses: `--san IP:192.168.1.1` (the `IP:` prefix is required; a bare address is treated as a DNS name)

`generate` accepts only DNS and IP SANs. Use `csr` and `sign` for email and URI SANs.

### Examples

```bash
# Basic certificate
cert generate --cn myapp.local

# With multiple SANs
cert generate --cn myapp.local \
  --san myapp.local \
  --san "*.myapp.local" \
  --san localhost \
  --san IP:127.0.0.1

# Custom validity and key size
cert generate --cn secure.app \
  --days 730 \
  --key-size 4096

# Output to specific directory
cert generate --cn myapp.local \
  --output /etc/ssl/certs/
```

### Output Files

The generate command creates:
- `{cn}.crt` - Certificate file in PEM format
- `{cn}.key` - Private key file in PEM (PKCS#8) format

Private key files are written with `0600` permissions. Serial numbers are
random 128-bit values. The certificate has `Digital Signature` and
`Key Encipherment` key usage and the `Server Authentication` extended key
usage.

## csr

Generate a Certificate Signing Request (CSR) and private key.

### Synopsis

```bash
cert csr [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--cn` | | Common Name (required) | |
| `--org` | | Organization | |
| `--org-unit` | | Organizational Unit | |
| `--country` | | Country (2-letter code) | |
| `--state` | | State or Province | |
| `--locality` | | Locality or City | |
| `--email` | | Email Address | |
| `--san` | | Subject Alternative Name: DNS name, `IP:<address>`, `email:<address>`, or `uri:<uri>` (repeatable) | |
| `--key-size` | `-k` | RSA key size in bits | `2048` |
| `--output` | `-o` | Output directory for CSR and key files | `.` |

### Examples

```bash
# Basic CSR generation
cert csr --cn example.com

# CSR with organization details
cert csr --cn example.com --org "Example Inc" --country US --state CA

# CSR with Subject Alternative Names
cert csr --cn example.com --san example.com --san www.example.com --san IP:10.0.0.5

# CSR with custom output directory and key size
cert csr --cn secure.example.com --key-size 4096 --output /etc/ssl/
```

### Output Files

- `{cn}.csr` - PEM-encoded certificate request
- `{cn}.key` - PEM (PKCS#8) private key, written with `0600` permissions

Characters that are unsafe in file names (`/ \ : * ? " < > |` and spaces) are
replaced with `_`. In non-JSON mode the parsed CSR is displayed after
generation.

## ca

Create a self-signed Certificate Authority (CA) certificate and private key.

### Synopsis

```bash
cert ca [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--cn` | | Common Name for the CA (required) | |
| `--org` | | Organization name | |
| `--country` | | Country (2-letter code) | |
| `--days` | `-d` | Validity period in days | `3650` |
| `--key-size` | `-k` | RSA key size in bits | `4096` |
| `--output` | `-o` | Output directory for CA files | `.` |

### Examples

```bash
# Create a basic CA certificate
cert ca --cn "My Company CA"

# Create a CA with organization details
cert ca --cn "Example Corp Root CA" --org "Example Corporation" --country US

# Create a CA with custom validity period (10 years)
cert ca --cn "Internal CA" --days 3650

# Create a CA with larger key size for extra security
cert ca --cn "Secure CA" --key-size 4096 --output /etc/pki/
```

### Output Files

- `{cn}-ca.crt` - PEM-encoded CA certificate (`IsCA` set, no path length constraint)
- `{cn}-ca.key` - PEM (PKCS#8) private key, written with `0600` permissions

File names are sanitized the same way as for `csr`, so `"My Company CA"`
produces `My_Company_CA-ca.crt`. Keep the CA key offline or otherwise
protected; anyone holding it can issue certificates your clients trust.

## sign

Sign a Certificate Signing Request with a CA.

### Synopsis

```bash
cert sign [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--csr` | | Path to the CSR file to sign (required) | |
| `--ca` | | Path to the CA certificate (required); PEM or DER | |
| `--ca-key` | | Path to the CA private key (required); PKCS#8, PKCS#1 (RSA), or SEC1 (EC) | |
| `--days` | `-d` | Validity period in days | `365` |
| `--output` | `-o` | Output directory for the signed certificate | `.` |
| `--san` | | Subject Alternative Name: DNS name, `IP:<address>`, `email:<address>`, or `uri:<uri>` (repeatable; overrides the CSR SANs if specified) | |

### Examples

```bash
# Sign a CSR with a CA
cert sign --csr server.csr --ca ca.crt --ca-key ca.key

# Sign with custom validity period (1 year)
cert sign --csr server.csr --ca ca.crt --ca-key ca.key --days 365

# Sign and output to specific directory
cert sign --csr server.csr --ca ca.crt --ca-key ca.key --output /etc/ssl/certs/

# Sign with different SANs (replaces the SANs in the CSR)
cert sign --csr server.csr --ca ca.crt --ca-key ca.key --san server.local --san "*.server.local"
```

### Behaviour

- The CSR signature is verified before signing
- The output file is named after the CSR (`server.csr` or `server.req` becomes `server.crt`)
- The signed certificate carries `Digital Signature` and `Key Encipherment` key usage, and `Server Authentication` plus `Client Authentication` extended key usage
- When `--san` is given, the CSR's SANs are ignored entirely

## convert

Convert a certificate file between PEM and DER formats.

### Synopsis

```bash
cert convert <input> <output> [flags]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Output format (`pem` or `der`) | `pem` |

### Arguments

- `input` - Input certificate file (required)
- `output` - Output certificate file (required)

### Examples

```bash
# PEM to DER
cert convert certificate.pem certificate.der --format der

# DER to PEM
cert convert certificate.der certificate.pem --format pem

# Input format is detected automatically
cert convert input.crt output.der --format der
```

### Format Detection and Bundles

The input format is detected from the content, not the file extension: data
containing PEM blocks is treated as PEM, anything else as DER.

PEM output keeps every certificate in a bundle. DER can hold only one
certificate, so converting a multi-certificate file to DER is an error.

With `--json` the result is `{"success": true, "message": "Converted from PEM to DER", "files": ["certificate.der"]}`.

## verify

Verify a certificate's validity, expiration, and optionally check hostname
matching, CA chain validation, private key matching, and upcoming expiry.

### Synopsis

```bash
cert verify <certificate> [flags]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--host` | Hostname to verify against the certificate | |
| `--ca` | CA certificate file (PEM or DER) for chain verification | |
| `--key` | Private key file to check against the certificate | |
| `--expires-in` | Fail if the certificate expires within this window (e.g. `30d`, `720h`) | |

### Arguments

- `certificate` - Certificate file to verify (required)

### Examples

```bash
# Basic verification
cert verify server.crt

# Verify hostname match
cert verify server.crt --host example.com

# Verify against CA
cert verify server.crt --ca ca-bundle.crt

# Check that a private key matches the certificate
cert verify server.crt --key server.key

# Fail (exit 1) if the certificate expires within 30 days - useful in CI/cron
cert verify server.crt --expires-in 30d

# Complete verification
cert verify server.crt \
  --host api.example.com \
  --ca /etc/ssl/certs/ca-bundle.crt \
  --key server.key
```

### Verification Checks

Always:
- The file parses as a PEM or DER certificate
- The certificate is within its validity period (not before / not after)

With `--host`:
- The hostname matches the SANs (or the Common Name for legacy certificates), including wildcards such as `*.example.com`

With `--ca`:
- A trust chain can be built from the certificate to a certificate in the CA file
- Signatures along the chain verify
- When `--host` is also given, the chain is verified for that name

With `--key`:
- The private key's public half matches the certificate's public key
- Supports PKCS#8, PKCS#1 (RSA), and SEC1 (EC) keys, PEM or DER encoded

With `--expires-in`:
- Fails verification if the certificate expires within the given window
- Accepts days (`30d` or `30`) or any Go duration (`720h`, `24h30m`)

### Output and Exit Code

The terminal output lists errors, warnings, and a "Validation Checks" table
with PASS or FAIL per check. Any failing check makes the command exit 1 and
print `Error: verification failed` on stderr. With `--json` the output is:

| Field | Description |
|-------|-------------|
| `is_valid` | `true` only if every check passed |
| `errors`, `warnings` | Arrays of messages (omitted when empty) |
| `key_matches` | Present only when `--key` was given |
| `certificate` | The same object `cert inspect --json` produces |

## tls

Test which TLS versions a remote server supports.

### Synopsis

```bash
cert tls <hostname> [flags]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Port for TLS testing | `443` |
| `--timeout` | Network timeout per handshake, as a duration (e.g. `5s`, `2s`) | `5s` |

### Arguments

- `hostname` - Target hostname, `host:port`, or URL (scheme and path are ignored) (required)

### Examples

```bash
# Test TLS versions for a domain
cert tls google.com
cert tls https://example.com

# Test with custom port
cert tls api.example.com:8443
cert tls internal.service --port 443

# Test with custom timeout
cert tls slow-server.example.com --timeout 10s
```

### How It Works

Each of TLS 1.0, 1.1, 1.2, and 1.3 is probed with a handshake pinned to
that single version. The four probes run concurrently, so an unreachable
host fails after one timeout rather than four. If at least one version
succeeded, any version that failed is retried once sequentially before being
reported as unsupported; this guards against servers or middleboxes that
limit concurrent handshakes from one client.

### Output Details

Example output (`--plain`):

```
 TLS Version Support for google.com:443

TLS 1.0: [OK] Supported (TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA)
TLS 1.1: [OK] Supported (TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA)
TLS 1.2: [OK] Supported (TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
TLS 1.3: [OK] Supported (TLS_AES_128_GCM_SHA256)

Summary

  -> Minimum supported version: TLS 1.0
  -> Maximum supported version: TLS 1.3

[!] Security Warning:
  -> TLS 1.0 is enabled but deprecated
  -> TLS 1.1 is enabled but deprecated

Recommendation: Consider disabling TLS 1.0 and TLS 1.1 for improved security.
```

In the default mode the same information is shown in a bordered panel with
check marks and colors. Each supported version shows the cipher suite that
was negotiated.

### JSON Output

```bash
# Get structured TLS version data
cert tls google.com --json | jq .

# Check the newest supported version
cert tls example.com --json | jq -r '.max_supported'

# Fail if deprecated versions are enabled
cert tls example.com --json \
  | jq -e '[.versions[] | select(.name == "TLS 1.0" or .name == "TLS 1.1") | .supported] | any' >/dev/null \
  && echo "FAIL: deprecated TLS enabled"
```

Fields: `host`, `port`, `versions` (array of `{version, name, supported, error, cipher_suite}` where `version` is the hex protocol number such as `0x0303`), `min_supported`, `max_supported` (empty strings when nothing is supported).

### Use Cases

- Security auditing: verify servers do not support deprecated TLS versions
- Compliance checking: ensure servers meet TLS requirements
- Migration testing: verify servers after TLS configuration changes
- Troubleshooting: diagnose TLS compatibility issues

## update

Update cert to the latest version.

### Synopsis

```bash
cert update [flags]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--force` | Force update even if already on latest version | `false` |

### Description

The update command downloads the installer script to a private temporary
file and runs it. The installer:

1. Checks the latest release on GitHub against your current version
2. Downloads and installs the new binary if it is newer (or always, with `--force`)
3. Detects your installation directory and backs up the existing binary as `cert.backup` alongside it

The temporary installer script is removed when the installer exits.

### Examples

```bash
# Check for and install updates
cert update

# Force reinstall current version (useful for fixing corrupted installations)
cert update --force
```

### Notes

- Not available on Windows. Windows users should download the latest release from the releases page.
- If the update fails, restore `cert.backup` manually.

## version

Show the version of cert.

### Synopsis

```bash
cert version
cert --version
```

### Examples

```bash
cert version
# Output: cert version 0.4.1
```

## completion

Generate shell completion scripts.

### Synopsis

```bash
cert completion [bash|zsh|fish|powershell]
```

### Examples

```bash
# Bash
cert completion bash > /etc/bash_completion.d/cert

# Zsh
cert completion zsh > "${fpath[1]}/_cert"

# Fish
cert completion fish > ~/.config/fish/completions/cert.fish

# PowerShell
cert completion powershell | Out-String | Invoke-Expression
```

## Exit Codes and Output Streams

- Exit code `0` on success and `1` on any failure, including a failed `verify` check
- Human-readable errors are printed to stderr, exactly once, as `Error: <message>`
- With `--json`, errors are printed to stdout as `{"success": false, "error": "<message>"}` so scripts can parse them; nothing is written to stderr
- Regular output (tables, JSON documents) goes to stdout

```bash
# Silence the error text but keep the exit code
cert verify server.crt --expires-in 14d 2>/dev/null || echo "renew soon"

# Capture a JSON error
cert inspect missing.pem --json | jq -r '.error'
```

## Environment Variables

certwiz reads these environment variables:

| Variable | Effect |
|----------|--------|
| `CI`, `CONTINUOUS_INTEGRATION`, `GITHUB_ACTIONS`, `GITLAB_CI`, `JENKINS`, `CIRCLECI` | When any is set, ASCII symbols and plain borders are used instead of emojis and rounded borders |
| `TERM` | `dumb` or empty disables Unicode borders |
| `LANG`, `LC_ALL` | A locale without `utf` disables Unicode borders |
| `XDG_CONFIG_HOME` | Overrides the config directory (default `~/.config`) |

Colors and emojis can also be turned off with `--plain` or the config file;
see the [Configuration](#configuration) section.

## Configuration

`--plain` disables borders, colors, and emojis for a single run. For
persistent defaults, create a config file at the first of these locations
that exists:

1. `$XDG_CONFIG_HOME/certwiz/config.yaml` (default `~/.config/certwiz/config.yaml`)
2. `~/.certwiz.yaml`

```yaml
output:
  plain: false      # Master switch for plain mode
  borders: true     # Show bordered panels
  colors: true      # Use colored output
  emojis: true      # Show emojis (check marks, etc.)
```

Priority: `--plain` flag, then the config file, then CI detection, then the
built-in defaults.

## Output Formats

### Default Output

Human-readable formatted output with colors, bordered tables, status
indicators, and wrapping to the terminal width (measured on stdout; 80
columns when stdout is not a terminal).

### JSON Output

All commands support `--json` for machine-readable output.

```bash
# Inspect with JSON
cert inspect google.com --json | jq '.subject.common_name'

# Verify with JSON
cert verify server.crt --host example.com --json | jq '.is_valid'

# Generate and list output files
cert generate --cn myapp.local --json | jq '.files[]'
```

`generate`, `csr`, `ca`, `sign`, and `convert` return `{"success": true, "message": "...", "files": [...]}`.

### Piping and Redirection

```bash
# Search for specific information
cert inspect google.com --plain | grep "Valid To"

# Save human-readable output
cert inspect example.com --full --plain > cert-details.txt
```

## Advanced Usage

### Batch Operations

```bash
# Check multiple domains
for domain in $(cat domains.txt); do
  cert inspect "$domain" --plain | grep Status
done

# Generate multiple certificates
while IFS= read -r domain; do
  cert generate --cn "$domain" --san "$domain"
done < domains.txt

# Convert all certificates in directory
for cert in *.pem; do
  cert convert "$cert" "${cert%.pem}.der" --format der
done
```

### Integration with Other Tools

```bash
# With OpenSSL
openssl s_client -connect example.com:443 </dev/null 2>/dev/null | cert inspect - --full

# With curl
curl -s https://example.com/cert.pem | cert inspect -

# With find
find /etc/ssl -name "*.crt" -exec cert verify {} \;
```

### Scripting

```bash
#!/bin/bash
# Certificate expiration monitor

check_cert() {
  local domain=$1
  local output

  if ! output=$(cert inspect "$domain" --plain 2>&1); then
    echo "ERROR: Failed to check $domain: $output"
    return 1
  fi

  if echo "$output" | grep -q "EXPIRED"; then
    echo "CRITICAL: $domain certificate has expired"
    return 2
  elif echo "$output" | grep -q "EXPIRING SOON"; then
    echo "WARNING: $domain certificate expiring soon"
    return 1
  else
    echo "OK: $domain certificate is valid"
    return 0
  fi
}

# Check all domains
for domain in example.com api.example.com www.example.com; do
  check_cert "$domain"
done
```

For thresholds other than the built-in 30 days, prefer `--json` and
`days_until_expiry`, or `cert verify --expires-in`.
