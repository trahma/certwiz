package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"certwiz/internal/testutil"
)

func TestFormatFingerprint(t *testing.T) {
	tests := []struct {
		in   []byte
		want string
	}{
		{nil, ""},
		{[]byte{0x00}, "00"},
		{[]byte{0xab, 0x01, 0xff}, "AB:01:FF"},
	}
	for _, tt := range tests {
		if got := formatFingerprint(tt.in); got != tt.want {
			t.Errorf("formatFingerprint(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// newTestKeyPair returns a self-signed RSA certificate for 127.0.0.1/localhost
// with the given validity, suitable for serving TLS in tests.
func newTestKeyPair(t *testing.T) (*x509.Certificate, []byte, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return parsed, der, key
}

// startTLSServer starts an in-process TLS listener with a self-signed cert
// that accepts the given version range and returns its host and port.
func startTLSServer(t *testing.T, minVersion, maxVersion uint16) (string, int) {
	t.Helper()
	_, der, key := newTestKeyPair(t)
	return startTLSServerWithCert(t, tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, minVersion, maxVersion)
}

// startTLSServerWithCert starts an in-process TLS listener serving the given
// certificate (which may carry a chain) for the given version range.
func startTLSServerWithCert(t *testing.T, tlsCert tls.Certificate, minVersion, maxVersion uint16) (string, int) {
	t.Helper()

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   minVersion,
		MaxVersion:   maxVersion,
	})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				// Drive the handshake so the client sees a completed connection.
				if tc, ok := c.(*tls.Conn); ok {
					_ = tc.Handshake()
				}
				_ = c.Close()
			}(conn)
		}
	}()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func TestCheckTLSVersionsLocalServer(t *testing.T) {
	host, port := startTLSServer(t, tls.VersionTLS12, tls.VersionTLS13)

	result, err := CheckTLSVersions(host, port, 5*time.Second)
	if err != nil {
		t.Fatalf("CheckTLSVersions: %v", err)
	}

	if len(result.Versions) != 4 {
		t.Fatalf("expected 4 version entries, got %d", len(result.Versions))
	}

	// Results must be in ascending version order regardless of probe completion order.
	for i := 1; i < len(result.Versions); i++ {
		if result.Versions[i].Version <= result.Versions[i-1].Version {
			t.Errorf("versions out of order at index %d: %v then %v",
				i, result.Versions[i-1].Version, result.Versions[i].Version)
		}
	}

	supported := map[TLSVersion]bool{}
	for _, v := range result.Versions {
		supported[v.Version] = v.Supported
		if v.Name == "" {
			t.Errorf("version %v has empty name", v.Version)
		}
		if v.Supported && v.CipherSuite == 0 {
			t.Errorf("%s reported supported but no cipher suite recorded", v.Name)
		}
	}

	if !supported[TLSVersionTLS12] {
		t.Error("expected TLS 1.2 to be supported")
	}
	if !supported[TLSVersionTLS13] {
		t.Error("expected TLS 1.3 to be supported")
	}
	if supported[TLSVersionTLS10] || supported[TLSVersionTLS11] {
		t.Error("expected TLS 1.0 and 1.1 to be unsupported by a 1.2+ server")
	}

	if result.MinSupported != TLSVersionTLS12 {
		t.Errorf("MinSupported = %v, want TLS 1.2", result.MinSupported)
	}
	if result.MaxSupported != TLSVersionTLS13 {
		t.Errorf("MaxSupported = %v, want TLS 1.3", result.MaxSupported)
	}
}

func TestConvertBundle(t *testing.T) {
	tmp := t.TempDir()
	bundle := testutil.TestdataPath("fullchain.pem")

	// PEM output must preserve every certificate in the bundle.
	outPEM := filepath.Join(tmp, "bundle.pem")
	format, err := Convert(bundle, outPEM, "pem")
	if err != nil {
		t.Fatalf("Convert to PEM: %v", err)
	}
	if format != FormatPEM {
		t.Errorf("input format = %q, want PEM", format)
	}
	in, err := InspectFileAll(bundle)
	if err != nil {
		t.Fatalf("inspect input: %v", err)
	}
	out, err := InspectFileAll(outPEM)
	if err != nil {
		t.Fatalf("inspect output: %v", err)
	}
	if len(in) < 2 {
		t.Fatalf("test bundle should hold multiple certs, got %d", len(in))
	}
	if len(out) != len(in) {
		t.Errorf("converted bundle has %d certs, want %d", len(out), len(in))
	}

	// DER output cannot hold more than one certificate.
	if _, err := Convert(bundle, filepath.Join(tmp, "bundle.der"), "der"); err == nil {
		t.Error("expected error converting a multi-cert bundle to DER")
	}

	// DER input should be reported as DER.
	format, err = Convert(testutil.TestdataPath("valid.der"), filepath.Join(tmp, "single.pem"), "pem")
	if err != nil {
		t.Fatalf("Convert DER to PEM: %v", err)
	}
	if format != FormatDER {
		t.Errorf("input format = %q, want DER", format)
	}
}

func TestGeneratedKeysArePrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission semantics only")
	}
	tmp := t.TempDir()

	// Pre-create a world-readable key file to confirm it gets tightened.
	csrKey := filepath.Join(tmp, "req.key")
	if err := os.WriteFile(csrKey, []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateCSR(CSROptions{CommonName: "req.local", KeySize: 2048}, filepath.Join(tmp, "req.csr"), csrKey); err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}
	if err := GenerateCA(CAOptions{CommonName: "Test CA", Days: 1, KeySize: 2048}, filepath.Join(tmp, "ca.crt"), filepath.Join(tmp, "ca.key")); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}
	if err := Generate(GenerateOptions{CommonName: "gen.local", Days: 1, KeySize: 2048, OutputDir: tmp}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, name := range []string{"req.key", "ca.key", "gen.local.key"} {
		info, err := os.Stat(filepath.Join(tmp, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("%s has permissions %o, want 0600", name, perm)
		}
	}

	// The CSR key must still be a valid PKCS#8 key (the stale content was replaced).
	data, err := os.ReadFile(csrKey)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		t.Fatalf("CSR key file is not a PEM private key")
	}
	if _, err := parsePrivateKey(data); err != nil {
		t.Errorf("CSR key does not parse: %v", err)
	}
}

func TestSignCSRWithPKCS1AndDERCA(t *testing.T) {
	tmp := t.TempDir()

	// Build a CA whose key is stored as PKCS#1 and whose cert is DER.
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	serial, _ := newSerialNumber()
	caTmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "PKCS1 CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCertPath := filepath.Join(tmp, "ca.der")
	if err := os.WriteFile(caCertPath, caDER, 0644); err != nil {
		t.Fatal(err)
	}
	caKeyPath := filepath.Join(tmp, "ca-pkcs1.key")
	if err := os.WriteFile(caKeyPath, pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	}), 0600); err != nil {
		t.Fatal(err)
	}

	csrPath := filepath.Join(tmp, "server.csr")
	if err := GenerateCSR(CSROptions{CommonName: "server.local", KeySize: 2048, SANs: []string{"server.local"}}, csrPath, filepath.Join(tmp, "server.key")); err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}

	outPath := filepath.Join(tmp, "server.crt")
	if err := SignCSR(SignOptions{CSRPath: csrPath, CACert: caCertPath, CAKey: caKeyPath, Days: 1}, outPath); err != nil {
		t.Fatalf("SignCSR: %v", err)
	}

	signed, err := InspectFile(outPath)
	if err != nil {
		t.Fatalf("inspect signed cert: %v", err)
	}
	if signed.Issuer.CommonName != "PKCS1 CA" {
		t.Errorf("issuer = %q, want PKCS1 CA", signed.Issuer.CommonName)
	}
	if err := signed.CheckSignatureFrom(&x509.Certificate{
		Raw: caDER, PublicKey: &caKey.PublicKey, PublicKeyAlgorithm: x509.RSA,
		BasicConstraintsValid: true, IsCA: true, KeyUsage: x509.KeyUsageCertSign,
	}); err != nil {
		t.Errorf("signature does not verify against CA: %v", err)
	}
}
