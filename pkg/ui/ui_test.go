package ui

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"certwiz/pkg/cert"
)

// captureOutput captures stdout during test execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestFormatSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  pkix.Name
		expected string
	}{
		{
			name: "Full subject",
			subject: pkix.Name{
				CommonName:         "test.example.com",
				Organization:       []string{"Test Org"},
				OrganizationalUnit: []string{"Test Unit"},
				Country:            []string{"US"},
			},
			expected: "CN=test.example.com, O=Test Org, OU=Test Unit, C=US",
		},
		{
			name: "Only CommonName",
			subject: pkix.Name{
				CommonName: "simple.example.com",
			},
			expected: "CN=simple.example.com",
		},
		{
			name:     "Empty subject",
			subject:  pkix.Name{},
			expected: "Unknown",
		},
		{
			name: "Multiple organizations",
			subject: pkix.Name{
				CommonName:   "multi.example.com",
				Organization: []string{"Org1", "Org2"},
			},
			expected: "CN=multi.example.com, O=Org1, Org2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSubject(tt.subject)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatPublicKey(t *testing.T) {
	tests := []struct {
		name     string
		key      interface{}
		expected string
	}{
		{
			name: "RSA 2048",
			key: func() interface{} {
				key, _ := rsa.GenerateKey(rand.Reader, 2048)
				return &key.PublicKey
			}(),
			expected: "RSA 2048 bits",
		},
		{
			name: "RSA 4096",
			key: func() interface{} {
				key, _ := rsa.GenerateKey(rand.Reader, 4096)
				return &key.PublicKey
			}(),
			expected: "RSA 4096 bits",
		},
		{
			name: "ECDSA P256",
			key: func() interface{} {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				return &key.PublicKey
			}(),
			expected: "ECDSA P-256",
		},
		{
			name: "ECDSA P384",
			key: func() interface{} {
				key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
				return &key.PublicKey
			}(),
			expected: "ECDSA P-384",
		},
		{
			name:     "Unknown key type",
			key:      "not a key",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPublicKey(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatSANs(t *testing.T) {
	tests := []struct {
		name     string
		sans     []string
		maxWidth int
	}{
		{
			name:     "Single SAN",
			sans:     []string{"test.example.com"},
			maxWidth: 80,
		},
		{
			name: "Multiple SANs",
			sans: []string{
				"test1.example.com",
				"test2.example.com",
				"*.wildcard.example.com",
			},
			maxWidth: 80,
		},
		{
			name: "Many SANs requiring wrapping",
			sans: []string{
				"very-long-subdomain-name-1.example.com",
				"very-long-subdomain-name-2.example.com",
				"very-long-subdomain-name-3.example.com",
				"very-long-subdomain-name-4.example.com",
				"very-long-subdomain-name-5.example.com",
			},
			maxWidth: 60,
		},
		{
			name:     "Empty SANs",
			sans:     []string{},
			maxWidth: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSANs(tt.sans)

			// Check that result contains all SANs
			for _, san := range tt.sans {
				if !strings.Contains(result, san) {
					t.Errorf("Result does not contain SAN %q", san)
				}
			}

			// Check that lines don't exceed expected width (approximation)
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				// Remove leading spaces from continuation lines
				trimmed := strings.TrimLeft(line, " ")
				// The actual content should fit within reasonable bounds
				if len(trimmed) > 100 {
					t.Errorf("Line too long: %d characters", len(trimmed))
				}
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name     string
		cert     *cert.Certificate
		contains string
	}{
		{
			name: "Valid certificate",
			cert: &cert.Certificate{
				Certificate:     &x509.Certificate{},
				IsExpired:       false,
				DaysUntilExpiry: 100,
			},
			contains: "Valid",
		},
		{
			name: "Expired certificate",
			cert: &cert.Certificate{
				Certificate:     &x509.Certificate{},
				IsExpired:       true,
				DaysUntilExpiry: -10,
			},
			contains: "EXPIRED",
		},
		{
			name: "Expiring soon",
			cert: &cert.Certificate{
				Certificate:     &x509.Certificate{},
				IsExpired:       false,
				DaysUntilExpiry: 15,
			},
			contains: "EXPIRING SOON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.cert)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected status to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestShowError(t *testing.T) {
	output := captureOutput(func() {
		ShowError("test error message")
	})

	if !strings.Contains(output, "Error") {
		t.Error("Output should contain 'Error'")
	}
	if !strings.Contains(output, "test error message") {
		t.Error("Output should contain the error message")
	}
}

func TestShowSuccess(t *testing.T) {
	output := captureOutput(func() {
		ShowSuccess("test success message")
	})

	if !strings.Contains(output, "test success message") {
		t.Error("Output should contain the success message")
	}
}

func TestShowInfo(t *testing.T) {
	output := captureOutput(func() {
		ShowInfo("test info message")
	})

	if !strings.Contains(output, "test info message") {
		t.Error("Output should contain the info message")
	}
}

func TestDisplayGenerationResult(t *testing.T) {
	output := captureOutput(func() {
		DisplayGenerationResult("/path/to/cert.crt", "/path/to/cert.key")
	})

	if !strings.Contains(output, "Certificate generated successfully") {
		t.Error("Output should contain success message")
	}
	if !strings.Contains(output, "/path/to/cert.crt") {
		t.Error("Output should contain certificate path")
	}
	if !strings.Contains(output, "/path/to/cert.key") {
		t.Error("Output should contain key path")
	}
}

func TestDisplayConversionResult(t *testing.T) {
	output := captureOutput(func() {
		DisplayConversionResult("input.pem", "output.der", "pem", "der")
	})

	if !strings.Contains(output, "Converted from PEM to DER") {
		t.Error("Output should contain conversion message")
	}
	if !strings.Contains(output, "input.pem") {
		t.Error("Output should contain input path")
	}
	if !strings.Contains(output, "output.der") {
		t.Error("Output should contain output path")
	}
}

func TestDisplayVerificationResult(t *testing.T) {
	now := time.Now()
	x509Cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore: now.Add(-24 * time.Hour),
		NotAfter:  now.Add(30 * 24 * time.Hour),
	}

	tests := []struct {
		name   string
		result *cert.VerificationResult
		checks []string
	}{
		{
			name: "Valid certificate",
			result: &cert.VerificationResult{
				Certificate: &cert.Certificate{
					Certificate: x509Cert,
				},
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
			checks: []string{"Certificate is valid"},
		},
		{
			name: "Invalid certificate with errors",
			result: &cert.VerificationResult{
				Certificate: &cert.Certificate{
					Certificate: x509Cert,
				},
				IsValid: false,
				Errors: []string{
					"Certificate has expired",
					"Hostname verification failed",
				},
				Warnings: []string{},
			},
			checks: []string{
				"Certificate validation failed",
				"Certificate has expired",
				"Hostname verification failed",
			},
		},
		{
			name: "Certificate with warnings",
			result: &cert.VerificationResult{
				Certificate: &cert.Certificate{
					Certificate: x509Cert,
				},
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{"Certificate expires in 10 days"},
			},
			checks: []string{
				"Certificate is valid",
				"Certificate expires in 10 days",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				DisplayVerificationResult(tt.result)
			})

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Output should contain %q, got: %s", check, output)
				}
			}
		})
	}
}

func TestDisplayCertificate(t *testing.T) {
	now := time.Now()
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	x509Cert := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		Issuer: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test CA Org"},
		},
		NotBefore:             now.Add(-24 * time.Hour),
		NotAfter:              now.Add(30 * 24 * time.Hour),
		DNSNames:              []string{"test.example.com", "*.test.example.com"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  false,
		BasicConstraintsValid: true,
		PublicKey:             &rsaKey.PublicKey,
		SignatureAlgorithm:    x509.SHA256WithRSA,
	}

	testCert := &cert.Certificate{
		Certificate:     x509Cert,
		Source:          "test.pem",
		Format:          "PEM",
		IsExpired:       false,
		DaysUntilExpiry: 30,
	}

	tests := []struct {
		name     string
		cert     *cert.Certificate
		showFull bool
		checks   []string
	}{
		{
			name:     "Basic display",
			cert:     testCert,
			showFull: false,
			checks: []string{
				"Certificate from test.pem",
				"CN=test.example.com",
				"CN=Test CA",
				"test.example.com",
				"*.test.example.com",
				"127.0.0.1",
				"RSA 2048 bits",
				"SHA256-RSA",
			},
		},
		{
			name:     "Full display",
			cert:     testCert,
			showFull: true,
			checks: []string{
				"Certificate from test.pem",
				"CN=test.example.com",
				"test.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				DisplayCertificate(tt.cert, tt.showFull)
			})

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Output should contain %q, got: %s", check, output)
				}
			}
		})
	}
}

func TestDisplayCertificateChain(t *testing.T) {
	now := time.Now()

	// Create a mock certificate chain
	chain := []*cert.Certificate{
		{
			Certificate: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "Intermediate CA",
				},
				Issuer: pkix.Name{
					CommonName: "Root CA",
				},
				NotBefore: now.Add(-365 * 24 * time.Hour),
				NotAfter:  now.Add(365 * 24 * time.Hour),
			},
			IsExpired:       false,
			DaysUntilExpiry: 365,
		},
		{
			Certificate: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "Root CA",
				},
				Issuer: pkix.Name{
					CommonName: "Root CA",
				},
				NotBefore: now.Add(-3650 * 24 * time.Hour),
				NotAfter:  now.Add(3650 * 24 * time.Hour),
			},
			IsExpired:       false,
			DaysUntilExpiry: 3650,
		},
	}

	output := captureOutput(func() {
		DisplayCertificateChain(chain)
	})

	checks := []string{
		"Certificate Chain",
		"Chain[1]",
		"Chain[2]",
		"Intermediate CA",
		"Root CA",
		"Valid",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output should contain %q", check)
		}
	}
}

func TestFormatTable(t *testing.T) {
	data := [][]string{
		{"Key1", "Value1"},
		{"LongerKey", "Value2"},
		{"K", "Value3"},
	}

	result := formatTable(data)

	// Check alignment
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Check that all keys are aligned
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			t.Error("Each line should contain a colon separator")
		}
	}
}

func TestGetPolicyName(t *testing.T) {
	tests := []struct {
		oid      string
		expected string
	}{
		{
			oid:      "2.5.29.32.0",
			expected: "Any Policy",
		},
		{
			oid:      "2.23.140.1.2.1",
			expected: "Domain Validated",
		},
		{
			oid:      "2.23.140.1.2.2",
			expected: "Organization Validated",
		},
		{
			oid:      "1.2.3.4.5",
			expected: "1.2.3.4.5", // Unknown OID returns as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.oid, func(t *testing.T) {
			result := getPolicyName(tt.oid)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %q to contain %q", result, tt.expected)
			}
		})
	}
}

func TestIsExtensionCritical(t *testing.T) {
	oid1 := "2.5.29.15"
	oid2 := "2.5.29.17"

	x509Cert := &x509.Certificate{
		Extensions: []pkix.Extension{
			{
				Id:       parseOID(oid1),
				Critical: true,
				Value:    []byte{},
			},
			{
				Id:       parseOID(oid2),
				Critical: false,
				Value:    []byte{},
			},
		},
	}

	if !isExtensionCritical(x509Cert, oid1) {
		t.Errorf("Extension %s should be critical", oid1)
	}

	if isExtensionCritical(x509Cert, oid2) {
		t.Errorf("Extension %s should not be critical", oid2)
	}

	if isExtensionCritical(x509Cert, "1.2.3.4") {
		t.Error("Non-existent extension should not be critical")
	}
}

// Helper function to parse OID string
func parseOID(oid string) []int {
	parts := strings.Split(oid, ".")
	result := make([]int, len(parts))
	for i, part := range parts {
		fmt.Sscanf(part, "%d", &result[i])
	}
	return result
}

// Benchmark tests
func BenchmarkFormatSANs(b *testing.B) {
	sans := []string{
		"test1.example.com",
		"test2.example.com",
		"test3.example.com",
		"*.wildcard.example.com",
		"very-long-subdomain-name.example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatSANs(sans)
	}
}

func BenchmarkFormatSubject(b *testing.B) {
	subject := pkix.Name{
		CommonName:         "test.example.com",
		Organization:       []string{"Test Org"},
		OrganizationalUnit: []string{"Test Unit"},
		Country:            []string{"US"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatSubject(subject)
	}
}
