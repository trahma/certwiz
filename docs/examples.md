# Real-World Examples

Practical examples of using certwiz in common scenarios.

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

# Verify it works
cert verify localhost.crt --host localhost
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

## DevOps & System Administration

### Certificate Expiration Monitoring

Monitor certificates across your infrastructure:

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

for domain in "${DOMAINS[@]}"; do
    echo "Checking $domain..."
    
    output=$(cert inspect "$domain" 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "❌ ERROR: Cannot connect to $domain"
        continue
    fi
    
    # Extract days remaining
    days=$(echo "$output" | grep -oE '[0-9]+ days remaining' | awk '{print $1}')
    
    if [[ -z "$days" ]]; then
        # Check if expired
        if echo "$output" | grep -q "EXPIRED"; then
            echo "🔴 CRITICAL: $domain certificate has EXPIRED"
            # Send alert
            curl -X POST https://alerts.internal/webhook \
                -d "{'alert': 'Certificate expired: $domain'}"
        fi
    elif [[ $days -le $CRITICAL_DAYS ]]; then
        echo "🔴 CRITICAL: $domain expires in $days days"
        # Send urgent alert
    elif [[ $days -le $WARN_DAYS ]]; then
        echo "🟡 WARNING: $domain expires in $days days"
        # Send warning
    else
        echo "✅ OK: $domain valid for $days days"
    fi
done
```

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

# Verify the secret
kubectl get secret myapp-tls -n production -o yaml | \
  grep tls.crt | awk '{print $2}' | \
  base64 -d > /tmp/k8s-cert.pem

cert inspect /tmp/k8s-cert.pem
```

### Docker Container SSL

Set up SSL for containerized applications:

```bash
# Generate certificate
cert generate --cn docker.local \
  --san docker.local \
  --san "*.docker.local" \
  --san IP:172.17.0.1

# Create Docker network with certificates
docker run -d \
  --name myapp \
  -v $(pwd)/docker.local.crt:/etc/ssl/certs/server.crt:ro \
  -v $(pwd)/docker.local.key:/etc/ssl/private/server.key:ro \
  -p 443:443 \
  myapp:latest

# Verify from host
cert inspect docker.local:443
```

## Security & Compliance

### Certificate Compliance Audit

Audit certificates for compliance requirements:

```bash
#!/bin/bash
# compliance-audit.sh - Check certificates meet security requirements

MIN_KEY_SIZE=2048
MAX_VALIDITY_DAYS=397  # CAB Forum requirement
REQUIRED_KEY_USAGE="Digital Signature, Key Encipherment"

check_compliance() {
    local cert_file=$1
    local issues=()
    
    echo "Auditing: $cert_file"
    
    # Get full certificate details
    details=$(cert inspect "$cert_file" --full)
    
    # Check key size
    key_info=$(echo "$details" | grep "Public Key" | grep -oE '[0-9]+ bits')
    key_size=${key_info% bits}
    
    if [[ $key_size -lt $MIN_KEY_SIZE ]]; then
        issues+=("Key size $key_size < required $MIN_KEY_SIZE")
    fi
    
    # Check validity period
    days=$(echo "$details" | grep -oE '[0-9]+ days remaining' | awk '{print $1}')
    if [[ $days -gt $MAX_VALIDITY_DAYS ]]; then
        issues+=("Validity period exceeds $MAX_VALIDITY_DAYS days")
    fi
    
    # Check for required key usage
    if ! echo "$details" | grep -q "Digital Signature"; then
        issues+=("Missing required Digital Signature key usage")
    fi
    
    # Report findings
    if [[ ${#issues[@]} -eq 0 ]]; then
        echo "✅ COMPLIANT"
    else
        echo "❌ NON-COMPLIANT:"
        for issue in "${issues[@]}"; do
            echo "  - $issue"
        done
    fi
    echo
}

# Audit all certificates
for cert in /etc/ssl/certs/*.crt; do
    check_compliance "$cert"
done
```

### Chain of Trust Verification

Verify complete certificate chains:

```bash
# Download and verify certificate chain
domain="secure.example.com"

# Get the certificate and chain
cert inspect $domain --chain > chain-analysis.txt

# Extract each certificate
cert inspect $domain --full | grep -A 100 "Certificate Chain" > chain.txt

# Verify each link in the chain
echo "Verifying chain for $domain"
cert verify server.crt --ca intermediate.crt
cert verify intermediate.crt --ca root.crt
```

## Troubleshooting

### Debug SSL/TLS Issues

Diagnose connection problems:

```bash
# Check problematic service
problem_site="https://broken.example.com:8443"

echo "=== Basic Connection Test ==="
cert inspect $problem_site 2>&1

echo -e "\n=== Certificate Details ==="
cert inspect $problem_site --full 2>&1

echo -e "\n=== Certificate Chain ==="
cert inspect $problem_site --chain 2>&1

echo -e "\n=== Common Issues to Check ==="
output=$(cert inspect $problem_site 2>&1)

# Check for expired certificate
if echo "$output" | grep -q "EXPIRED"; then
    echo "❌ Certificate is expired"
fi

# Check for self-signed
if echo "$output" | grep -q "self-signed"; then
    echo "⚠️ Certificate appears to be self-signed"
fi

# Check SANs match
echo -e "\n=== SAN Verification ==="
cert inspect $problem_site | grep -A10 "SANs"
```

### Test Dual-Certificate Configurations

Test servers that support both ECDSA and RSA certificates:

```bash
# Check which certificate types are available
echo "=== Testing ECDSA Certificate ==="
cert inspect cloudflare.com --sig-alg ecdsa | grep "Public Key"

echo -e "\n=== Testing RSA Certificate ==="
cert inspect cloudflare.com --sig-alg rsa | grep "Public Key"

# Script to verify dual-cert setup
test_dual_certs() {
    local domain=$1
    
    echo "Testing dual-certificate configuration for $domain"
    
    # Try ECDSA
    ecdsa_result=$(cert inspect $domain --sig-alg ecdsa 2>&1)
    if echo "$ecdsa_result" | grep -q "ECDSA"; then
        echo "✓ ECDSA certificate available"
        echo "$ecdsa_result" | grep "Public Key"
    else
        echo "✗ No ECDSA certificate"
    fi
    
    # Try RSA  
    rsa_result=$(cert inspect $domain --sig-alg rsa 2>&1)
    if echo "$rsa_result" | grep -q "RSA"; then
        echo "✓ RSA certificate available"
        echo "$rsa_result" | grep "Public Key"
    else
        echo "✗ No RSA certificate"
    fi
}

# Test multiple domains
for domain in cloudflare.com google.com github.com; do
    test_dual_certs $domain
    echo "---"
done
```

### Certificate Migration

Migrate certificates between systems:

```bash
# Export certificates from old system
OLD_SERVER="old.server.com"
NEW_SERVER="new.server.com"

# Inspect current certificate
echo "Current certificate on $OLD_SERVER:"
cert inspect $OLD_SERVER

# After migration, verify new setup
echo "New certificate on $NEW_SERVER:"
cert inspect $NEW_SERVER

# Compare certificates
diff <(cert inspect $OLD_SERVER) <(cert inspect $NEW_SERVER)
```

## API & Microservices

### mTLS Setup

Set up mutual TLS for microservices:

```bash
# Generate CA certificate
cert generate --cn "MicroServices CA" \
  --days 3650 \
  --key-size 4096

# Generate server certificate
cert generate --cn api.internal \
  --san api.internal \
  --san "*.api.internal" \
  --days 365

# Generate client certificate
cert generate --cn client.internal \
  --san client.internal \
  --days 365

# Verify the certificates
cert verify api.internal.crt --host api.internal
cert verify client.internal.crt --host client.internal
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

Integrate certwiz in CI/CD pipeline:

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
          go install github.com/certwiz/certwiz@latest
          
      - name: Check production certificates
        run: |
          domains="api.example.com www.example.com admin.example.com"
          for domain in $domains; do
            echo "Checking $domain..."
            cert inspect $domain || exit 1
            
            # Fail if expiring soon
            output=$(cert inspect $domain)
            if echo "$output" | grep -E "EXPIRING SOON|EXPIRED"; then
              echo "::error::Certificate issue for $domain"
              exit 1
            fi
          done
          
      - name: Generate report
        run: |
          {
            echo "# Certificate Status Report"
            echo "Date: $(date)"
            echo ""
            for domain in api.example.com www.example.com; do
              echo "## $domain"
              cert inspect $domain
              echo ""
            done
          } > certificate-report.md
          
      - name: Upload report
        uses: actions/upload-artifact@v2
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
                sh 'go install github.com/certwiz/certwiz@latest'
            }
        }
        
        stage('Check Certificates') {
            steps {
                script {
                    def domains = ['api.example.com', 'www.example.com']
                    domains.each { domain ->
                        sh """
                            echo "Checking ${domain}..."
                            cert inspect ${domain}
                            
                            # Extract days remaining
                            days=\$(cert inspect ${domain} | grep -oE '[0-9]+ days remaining' | awk '{print \$1}')
                            
                            if [ "\$days" -lt "30" ]; then
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
                    cert generate --cn jenkins.local \
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
# Check GCP Load Balancer certificates
gcloud compute ssl-certificates list --format="value(name)" | while read cert; do
    echo "Checking certificate: $cert"
    # Export and check
    gcloud compute ssl-certificates describe $cert --format="get(certificate)" > /tmp/$cert.pem
    cert inspect /tmp/$cert.pem
done
```