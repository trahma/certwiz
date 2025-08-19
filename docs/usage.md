# Usage Guide

This guide covers the basic usage of certwiz commands with practical examples.

## Basic Commands

certwiz has four main commands:
- `inspect` - View certificate information
- `generate` - Create certificates
- `convert` - Convert between formats
- `verify` - Validate certificates

## Inspecting Certificates

### Inspect a Website

The simplest use case - check a website's certificate:

```bash
certwiz inspect google.com
```

Output shows:
- Subject and Issuer
- Validity dates and expiration status
- Public key type and signature algorithm
- Subject Alternative Names (SANs)

### Inspect a Certificate File

```bash
# PEM format
certwiz inspect server.crt

# DER format  
certwiz inspect certificate.der

# Automatic format detection
certwiz inspect mycert.pem
```

### Custom Ports

```bash
# Specify port explicitly
certwiz inspect example.com:8443

# Or use the --port flag
certwiz inspect example.com --port 8443
```

### View Certificate Chain

See the complete trust chain:

```bash
certwiz inspect github.com --chain
```

This shows:
- Server certificate
- Intermediate certificates
- Path to root CA

### Detailed Extension Information

View all certificate extensions with human-readable formatting:

```bash
certwiz inspect google.com --full
```

Shows:
- Key Usage flags
- Extended Key Usage
- Basic Constraints
- Authority Info Access URLs
- Certificate Policies
- And more...

### Combine Options

```bash
# View everything
certwiz inspect example.com --full --chain

# Check a specific port with full details
certwiz inspect api.example.com:8443 --full
```

## Generating Certificates

### Basic Self-Signed Certificate

```bash
certwiz generate --cn myapp.local
```

Creates:
- `myapp.local.crt` - Certificate file
- `myapp.local.key` - Private key file

### With Subject Alternative Names (SANs)

```bash
certwiz generate --cn myapp.local \
  --san myapp.local \
  --san "*.myapp.local" \
  --san localhost \
  --san IP:127.0.0.1 \
  --san IP:192.168.1.100
```

### Custom Validity Period

```bash
# Valid for 2 years
certwiz generate --cn myapp.local --days 730

# Valid for 90 days (Let's Encrypt style)
certwiz generate --cn myapp.local --days 90
```

### Custom Key Size

```bash
# 4096-bit RSA key
certwiz generate --cn myapp.local --key-size 4096

# Default is 2048-bit
certwiz generate --cn myapp.local --key-size 2048
```

### Specify Output Directory

```bash
certwiz generate --cn myapp.local --output /etc/ssl/certs/
```

## Converting Certificates

### PEM to DER

```bash
certwiz convert certificate.pem certificate.der --format der
```

### DER to PEM

```bash
certwiz convert certificate.der certificate.pem --format pem
```

### Auto-detect Input Format

certwiz automatically detects the input format:

```bash
# Converts to DER if input is PEM
certwiz convert input.crt output.der --format der

# Converts to PEM if input is DER
certwiz convert input.der output.pem --format pem
```

## Verifying Certificates

### Basic Verification

```bash
certwiz verify server.crt
```

Checks:
- Certificate validity dates
- Certificate structure
- Basic constraints

### Verify Against Hostname

```bash
certwiz verify server.crt --host example.com
```

Verifies:
- Hostname matches CN or SANs
- Certificate is valid for the specified domain

### Verify Against CA

```bash
certwiz verify server.crt --ca ca-bundle.crt
```

Validates:
- Certificate chain
- Signature verification
- Trust path to CA

### Combined Verification

```bash
certwiz verify server.crt \
  --host api.example.com \
  --ca /etc/ssl/certs/ca-bundle.crt
```

## Understanding the Output

### Color Coding

certwiz uses colors to highlight important information:

- 🟢 **Green**: Valid, healthy, good
- 🟡 **Yellow**: Warning, expiring soon (< 30 days)
- 🔴 **Red**: Error, expired, critical issue
- 🔵 **Blue**: Informational, neutral

### Status Messages

```
Valid (365 days remaining)         # Healthy certificate
EXPIRING SOON (15 days remaining)  # Needs renewal soon
EXPIRED (10 days ago)               # Certificate has expired
```

### Icons and Symbols

- ✓ Enabled/Valid/Success
- ✗ Disabled/Invalid/Failed
- → Indicates a value or detail
- 🔗 Clickable URL or link
- [CRITICAL] Extension that must be understood

## Tips and Tricks

### Quick Domain Check

```bash
# Check multiple domains quickly
for domain in google.com github.com cloudflare.com; do
  echo "=== $domain ==="
  certwiz inspect $domain | grep -E "Status|Valid"
done
```

### Export Certificate from Website

```bash
# Save certificate to file
certwiz inspect example.com > example.com.info.txt
```

### Check Internal Services

```bash
# Check internal service with self-signed cert
certwiz inspect internal.service.local:8443

# Verify against internal CA
certwiz verify internal.crt --ca /path/to/internal-ca.crt
```

### Batch Certificate Generation

```bash
# Generate certificates for multiple domains
for domain in app1.local app2.local app3.local; do
  certwiz generate --cn $domain --san $domain --san "*.$domain"
done
```

### Certificate Monitoring

```bash
# Simple expiration check script
#!/bin/bash
domains=("example.com" "api.example.com" "www.example.com")

for domain in "${domains[@]}"; do
  output=$(certwiz inspect $domain | grep Status)
  if [[ $output == *"EXPIRING SOON"* ]] || [[ $output == *"EXPIRED"* ]]; then
    echo "ALERT: $domain - $output"
  fi
done
```

## Common Workflows

### Setting Up Local Development

```bash
# 1. Generate a certificate for local development
certwiz generate --cn myapp.local \
  --san myapp.local \
  --san "*.myapp.local" \
  --san localhost \
  --san IP:127.0.0.1

# 2. Move to appropriate directory
sudo mv myapp.local.crt /usr/local/etc/ssl/certs/
sudo mv myapp.local.key /usr/local/etc/ssl/private/

# 3. Verify the certificate
certwiz verify /usr/local/etc/ssl/certs/myapp.local.crt --host myapp.local
```

### Debugging SSL Issues

```bash
# 1. Check the problematic certificate
certwiz inspect problematic-site.com --full

# 2. View the certificate chain
certwiz inspect problematic-site.com --chain

# 3. Check specific port if non-standard
certwiz inspect problematic-site.com:8443

# 4. Save details for analysis
certwiz inspect problematic-site.com --full --chain > cert-analysis.txt
```

### Certificate Renewal Process

```bash
# 1. Check current certificate
certwiz inspect mysite.com

# 2. Generate new certificate
certwiz generate --cn mysite.com --san mysite.com --san www.mysite.com

# 3. Verify new certificate
certwiz verify mysite.com.crt --host mysite.com

# 4. Convert if needed
certwiz convert mysite.com.crt mysite.com.der --format der
```

## Next Steps

- Explore [Command Reference](commands.md) for all options
- See [Examples](examples.md) for real-world scenarios
- Read [FAQ](faq.md) for common questions