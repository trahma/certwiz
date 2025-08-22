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
cert inspect google.com
```

Output shows:
- Subject and Issuer
- Validity dates and expiration status
- Public key type and signature algorithm
- Subject Alternative Names (SANs)

### Inspect a Certificate File

```bash
# PEM format
cert inspect server.crt

# DER format  
cert inspect certificate.der

# Automatic format detection
cert inspect mycert.pem
```

### Custom Ports

```bash
# Specify port explicitly
cert inspect example.com:8443

# Or use the --port flag
cert inspect example.com --port 8443
```

### View Certificate Chain

See the complete trust chain:

```bash
cert inspect github.com --chain
```

This shows:
- Server certificate
- Intermediate certificates
- Path to root CA

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
- Certificate Policies
- And more...

### Combine Options

```bash
# View everything
cert inspect example.com --full --chain

# Check a specific port with full details
cert inspect api.example.com:8443 --full
```

## Generating Certificates

### Basic Self-Signed Certificate

```bash
cert generate --cn myapp.local
```

Creates:
- `myapp.local.crt` - Certificate file
- `myapp.local.key` - Private key file

### With Subject Alternative Names (SANs)

```bash
cert generate --cn myapp.local \
  --san myapp.local \
  --san "*.myapp.local" \
  --san localhost \
  --san IP:127.0.0.1 \
  --san IP:192.168.1.100
```

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

## Converting Certificates

### PEM to DER

```bash
cert convert certificate.pem certificate.der --format der
```

### DER to PEM

```bash
cert convert certificate.der certificate.pem --format pem
```

### Auto-detect Input Format

certwiz automatically detects the input format:

```bash
# Converts to DER if input is PEM
cert convert input.crt output.der --format der

# Converts to PEM if input is DER
cert convert input.der output.pem --format pem
```

## Verifying Certificates

### Basic Verification

```bash
cert verify server.crt
```

Checks:
- Certificate validity dates
- Certificate structure
- Basic constraints

### Verify Against Hostname

```bash
cert verify server.crt --host example.com
```

Verifies:
- Hostname matches CN or SANs
- Certificate is valid for the specified domain

### Verify Against CA

```bash
cert verify server.crt --ca ca-bundle.crt
```

Validates:
- Certificate chain
- Signature verification
- Trust path to CA

### Combined Verification

```bash
cert verify server.crt \
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
  cert inspect $domain | grep -E "Status|Valid"
done
```

### Export Certificate from Website

```bash
# Save certificate to file
cert inspect example.com > example.com.info.txt
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
  cert generate --cn $domain --san $domain --san "*.$domain"
done
```

### Certificate Monitoring

```bash
# Simple expiration check script
#!/bin/bash
domains=("example.com" "api.example.com" "www.example.com")

for domain in "${domains[@]}"; do
  output=$(cert inspect $domain | grep Status)
  if [[ $output == *"EXPIRING SOON"* ]] || [[ $output == *"EXPIRED"* ]]; then
    echo "ALERT: $domain - $output"
  fi
done
```

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

# 3. Check specific port if non-standard
cert inspect problematic-site.com:8443

# 4. Save details for analysis
cert inspect problematic-site.com --full --chain > cert-analysis.txt
```

### Certificate Renewal Process

```bash
# 1. Check current certificate
cert inspect mysite.com

# 2. Generate new certificate
cert generate --cn mysite.com --san mysite.com --san www.mysite.com

# 3. Verify new certificate
cert verify mysite.com.crt --host mysite.com

# 4. Convert if needed
cert convert mysite.com.crt mysite.com.der --format der
```

## Next Steps

- Explore [Command Reference](commands.md) for all options
- See [Examples](examples.md) for real-world scenarios
- Read [FAQ](faq.md) for common questions