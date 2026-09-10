# Real-World Examples

Practical examples of using certwiz in common scenarios. All examples use the
`cert` binary; see [Installation](installation.md) for how to get it.

## Web Development

### Local Development with HTTPS

Set up HTTPS for local development:

```bash
# Generate certificate for local development
cert generate --cn localhost \
  --san localhost \
  --san "*.localhost" \
  --san myapp.local \
  --san "*.myapp.local" \
  --san IP:127.0.0.1 \
  --san IP:::1 \
  --days 365

# Install certificate (macOS)
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain localhost.crt

# Install certificate (Linux)
sudo cp localhost.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Verify it works and that the key matches
cert verify localhost.crt --host localhost --key localhost.key
```

### Multi-domain Development Certificate

Create a certificate for multiple local domains:

```bash
# Generate wildcard certificate for development
cert generate --cn "dev.local" \
  --san "*.dev.local" \
  --san "*.app.local" \
  --san "*.test.local" \
  --san "*.staging.local" \
  --key-size 4096 \
  --days 730

# Use with nginx
server {
    listen 443 ssl;
    server_name *.dev.local;

    ssl_certificate /path/to/dev.local.crt;
    ssl_certificate_key /path/to/dev.local.key;
}
```

### Private CA for a Team

Issue certificates from your own CA instead of trusting one self-signed
certificate per service:

```bash
# 1. Create the CA (keep ca.key offline)
cert ca --cn "Dev Team Root CA" --org "Example Corp" --days 3650

# 2. Each service creates a CSR and keeps its own key
cert csr --cn api.internal --san api.internal --san "*.api.internal"

# 3. Sign the CSR with the CA
cert sign --csr api.internal.csr --ca Dev_Team_Root_CA-ca.crt \
  --ca-key Dev_Team_Root_CA-ca.key --days 365

# 4. Verify the result against the CA
cert verify api.internal.crt --ca Dev_Team_Root_CA-ca.crt --host api.internal
```

## DevOps and System Administration

### Certificate Expiration Monitoring

Monitor remote certificates across your infrastructure. `--json` gives you the
days remaining without parsing terminal output:

```bash
#!/bin/bash
# cert-monitor.sh - Certificate expiration monitor

DOMAINS=(
    "api.production.com"
    "www.production.com"
    "admin.production.com:8443"
    "database.internal:5432"
)

WARN_DAYS=30
CRITICAL_DAYS=7
status=0

for domain in "${DOMAINS[@]}"; do
    if ! json=$(cert inspect "$domain" --json 2>/dev/null); then
        echo "ERROR: cannot retrieve certificate from $domain"
        status=1
        continue
    fi

    days=$(echo "$json" | jq '.days_until_expiry')
    expired=$(echo "$json" | jq '.is_expired')

    if [[ "$expired" == "true" ]]; then
        echo "CRITICAL: $domain certificate has EXPIRED"
        curl -s -X POST https://alerts.internal/webhook \
            -d "{\"alert\": \"Certificate expired: $domain\"}"
        status=1
    elif (( days <= CRITICAL_DAYS )); then
        echo "CRITICAL: $domain expires in $days days"
        status=1
    elif (( days <= WARN_DAYS )); then
        echo "WARNING: $domain expires in $days days"
    else
        echo "OK: $domain valid for $days days"
    fi
done

exit $status
```

### Expiry Check for Local Certificates (cron)

For certificates on disk, `cert verify --expires-in` does the whole check and
exits 1 when the certificate is inside the window:

```bash
# /etc/cron.daily/check-certs
#!/bin/bash
for f in /etc/ssl/certs/site/*.pem; do
    if ! cert verify "$f" --expires-in 14d --plain >/dev/null 2>/tmp/cert-err; then
        mail -s "Certificate check failed: $f" ops@example.com < /tmp/cert-err
    fi
done
```

Errors are printed to stderr and the exit code is 1, so the mail contains the
reason (for example `Certificate expires in 9 days (within the 14-day threshold)`).

### Kubernetes Ingress Certificate

Generate and deploy certificate for Kubernetes:

```bash
# Generate certificate
cert generate --cn myapp.example.com \
  --san myapp.example.com \
  --san www.myapp.example.com \
  --days 90

# Create Kubernetes secret
kubectl create secret tls myapp-tls \
  --cert=myapp.example.com.crt \
  --key=myapp.example.com.key \
  --namespace=production

# Verify the secret without writing it to disk
kubectl get secret myapp-tls -n production -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | cert inspect -
```

### Inspect a Live Server the Way OpenSSL Sees It

Pipe `openssl s_client` output straight into certwiz. The whole presented chain
is parsed, so `--chain` shows the intermediates too:

```bash
openssl s_client -connect example.com:443 -showcerts </dev/null 2>/dev/null | \
  cert inspect - --chain
```

### Inspect a Bundle

`fullchain.pem` from Let's Encrypt or your CA contains several certificates.
Without `--chain` the first one is shown and a hint tells you how many more
there are:

```bash
cert inspect /etc/letsencrypt/live/example.com/fullchain.pem --chain
```

### Docker Container SSL

Set up SSL for containerized applications:

```bash
# Generate certificate
cert generate --cn docker.local \
  --san docker.local \
  --san "*.docker.local" \
  --san IP:172.17.0.1

# Run the container with the certificate mounted
docker run -d \
  --name myapp \
  -v $(pwd)/docker.local.crt:/etc/ssl/certs/server.crt:ro \
  -v $(pwd)/docker.local.key:/etc/ssl/private/server.key:ro \
  -p 443:443 \
  myapp:latest

# Verify from host
cert inspect docker.local:443
```

## Security and Compliance

### Certificate Compliance Audit

Audit certificates for compliance requirements using the JSON output rather
than scraping the terminal view:

```bash
#!/bin/bash
# compliance-audit.sh - Check certificates meet security requirements

MIN_KEY_SIZE=2048
MAX_VALIDITY_DAYS=397  # CA/Browser Forum limit

check_compliance() {
    local cert_file=$1
    local issues=()

    echo "Auditing: $cert_file"
    json=$(cert inspect "$cert_file" --json) || { echo "  cannot parse"; return; }

    key_size=$(echo "$json" | jq '.public_key_size')
    if (( key_size < MIN_KEY_SIZE )); then
        issues+=("Key size $key_size < required $MIN_KEY_SIZE")
    fi

    days=$(echo "$json" | jq '.days_until_expiry')
    if (( days > MAX_VALIDITY_DAYS )); then
        issues+=("Validity period exceeds $MAX_VALIDITY_DAYS days")
    fi

    if ! echo "$json" | jq -e '.key_usage | index("Digital Signature")' >/dev/null; then
        issues+=("Missing required Digital Signature key usage")
    fi

    if [[ ${#issues[@]} -eq 0 ]]; then
        echo "  COMPLIANT"
    else
        echo "  NON-COMPLIANT:"
        for issue in "${issues[@]}"; do
            echo "    - $issue"
        done
    fi
}

for cert in /etc/ssl/certs/*.crt; do
    check_compliance "$cert"
done
```

### Chain of Trust Verification

Verify complete certificate chains:

```bash
domain="secure.example.com"

# See what the server presents
cert inspect $domain --chain

# Verify each link in a chain you have on disk
cert verify server.crt --ca intermediate.crt --host $domain
cert verify intermediate.crt --ca root.crt

# Or verify the leaf against a bundle containing the whole path
cert verify server.crt --ca fullchain.pem --host $domain
```

### Fingerprint Pinning

Compare a live certificate against a known fingerprint:

```bash
EXPECTED="8F:95:CC:30:E8:8F:6B:71:EF:35:1F:71:03:32:85:0C:55:44:E7:59:4A:F4:0B:1A:7E:11:9E:18:EF:D6:10:22"
actual=$(cert inspect example.com --json | jq -r '.fingerprint_sha256')
[[ "$actual" == "$EXPECTED" ]] || { echo "Fingerprint mismatch!"; exit 1; }
```

### TLS Version Audit

Find servers that still accept deprecated protocol versions:

```bash
for host in api.example.com www.example.com legacy.example.com; do
    min=$(cert tls "$host" --json | jq -r '.min_supported')
    case "$min" in
        "TLS 1.0"|"TLS 1.1") echo "WARN: $host still accepts $min" ;;
        "")                  echo "ERROR: $host did not complete any handshake" ;;
        *)                   echo "OK: $host minimum is $min" ;;
    esac
done
```

The versions are probed concurrently, so a full audit of a host takes about one
handshake's worth of time.

## Troubleshooting

### Debug SSL/TLS Issues

Diagnose connection problems:

```bash
problem_site="https://broken.example.com:8443"

echo "=== Basic Connection Test ==="
cert inspect $problem_site --timeout 10s

echo "=== Certificate Details ==="
cert inspect $problem_site --full

echo "=== Certificate Chain ==="
cert inspect $problem_site --chain

echo "=== Supported TLS Versions ==="
cert tls $problem_site

echo "=== Common Issues to Check ==="
json=$(cert inspect $problem_site --json)

if [[ $(echo "$json" | jq '.is_expired') == "true" ]]; then
    echo "Certificate is expired"
fi

# A self-signed certificate has the same subject and issuer
if [[ $(echo "$json" | jq '.subject == .issuer') == "true" ]]; then
    echo "Certificate appears to be self-signed"
fi

echo "=== SANs ==="
echo "$json" | jq -r '.dns_names[]?, .ip_addresses[]?'
```

### Test Dual-Certificate Configurations

Test servers that support both ECDSA and RSA certificates. With `--sig-alg`
certwiz only offers cipher suites for that key type, so a server without a
matching certificate fails the handshake:

```bash
test_dual_certs() {
    local domain=$1
    echo "Testing dual-certificate configuration for $domain"

    for alg in ecdsa rsa; do
        if key=$(cert inspect "$domain" --sig-alg $alg --json 2>/dev/null | jq -r '.public_key_algorithm'); then
            echo "  [OK] $alg: server presented a $key certificate"
        else
            echo "  [--] $alg: no certificate for this key type"
        fi
    done
}

for domain in cloudflare.com google.com github.com; do
    test_dual_certs $domain
done
```

### Certificate Migration

Compare certificates before and after a migration:

```bash
OLD_SERVER="old.server.com"
NEW_SERVER="new.server.com"

diff <(cert inspect $OLD_SERVER --json | jq 'del(.source)') \
     <(cert inspect $NEW_SERVER --json | jq 'del(.source)')
```

## API and Microservices

### mTLS Setup

Set up mutual TLS for microservices with a private CA:

```bash
# Create the CA
cert ca --cn "MicroServices CA" --days 3650 --key-size 4096

# Server certificate: CSR, then sign
cert csr --cn api.internal --san api.internal --san "*.api.internal"
cert sign --csr api.internal.csr --ca MicroServices_CA-ca.crt \
  --ca-key MicroServices_CA-ca.key --days 365

# Client certificate
cert csr --cn client.internal --san client.internal
cert sign --csr client.internal.csr --ca MicroServices_CA-ca.crt \
  --ca-key MicroServices_CA-ca.key --days 365

# Verify both chain back to the CA and match their keys
cert verify api.internal.crt --ca MicroServices_CA-ca.crt --host api.internal --key api.internal.key
cert verify client.internal.crt --ca MicroServices_CA-ca.crt --key client.internal.key
```

### API Gateway Certificate

Configure API gateway with proper certificates:

```bash
# Generate certificate for API gateway
cert generate --cn api.company.com \
  --san api.company.com \
  --san api-v1.company.com \
  --san api-v2.company.com \
  --san api-staging.company.com \
  --key-size 4096

# Verify certificate covers all endpoints
for endpoint in api api-v1 api-v2 api-staging; do
    cert verify api.company.com.crt --host $endpoint.company.com
done
```

## CI/CD Integration

### GitHub Actions

Install a release binary and fail the job when a certificate is close to
expiry. Errors go to stderr and the exit code is 1, so no output parsing is
needed:

```yaml
name: Certificate Check
on:
  schedule:
    - cron: '0 9 * * *'  # Daily at 9 AM
  workflow_dispatch:

jobs:
  check-certificates:
    runs-on: ubuntu-latest
    steps:
      - name: Install certwiz
        run: |
          curl -sL https://github.com/trahma/certwiz/releases/latest/download/cert-linux-x86_64.tar.gz | tar xz cert
          sudo mv cert /usr/local/bin/cert
          cert version

      - name: Check production certificates
        run: |
          for domain in api.example.com www.example.com admin.example.com; do
            days=$(cert inspect "$domain" --json | jq '.days_until_expiry')
            echo "$domain: $days days remaining"
            if (( days < 30 )); then
              echo "::error::$domain expires in $days days"
              exit 1
            fi
          done

      - name: Check deployed certificate files
        run: |
          cert verify certs/site.pem --expires-in 30d --plain

      - name: Generate report
        run: |
          {
            echo "# Certificate Status Report"
            echo "Date: $(date)"
            echo ""
            for domain in api.example.com www.example.com; do
              echo "## $domain"
              echo '```'
              cert inspect $domain --plain
              echo '```'
            done
          } > certificate-report.md

      - name: Upload report
        uses: actions/upload-artifact@v7
        with:
          name: certificate-report
          path: certificate-report.md
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any

    triggers {
        cron('0 0 * * *')  // Daily
    }

    stages {
        stage('Install certwiz') {
            steps {
                sh '''
                    curl -sL https://github.com/trahma/certwiz/releases/latest/download/cert-linux-x86_64.tar.gz | tar xz cert
                    chmod +x cert
                '''
            }
        }

        stage('Check Certificates') {
            steps {
                script {
                    def domains = ['api.example.com', 'www.example.com']
                    domains.each { domain ->
                        sh """
                            days=\$(./cert inspect ${domain} --json | jq '.days_until_expiry')
                            echo "${domain}: \$days days remaining"
                            if [ "\$days" -lt 30 ]; then
                                echo "WARNING: ${domain} expires in \$days days"
                                exit 1
                            fi
                        """
                    }
                }
            }
        }

        stage('Generate Certificates') {
            when {
                expression { params.GENERATE_NEW }
            }
            steps {
                sh '''
                    ./cert generate --cn jenkins.local \
                        --san jenkins.local \
                        --san "*.jenkins.local" \
                        --days 365
                '''
            }
        }
    }

    post {
        failure {
            emailext (
                subject: "Certificate Check Failed",
                body: "Certificate validation failed. Check Jenkins for details.",
                to: "ops-team@example.com"
            )
        }
    }
}
```

### Docker Image

There is no official image yet. A minimal one that downloads the release
binary:

```dockerfile
FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl \
 && curl -sL https://github.com/trahma/certwiz/releases/latest/download/cert-linux-x86_64.tar.gz \
    | tar xz -C /usr/local/bin cert
ENTRYPOINT ["cert"]
```

```bash
docker build -t cert .
docker run --rm cert inspect example.com --plain
```

## Cloud Providers

### AWS Certificate Verification

Verify certificates in AWS environment:

```bash
# Check ALB certificates
for alb in $(aws elbv2 describe-load-balancers --query 'LoadBalancers[*].DNSName' --output text); do
    echo "Checking ALB: $alb"
    cert inspect $alb:443
done

# Check CloudFront distributions
for dist in $(aws cloudfront list-distributions --query 'DistributionList.Items[*].DomainName' --output text); do
    echo "Checking CloudFront: $dist"
    cert inspect $dist
done
```

### Azure App Service

```bash
# Check Azure App Service certificates
az webapp list --query '[].{name:name, url:defaultHostName}' -o tsv | while read name url; do
    echo "Checking $name at $url"
    cert inspect $url
done
```

### Google Cloud Platform

```bash
# Check GCP Load Balancer certificates; pipe the PEM straight in
gcloud compute ssl-certificates list --format="value(name)" | while read name; do
    echo "Checking certificate: $name"
    gcloud compute ssl-certificates describe $name --format="get(certificate)" | cert inspect -
done
```
