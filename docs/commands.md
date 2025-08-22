# Command Reference

Complete reference for all certwiz commands and options.

## Global Options

These options work with all commands:

```
-h, --help      Show help for any command
    --version   Show version information
```

## inspect

Inspect a certificate from a file or URL.

### Synopsis

```bash
cert inspect [file|url] [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--full` | | Show full certificate details including all extensions | `false` |
| `--chain` | | Show certificate chain (for URLs only) | `false` |
| `--port` | `-p` | Port for remote inspection | `443` |
| `--connect` | | Connect to a different host while validating cert for target | |

### Arguments

- `target` - Certificate file path or URL/domain name (required)

### Examples

```bash
# Inspect local certificate file
cert inspect server.crt
cert inspect /path/to/certificate.pem

# Inspect remote certificate
cert inspect google.com
cert inspect https://example.com
cert inspect api.example.com:8443

# With options
cert inspect google.com --full
cert inspect github.com --chain
cert inspect example.com --full --chain
cert inspect internal.service --port 8443

# Through proxy or tunnel
cert inspect api.example.com --connect localhost:8080
cert inspect prod.internal --connect tunnel.local --port 443
cert inspect backend.local --connect 127.0.0.1:3000
```

### Connect Flag Usage

The `--connect` flag is useful for:
- Testing certificates through SSH tunnels
- Inspecting certificates behind proxies
- Validating certificates in local development environments
- Checking certificates on different servers with the same hostname

When using `--connect`:
- The connection is made to the host specified in `--connect`
- The certificate is validated for the original target hostname (SNI)
- Port can be specified in the connect host (e.g., `localhost:8080`) or via `--port`
- If port is in both, the one in `--connect` takes precedence

### Output Details

The inspect command shows:
- **Subject**: Certificate subject DN
- **Issuer**: Certificate issuer DN
- **Serial Number**: Unique certificate identifier
- **Valid From/To**: Certificate validity period
- **Status**: Current validity status with days remaining
- **Public Key**: Key type and size
- **Signature Algorithm**: Algorithm used to sign the certificate
- **SANs**: All Subject Alternative Names

With `--full`:
- **Key Usage**: Permitted key usage flags
- **Extended Key Usage**: Extended usage purposes
- **Basic Constraints**: CA status and path length
- **Authority Info Access**: OCSP and CA issuer URLs
- **CRL Distribution Points**: Certificate revocation list URLs
- **Certificate Policies**: Policy OIDs
- **Other Extensions**: Any additional extensions

With `--chain`:
- Complete certificate chain from server to root
- Each certificate in the chain with basic info
- Validity status for each certificate

## generate

Generate a self-signed certificate.

### Synopsis

```bash
cert generate [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--cn` | | Common Name for the certificate (required) | |
| `--san` | | Subject Alternative Name (can be repeated) | |
| `--days` | `-d` | Validity period in days | `365` |
| `--key-size` | `-k` | RSA key size in bits | `2048` |
| `--output` | `-o` | Output directory | `.` (current) |

### SAN Format

SANs can be specified as:
- DNS names: `--san example.com`
- Wildcards: `--san "*.example.com"`
- IP addresses: `--san IP:192.168.1.1`

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
- `{cn}.key` - Private key file in PEM format

## convert

Convert certificate between PEM and DER formats.

### Synopsis

```bash
cert convert <input> <output> [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format (pem or der) | `pem` |

### Arguments

- `input` - Input certificate file (required)
- `output` - Output certificate file (required)

### Examples

```bash
# PEM to DER
cert convert certificate.pem certificate.der --format der

# DER to PEM
cert convert certificate.der certificate.pem --format pem

# Auto-detect input format
cert convert input.crt output.der --format der
```

### Format Detection

certwiz automatically detects the input format:
- Files starting with `-----BEGIN` are treated as PEM
- Binary files are treated as DER
- Extensions (.pem, .der, .crt) are used as hints

## verify

Verify a certificate's validity and optionally check against a hostname.

### Synopsis

```bash
cert verify <certificate> [flags]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--host` | | Hostname to verify against | |
| `--ca` | | CA certificate for chain verification | |

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

# Complete verification
cert verify server.crt \
  --host api.example.com \
  --ca /etc/ssl/certs/ca-bundle.crt
```

### Verification Checks

The verify command performs:

**Basic checks:**
- Certificate structure validity
- Expiration date check
- Basic constraints validation

**With --host:**
- Common Name matches hostname
- SANs contain hostname
- Wildcard matching (*.example.com)

**With --ca:**
- Certificate chain validation
- Signature verification
- Trust path to CA

### Exit Codes

All commands use standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - Certificate validation failed

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

The update command checks for the latest version of cert and installs it if an update is available. It will:

1. Check the latest version from GitHub releases
2. Compare with your current version
3. Download and install if newer version is available
4. Automatically detect your installation directory
5. Create a backup of the current binary before upgrading

### Examples

```bash
# Check for and install updates
cert update

# Force reinstall current version (useful for fixing corrupted installations)
cert update --force
```

### Notes

- The update command is not available on Windows. Windows users should download the latest version from the releases page.
- The installer creates a backup of your current binary as `cert.backup` in the same directory
- If the update fails, you can restore the backup manually

## version

Show the version of cert.

### Synopsis

```bash
cert version
```

### Description

Displays the current version of cert installed on your system.

### Examples

```bash
cert version
# Output: cert version 0.1.4
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

## Environment Variables

certwiz respects these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `NO_COLOR` | Disable colored output | |
| `FORCE_COLOR` | Force colored output even in pipes | |
| `CERTWIZ_PORT` | Default port for inspect command | `443` |

### Examples

```bash
# Disable colors
NO_COLOR=1 cert inspect google.com

# Force colors in pipe
FORCE_COLOR=1 cert inspect google.com | less -R

# Set default port
export CERTWIZ_PORT=8443
cert inspect internal.service
```

## Output Formats

### Default Output

Human-readable formatted output with:
- Colors and styling
- Bordered tables
- Status indicators
- Smart text wrapping

### Piping and Redirection

When output is piped, certwiz:
- Maintains structure but may reduce colors
- Preserves all information
- Suitable for grep/awk/sed processing

```bash
# Search for specific information
cert inspect google.com | grep "Valid To"

# Save to file
cert inspect example.com --full > cert-details.txt

# Process with jq (future JSON support)
# cert inspect example.com --json | jq '.subject'
```

## Advanced Usage

### Batch Operations

```bash
# Check multiple domains
for domain in $(cat domains.txt); do
  cert inspect "$domain" | grep Status
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
cert inspect example.com --full | grep -A2 "Key Usage"

# With curl
curl -k https://example.com/cert.pem | cert inspect -

# With find
find /etc/ssl -name "*.crt" -exec cert verify {} \;
```

### Scripting

```bash
#!/bin/bash
# Certificate expiration monitor

check_cert() {
  local domain=$1
  local output=$(cert inspect "$domain" 2>&1)
  
  if [[ $? -ne 0 ]]; then
    echo "ERROR: Failed to check $domain"
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