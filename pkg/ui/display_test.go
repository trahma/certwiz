package ui

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"certwiz/internal/config"
	"certwiz/pkg/cert"
)

// usePlainConfig switches the UI to plain output (no colors, borders, or
// emojis) for the duration of the test so assertions are deterministic.
func usePlainConfig(t *testing.T) {
	t.Helper()
	prev := uiConfig
	cfg := config.DefaultConfig()
	cfg.ApplyPlainMode()
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(prev) })
}

// assertContains fails the test for every expected substring missing from output.
func assertContains(t *testing.T, output string, expected ...string) {
	t.Helper()
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q\n--- output ---\n%s", want, output)
		}
	}
}

// assertNotContains fails the test for every substring that must be absent.
func assertNotContains(t *testing.T, output string, unexpected ...string) {
	t.Helper()
	for _, notWant := range unexpected {
		if strings.Contains(output, notWant) {
			t.Errorf("output should not contain %q\n--- output ---\n%s", notWant, output)
		}
	}
}

// testCustomOID is a private-arc OID used for the unparsed critical extension.
var testCustomOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1}

// testUnknownEKU is an extended key usage OID that Go does not recognise.
var testUnknownEKU = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 2}

// newRichCertificate generates a self-signed certificate exercising every
// extension branch of the --full display and parses it back so that
// Extensions is populated.
func newRichCertificate(t *testing.T) *cert.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	uri, err := url.Parse("https://spiffe.example.com/workload")
	if err != nil {
		t.Fatalf("parse uri: %v", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(4242),
		Subject:      pkix.Name{CommonName: "rich.example.com", Organization: []string{"Rich Org"}},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(365 * 24 * time.Hour),

		KeyUsage:           x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		UnknownExtKeyUsage: []asn1.ObjectIdentifier{testUnknownEKU},

		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,

		DNSNames:       []string{"rich.example.com", "www.rich.example.com"},
		IPAddresses:    []net.IP{net.ParseIP("10.0.0.1")},
		EmailAddresses: []string{"admin@rich.example.com"},
		URIs:           []*url.URL{uri},

		OCSPServer:            []string{"http://ocsp.example.com"},
		IssuingCertificateURL: []string{"http://ca.example.com/issuer.crt"},
		CRLDistributionPoints: []string{"http://crl.example.com/root.crl"},
		PolicyIdentifiers:     []asn1.ObjectIdentifier{{2, 23, 140, 1, 2, 1}},

		ExtraExtensions: []pkix.Extension{{
			Id:       testCustomOID,
			Critical: true,
			Value:    []byte{0x05, 0x00}, // ASN.1 NULL
		}},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	return &cert.Certificate{
		Certificate:     parsed,
		Source:          "rich.pem",
		Format:          cert.FormatPEM,
		DaysUntilExpiry: 365,
	}
}

func TestDisplayCertificateFullExtensions(t *testing.T) {
	usePlainConfig(t)
	c := newRichCertificate(t)

	output := captureOutput(func() {
		DisplayCertificate(c, true)
	})

	expected := []string{
		"Certificate Extensions",
		"Key Usage [CRITICAL]",
		"Extended Key Usage",
		"Server Authentication",
		"Client Authentication",
		testUnknownEKU.String(),
		"Basic Constraints [CRITICAL]",
		"Certificate Authority: Yes",
		"Max Path Length: 0",
		"Subject Alternative Name",
		"5 SANs (2 DNS, 1 IP, 1 Email, 1 URI)",
		"Authority Info Access",
		"OCSP:",
		"http://ocsp.example.com",
		"CA Issuers:",
		"http://ca.example.com/issuer.crt",
		"CRL Distribution Points",
		"http://crl.example.com/root.crl",
		"Certificate Policies",
		"Domain Validated (2.23.140.1.2.1)",
		"Other Extensions",
		testCustomOID.String() + " [CRITICAL]",
		"Subject Key Identifier",
	}
	expected = append(expected, cert.KeyUsageNames(c.KeyUsage)...)
	assertContains(t, output, expected...)

	// The main panel should list every SAN type, not just DNS and IP.
	assertContains(t, output, "admin@rich.example.com", "https://spiffe.example.com/workload", "10.0.0.1")
}

func TestDisplayCertificateFullWithoutExtensions(t *testing.T) {
	usePlainConfig(t)
	now := time.Now()
	c := &cert.Certificate{
		Certificate: &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{CommonName: "bare.example.com"},
			NotBefore:    now.Add(-time.Hour),
			NotAfter:     now.Add(24 * time.Hour),
		},
		Source:          "bare.pem",
		DaysUntilExpiry: 1,
	}

	output := captureOutput(func() {
		DisplayCertificate(c, true)
	})

	assertContains(t, output, "CN=bare.example.com")
	assertNotContains(t, output, "Certificate Extensions", "Other Extensions")
}

func TestDisplayCertificateNonCABasicConstraints(t *testing.T) {
	usePlainConfig(t)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(7),
		Subject:               pkix.Name{CommonName: "leaf.example.com"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	output := captureOutput(func() {
		DisplayCertificate(&cert.Certificate{Certificate: parsed, DaysUntilExpiry: 1}, true)
	})

	assertContains(t, output, "Basic Constraints", "Certificate Authority: No", "Digital Signature")
	assertNotContains(t, output, "Max Path Length")
}

func TestDisplayCertificateSourceTitleAndSANCount(t *testing.T) {
	usePlainConfig(t)
	now := time.Now()

	var manySANs []string
	for i := 0; i < 12; i++ {
		manySANs = append(manySANs, fmt.Sprintf("host%02d.example.com", i))
	}

	newCert := func(source string, dns []string) *cert.Certificate {
		return &cert.Certificate{
			Certificate: &x509.Certificate{
				SerialNumber: big.NewInt(99),
				Subject:      pkix.Name{CommonName: "title.example.com"},
				NotBefore:    now.Add(-time.Hour),
				NotAfter:     now.Add(90 * 24 * time.Hour),
				DNSNames:     dns,
			},
			Source:          source,
			DaysUntilExpiry: 90,
		}
	}

	tests := []struct {
		name       string
		source     string
		dns        []string
		want       []string
		doNotWant  []string
		wantSANRow bool
	}{
		{
			name:      "URL source uses 'for'",
			source:    "https://example.com",
			dns:       []string{"example.com"},
			want:      []string{"Certificate for https://example.com"},
			doNotWant: []string{"Certificate from"},
		},
		{
			name:      "file source uses 'from'",
			source:    "server.pem",
			dns:       []string{"example.com"},
			want:      []string{"Certificate from server.pem"},
			doNotWant: []string{"Certificate for"},
		},
		{
			name:      "file named httpd.pem is still a file",
			source:    "httpd.pem",
			dns:       []string{"example.com"},
			want:      []string{"Certificate from httpd.pem"},
			doNotWant: []string{"Certificate for"},
		},
		{
			name:      "empty source uses generic title",
			source:    "",
			dns:       []string{"example.com"},
			want:      []string{"Certificate Information"},
			doNotWant: []string{"Certificate for", "Certificate from"},
		},
		{
			name:   "more than ten SANs shows a total",
			source: "many.pem",
			dns:    manySANs,
			want:   append([]string{"(12 total)"}, manySANs...),
		},
		{
			name:      "ten or fewer SANs shows no total",
			source:    "few.pem",
			dns:       manySANs[:10],
			want:      manySANs[:10],
			doNotWant: []string{"total)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				DisplayCertificate(newCert(tt.source, tt.dns), false)
			})
			assertContains(t, output, tt.want...)
			assertNotContains(t, output, tt.doNotWant...)
		})
	}
}

func TestDisplayCSRInfo(t *testing.T) {
	usePlainConfig(t)

	info := &cert.CSRInfo{
		Subject: pkix.Name{
			CommonName:   "csr.example.com",
			Organization: []string{"CSR Org"},
			Country:      []string{"US"},
		},
		SignatureAlgorithm: "SHA256-RSA",
		PublicKeyAlgorithm: "RSA",
		KeySize:            2048,
		SANs:               []string{"csr.example.com", "IP:192.168.1.10", "email:ops@example.com"},
	}

	output := captureOutput(func() {
		DisplayCSRInfo(info)
	})

	assertContains(t, output,
		"Subject",
		"CN=csr.example.com, O=CSR Org, C=US",
		"Signature Algorithm",
		"SHA256-RSA",
		"RSA 2048 bits",
		"Subject Alt Names",
		"csr.example.com, IP:192.168.1.10",
		"email:ops@example.com",
	)

	t.Run("without SANs", func(t *testing.T) {
		output := captureOutput(func() {
			DisplayCSRInfo(&cert.CSRInfo{
				Subject:            pkix.Name{CommonName: "plain.example.com"},
				SignatureAlgorithm: "ECDSA-SHA256",
				PublicKeyAlgorithm: "ECDSA",
				KeySize:            256,
			})
		})
		assertContains(t, output, "CN=plain.example.com", "ECDSA 256 bits")
		assertNotContains(t, output, "Subject Alt Names")
	})
}

func TestDisplayTLSVersionResults(t *testing.T) {
	usePlainConfig(t)

	t.Run("legacy versions enabled", func(t *testing.T) {
		result := &cert.TLSResult{
			Host: "legacy.example.com",
			Port: 8443,
			Versions: []cert.TLSVersionInfo{
				{Version: cert.TLSVersionTLS10, Name: "TLS 1.0", Supported: true},
				{Version: cert.TLSVersionTLS11, Name: "TLS 1.1", Supported: true},
				{Version: cert.TLSVersionTLS12, Name: "TLS 1.2", Supported: true, CipherSuite: tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
				{Version: cert.TLSVersionTLS13, Name: "TLS 1.3", Supported: false, Error: "protocol version not supported"},
			},
			MinSupported: cert.TLSVersionTLS10,
			MaxSupported: cert.TLSVersionTLS12,
		}

		output := captureOutput(func() {
			DisplayTLSVersionResults(result)
		})

		assertContains(t, output,
			"TLS Version Support for legacy.example.com:8443",
			"TLS 1.0", "TLS 1.1", "TLS 1.2", "TLS 1.3",
			"[OK] Supported",
			"[X] Not Supported",
			"(TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)",
			"Minimum supported version: TLS 1.0",
			"Maximum supported version: TLS 1.2",
			"[!] Security Warning:",
			"TLS 1.0 is enabled but deprecated",
			"TLS 1.1 is enabled but deprecated",
			"Recommendation: Consider disabling TLS 1.0 and TLS 1.1",
		)

		// Exactly one supported row should carry a cipher suite, and the
		// unsupported row must not.
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "TLS 1.3") && strings.Contains(line, "Not Supported") && strings.Contains(line, "(") {
				t.Errorf("unsupported version should not list a cipher suite: %q", line)
			}
		}
	})

	t.Run("modern only", func(t *testing.T) {
		result := &cert.TLSResult{
			Host: "modern.example.com",
			Port: 443,
			Versions: []cert.TLSVersionInfo{
				{Version: cert.TLSVersionTLS10, Name: "TLS 1.0"},
				{Version: cert.TLSVersionTLS11, Name: "TLS 1.1"},
				{Version: cert.TLSVersionTLS12, Name: "TLS 1.2"},
				{Version: cert.TLSVersionTLS13, Name: "TLS 1.3", Supported: true, CipherSuite: tls.TLS_AES_128_GCM_SHA256},
			},
			MinSupported: cert.TLSVersionTLS13,
			MaxSupported: cert.TLSVersionTLS13,
		}

		output := captureOutput(func() {
			DisplayTLSVersionResults(result)
		})

		assertContains(t, output,
			"Minimum supported version: TLS 1.3",
			"Maximum supported version: TLS 1.3",
			"(TLS_AES_128_GCM_SHA256)",
		)
		assertNotContains(t, output, "Security Warning", "deprecated", "Recommendation:")
	})

	t.Run("nothing supported", func(t *testing.T) {
		result := &cert.TLSResult{
			Host: "down.example.com",
			Port: 443,
			Versions: []cert.TLSVersionInfo{
				{Version: cert.TLSVersionTLS10, Name: "TLS 1.0", Error: "connection refused"},
				{Version: cert.TLSVersionTLS11, Name: "TLS 1.1", Error: "connection refused"},
				{Version: cert.TLSVersionTLS12, Name: "TLS 1.2", Error: "connection refused"},
				{Version: cert.TLSVersionTLS13, Name: "TLS 1.3", Error: "connection refused"},
			},
		}

		output := captureOutput(func() {
			DisplayTLSVersionResults(result)
		})

		assertContains(t, output, "Summary", "Not Supported")
		assertNotContains(t, output,
			"Minimum supported version",
			"Maximum supported version",
			"Security Warning",
			"[OK] Supported",
		)
	})
}

func TestDisplayVerificationResultChecks(t *testing.T) {
	usePlainConfig(t)
	now := time.Now()

	newResult := func(notBefore, notAfter time.Time) *cert.VerificationResult {
		return &cert.VerificationResult{
			Certificate: &cert.Certificate{
				Certificate: &x509.Certificate{
					Subject:   pkix.Name{CommonName: "verify.example.com"},
					NotBefore: notBefore,
					NotAfter:  notAfter,
				},
			},
			IsValid:  true,
			Errors:   []string{},
			Warnings: []string{},
		}
	}

	t.Run("valid with matching key", func(t *testing.T) {
		result := newResult(now.Add(-time.Hour), now.Add(24*time.Hour))
		result.KeyChecked = true
		result.KeyMatches = true

		output := captureOutput(func() { DisplayVerificationResult(result) })

		assertContains(t, output,
			"[OK] Certificate is valid",
			"Validation Checks:",
			"Date validity: PASS",
			"Private key match: PASS",
		)
		assertNotContains(t, output, "Errors:", "Warnings:", "FAIL")
	})

	t.Run("invalid with mismatched key, errors and warnings", func(t *testing.T) {
		result := newResult(now.Add(-time.Hour), now.Add(24*time.Hour))
		result.IsValid = false
		result.KeyChecked = true
		result.KeyMatches = false
		result.Errors = []string{"Private key does not match certificate"}
		result.Warnings = []string{"Certificate expires in 3 days"}

		output := captureOutput(func() { DisplayVerificationResult(result) })

		assertContains(t, output,
			"[FAIL] Certificate validation failed",
			"Errors:",
			"[X] Private key does not match certificate",
			"Warnings:",
			"[!] Certificate expires in 3 days",
			"Date validity: PASS",
			"Private key match: FAIL",
		)
	})

	t.Run("expired", func(t *testing.T) {
		result := newResult(now.Add(-48*time.Hour), now.Add(-24*time.Hour))
		result.IsValid = false
		result.Errors = []string{"Certificate has expired"}

		output := captureOutput(func() { DisplayVerificationResult(result) })

		assertContains(t, output, "Expired: FAIL")
		assertNotContains(t, output, "Private key match")
	})

	t.Run("not yet valid", func(t *testing.T) {
		result := newResult(now.Add(24*time.Hour), now.Add(48*time.Hour))
		result.IsValid = false
		result.Errors = []string{"Certificate is not yet valid"}

		output := captureOutput(func() { DisplayVerificationResult(result) })

		assertContains(t, output, "Not yet valid: FAIL")
	})
}

func TestDisplayCertificateChainStatuses(t *testing.T) {
	usePlainConfig(t)
	now := time.Now()

	chain := []*cert.Certificate{
		{
			Certificate: &x509.Certificate{
				Subject:   pkix.Name{CommonName: "Expired Intermediate"},
				Issuer:    pkix.Name{CommonName: "Root CA"},
				NotBefore: now.Add(-2 * 365 * 24 * time.Hour),
				NotAfter:  now.Add(-24 * time.Hour),
			},
			IsExpired:       true,
			DaysUntilExpiry: -1,
		},
		{
			Certificate: &x509.Certificate{
				Subject:   pkix.Name{CommonName: "Soon Root"},
				Issuer:    pkix.Name{CommonName: "Soon Root"},
				NotBefore: now.Add(-365 * 24 * time.Hour),
				NotAfter:  now.Add(10 * 24 * time.Hour),
			},
			DaysUntilExpiry: 10,
		},
		{
			Certificate: &x509.Certificate{
				Subject:   pkix.Name{CommonName: "Healthy Root"},
				Issuer:    pkix.Name{CommonName: "Healthy Root"},
				NotBefore: now.Add(-365 * 24 * time.Hour),
				NotAfter:  now.Add(365 * 24 * time.Hour),
			},
			DaysUntilExpiry: 365,
		},
	}

	output := captureOutput(func() {
		DisplayCertificateChain(chain)
	})

	assertContains(t, output,
		"Certificate Chain",
		"Chain[1]", "Chain[2]", "Chain[3]",
		"CN=Expired Intermediate",
		"Status", "EXPIRED",
		"Expiring in 10 days",
		"CN=Healthy Root",
		"Valid",
	)

	t.Run("empty chain prints nothing", func(t *testing.T) {
		output := captureOutput(func() {
			DisplayCertificateChain(nil)
		})
		if output != "" {
			t.Errorf("expected no output for empty chain, got %q", output)
		}
	})
}

func TestGetCriticalLabel(t *testing.T) {
	usePlainConfig(t)

	if got := getCriticalLabel(true); got != " [CRITICAL]" {
		t.Errorf("getCriticalLabel(true) = %q, want %q", got, " [CRITICAL]")
	}
	if got := getCriticalLabel(false); got != "" {
		t.Errorf("getCriticalLabel(false) = %q, want empty", got)
	}
}
