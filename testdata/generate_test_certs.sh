#!/bin/bash
# Generate test certificates for unit tests

set -e

echo "Generating test certificates..."

# Generate a self-signed certificate (valid)
openssl req -x509 -newkey rsa:2048 -keyout valid.key -out valid.pem -days 365 -nodes \
  -subj "/C=US/ST=Test/L=Test/O=Test/CN=test.example.com" \
  -addext "subjectAltName=DNS:test.example.com,DNS:*.test.example.com,IP:127.0.0.1" 2>/dev/null

# Convert to DER format
openssl x509 -in valid.pem -outform DER -out valid.der

# Generate an expired certificate
openssl req -x509 -newkey rsa:2048 -keyout expired.key -out expired.pem -days 1 -nodes \
  -subj "/C=US/ST=Test/L=Test/O=Test/CN=expired.example.com" 2>/dev/null

# Manually set the date to make it expired
# This is a bit hacky but works for testing
faketime '2020-01-01' openssl req -x509 -newkey rsa:2048 -keyout expired.key -out expired.pem -days 1 -nodes \
  -subj "/C=US/ST=Test/L=Test/O=Test/CN=expired.example.com" 2>/dev/null || \
  echo "Note: faketime not available, expired cert will actually be valid for 1 day"

# Generate a certificate with many SANs
openssl req -x509 -newkey rsa:2048 -keyout many-sans.key -out many-sans.pem -days 365 -nodes \
  -subj "/C=US/ST=Test/L=Test/O=Test/CN=many-sans.example.com" \
  -addext "subjectAltName=DNS:san1.example.com,DNS:san2.example.com,DNS:san3.example.com,DNS:san4.example.com,DNS:san5.example.com,DNS:*.wildcard.example.com,IP:192.168.1.1,IP:10.0.0.1" 2>/dev/null

# Generate a certificate with 4096-bit key
openssl req -x509 -newkey rsa:4096 -keyout strong.key -out strong.pem -days 365 -nodes \
  -subj "/C=US/ST=Test/L=Test/O=Test/CN=strong.example.com" 2>/dev/null

# Generate an invalid certificate (corrupted)
echo "-----BEGIN CERTIFICATE-----" > invalid.pem
echo "This is not a valid certificate content" >> invalid.pem
echo "-----END CERTIFICATE-----" >> invalid.pem

# Generate a certificate chain (CA -> Intermediate -> Server)
# Create CA key and cert
openssl genrsa -out ca.key 2048 2>/dev/null
openssl req -new -x509 -days 3650 -key ca.key -out ca.pem \
  -subj "/C=US/ST=Test/L=Test/O=Test CA/CN=Test Root CA" 2>/dev/null

# Create intermediate key and CSR
openssl genrsa -out intermediate.key 2048 2>/dev/null
openssl req -new -key intermediate.key -out intermediate.csr \
  -subj "/C=US/ST=Test/L=Test/O=Test CA/CN=Test Intermediate CA" 2>/dev/null

# Sign intermediate cert with CA
openssl x509 -req -in intermediate.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
  -out intermediate.pem -days 1825 -sha256 2>/dev/null

# Create server key and CSR
openssl genrsa -out chain-server.key 2048 2>/dev/null
openssl req -new -key chain-server.key -out chain-server.csr \
  -subj "/C=US/ST=Test/L=Test/O=Test/CN=chain.example.com" 2>/dev/null

# Sign server cert with intermediate
openssl x509 -req -in chain-server.csr -CA intermediate.pem -CAkey intermediate.key -CAcreateserial \
  -out chain-server.pem -days 365 -sha256 2>/dev/null

# Create full chain file
cat chain-server.pem intermediate.pem ca.pem > fullchain.pem

# Clean up CSR files
rm -f *.csr *.srl

echo "Test certificates generated successfully!"
echo ""
echo "Generated files:"
ls -la *.pem *.der *.key 2>/dev/null | awk '{print "  " $9}'