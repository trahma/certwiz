package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

const testDialTimeout = 5 * time.Second

func assertRemoteCert(t *testing.T, c *Certificate) {
	t.Helper()
	if c == nil {
		t.Fatal("expected a certificate, got nil")
	}
	if c.Format != FormatDER {
		t.Errorf("Format = %q, want %q", c.Format, FormatDER)
	}
	if c.TLSVersion == 0 {
		t.Error("TLSVersion should be recorded for remote inspection")
	}
	if c.CipherSuite == 0 {
		t.Error("CipherSuite should be recorded for remote inspection")
	}
	if c.Subject.CommonName != "localhost" {
		t.Errorf("CommonName = %q, want localhost", c.Subject.CommonName)
	}
	if c.IsExpired {
		t.Error("test certificate should not be expired")
	}
}

func TestInspectURLWithOptionsTargets(t *testing.T) {
	host, port := startTLSServer(t, tls.VersionTLS12, tls.VersionTLS13)

	tests := []struct {
		name       string
		target     string
		port       int
		wantSource string
	}{
		{
			name:       "bare host with port argument",
			target:     host,
			port:       port,
			wantSource: "https://" + host,
		},
		{
			name:       "https scheme with port argument",
			target:     "https://" + host,
			port:       port,
			wantSource: "https://" + host,
		},
		{
			name:       "scheme URL with its own port wins over the port argument",
			target:     fmt.Sprintf("https://%s:%d", host, port),
			port:       1, // deliberately wrong; must be ignored
			wantSource: fmt.Sprintf("https://%s:%d", host, port),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, chain, err := InspectURLWithOptions(tt.target, tt.port, "", testDialTimeout, "auto")
			if err != nil {
				t.Fatalf("InspectURLWithOptions(%q, %d): %v", tt.target, tt.port, err)
			}
			assertRemoteCert(t, c)
			if c.Source != tt.wantSource {
				t.Errorf("Source = %q, want %q", c.Source, tt.wantSource)
			}
			if len(chain) != 0 {
				t.Errorf("expected empty chain for a single-cert server, got %d", len(chain))
			}
			if c.TLSVersion != tls.VersionTLS13 {
				t.Errorf("auto negotiation should reach TLS 1.3, got %s", TLSVersionName(c.TLSVersion))
			}
		})
	}
}

func TestInspectURLWithOptionsConnectHost(t *testing.T) {
	host, port := startTLSServer(t, tls.VersionTLS12, tls.VersionTLS13)

	// Dial the local listener while presenting a different target hostname.
	target := "example.test"
	c, _, err := InspectURLWithOptions(target, port, host, testDialTimeout, "auto")
	if err != nil {
		t.Fatalf("InspectURLWithOptions via connect host: %v", err)
	}
	assertRemoteCert(t, c)
	if c.Source != "https://example.test" {
		t.Errorf("Source = %q, want the target URL, not the connect host", c.Source)
	}

	// A URL port must not override the explicit port when a connect host is given.
	c, _, err = InspectURLWithOptions("https://example.test:1", port, host, testDialTimeout, "auto")
	if err != nil {
		t.Fatalf("InspectURLWithOptions connect host with URL port: %v", err)
	}
	assertRemoteCert(t, c)
}

func TestInspectURLWithOptionsSigAlg(t *testing.T) {
	host, port := startTLSServer(t, tls.VersionTLS12, tls.VersionTLS13)

	t.Run("rsa preference pins TLS 1.2 with an RSA suite", func(t *testing.T) {
		c, _, err := InspectURLWithOptions(host, port, "", testDialTimeout, "rsa")
		if err != nil {
			t.Fatalf("rsa: %v", err)
		}
		assertRemoteCert(t, c)
		if c.TLSVersion != tls.VersionTLS12 {
			t.Errorf("TLSVersion = %s, want TLS 1.2", TLSVersionName(c.TLSVersion))
		}
		if name := tls.CipherSuiteName(c.CipherSuite); !strings.Contains(name, "RSA") {
			t.Errorf("cipher suite %q is not an RSA suite", name)
		}
	})

	t.Run("ecdsa preference fails against an RSA-only server", func(t *testing.T) {
		if _, _, err := InspectURLWithOptions(host, port, "", testDialTimeout, "ecdsa"); err == nil {
			t.Error("expected handshake failure when only ECDSA suites are offered to an RSA server")
		}
	})

	t.Run("sigAlg is case-insensitive", func(t *testing.T) {
		c, _, err := InspectURLWithOptions(host, port, "", testDialTimeout, "RSA")
		if err != nil {
			t.Fatalf("RSA: %v", err)
		}
		if c.TLSVersion != tls.VersionTLS12 {
			t.Errorf("TLSVersion = %s, want TLS 1.2", TLSVersionName(c.TLSVersion))
		}
	})
}

func TestInspectURLWithOptionsChain(t *testing.T) {
	// Build a CA and a leaf signed by it; serve leaf + CA.
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := x509.Certificate{
		SerialNumber:          big.NewInt(10),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	leafTmpl := x509.Certificate{
		SerialNumber: big.NewInt(11),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, &leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}

	host, port := startTLSServerWithCert(t, tls.Certificate{
		Certificate: [][]byte{leafDER, caDER},
		PrivateKey:  leafKey,
	}, tls.VersionTLS12, tls.VersionTLS13)

	c, chain, err := InspectURLWithOptions(host, port, "", testDialTimeout, "auto")
	if err != nil {
		t.Fatalf("InspectURLWithOptions: %v", err)
	}
	assertRemoteCert(t, c)
	if c.Issuer.CommonName != "Test Root CA" {
		t.Errorf("leaf issuer = %q, want Test Root CA", c.Issuer.CommonName)
	}

	if len(chain) != 1 {
		t.Fatalf("chain length = %d, want 1", len(chain))
	}
	if chain[0].Source != "Chain[1]" {
		t.Errorf("chain Source = %q, want Chain[1]", chain[0].Source)
	}
	if chain[0].Format != FormatDER {
		t.Errorf("chain Format = %q, want DER", chain[0].Format)
	}
	if chain[0].Subject.CommonName != "Test Root CA" {
		t.Errorf("chain CommonName = %q, want Test Root CA", chain[0].Subject.CommonName)
	}
	if !chain[0].IsCA {
		t.Error("chain certificate should be a CA")
	}
	if chain[0].TLSVersion != 0 || chain[0].CipherSuite != 0 {
		t.Error("chain certificates should not carry TLS connection info")
	}
}

func TestInspectURLWithOptionsTimeout(t *testing.T) {
	// A listener that accepts connections but never speaks TLS.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// Hold the connection open without responding.
			go func(c net.Conn) {
				buf := make([]byte, 1)
				_, _ = c.Read(buf)
				_ = c.Close()
			}(conn)
		}
	}()

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)

	start := time.Now()
	_, _, err = InspectURLWithOptions(host, port, "", 300*time.Millisecond, "auto")
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected a timeout error from a silent server")
	}
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("error = %q, want a 'failed to connect' error", err)
	}
	if elapsed > 5*time.Second {
		t.Errorf("timeout took %s, expected it to honour the 300ms deadline", elapsed)
	}
}

func TestInspectURLWithOptionsErrors(t *testing.T) {
	if _, _, err := InspectURLWithOptions("https://[::1", 443, "", testDialTimeout, "auto"); err == nil {
		t.Error("expected an invalid URL error")
	} else if !strings.Contains(err.Error(), "invalid URL") {
		t.Errorf("error = %q, want an invalid URL error", err)
	}

	// Nothing listening on this port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	_ = ln.Close()
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)

	if _, _, err := InspectURLWithOptions(host, port, "", testDialTimeout, "auto"); err == nil {
		t.Error("expected a connection error for a closed port")
	} else if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("error = %q, want a 'failed to connect' error", err)
	}
}

func TestInspectURLWrappers(t *testing.T) {
	host, port := startTLSServer(t, tls.VersionTLS12, tls.VersionTLS13)

	t.Run("InspectURL", func(t *testing.T) {
		c, err := InspectURL(host, port)
		if err != nil {
			t.Fatalf("InspectURL: %v", err)
		}
		assertRemoteCert(t, c)
	})

	t.Run("InspectURLWithChain", func(t *testing.T) {
		c, chain, err := InspectURLWithChain(host, port)
		if err != nil {
			t.Fatalf("InspectURLWithChain: %v", err)
		}
		assertRemoteCert(t, c)
		if len(chain) != 0 {
			t.Errorf("chain length = %d, want 0", len(chain))
		}
	})

	t.Run("InspectURLWithConnect", func(t *testing.T) {
		c, _, err := InspectURLWithConnect("example.test", port, host)
		if err != nil {
			t.Fatalf("InspectURLWithConnect: %v", err)
		}
		assertRemoteCert(t, c)
		if c.Source != "https://example.test" {
			t.Errorf("Source = %q, want https://example.test", c.Source)
		}
	})

	t.Run("InspectURLWithConnectTimeout", func(t *testing.T) {
		c, _, err := InspectURLWithConnectTimeout(host, port, "", testDialTimeout)
		if err != nil {
			t.Fatalf("InspectURLWithConnectTimeout: %v", err)
		}
		assertRemoteCert(t, c)
	})
}
