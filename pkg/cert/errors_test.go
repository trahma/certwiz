package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testCA is an in-test certificate authority used to mint leaf certificates
// for the error-path tests below.
type testCA struct {
	cert *x509.Certificate
	der  []byte
	key  *rsa.PrivateKey
}

func newTestCA(t *testing.T) *testCA {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	serial, err := newSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "Error Path Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(240 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}
	return &testCA{cert: parsed, der: der, key: key}
}

// signLeaf issues a server certificate for dnsName with the given validity.
func (ca *testCA) signLeaf(t *testing.T, dnsName string, notBefore, notAfter time.Time) ([]byte, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	serial, err := newSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: dnsName},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{dnsName},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("sign leaf: %v", err)
	}
	return der, key
}

func writeTestFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func certPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func keyPEM(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}

func assertHasError(t *testing.T, result *VerificationResult, want string) {
	t.Helper()
	if result.IsValid {
		t.Fatalf("expected invalid result, got valid (errors: %v)", result.Errors)
	}
	for _, e := range result.Errors {
		if strings.Contains(e, want) {
			return
		}
	}
	t.Fatalf("errors %v do not contain %q", result.Errors, want)
}

// missingDirPath returns a path inside a directory that does not exist, so
// any attempt to create the file fails regardless of platform or privileges.
func missingDirPath(dir, name string) string {
	return filepath.Join(dir, "does-not-exist", name)
}

func TestVerifyWithOptionsErrorPaths(t *testing.T) {
	tmp := t.TempDir()
	ca := newTestCA(t)
	const host = "leaf.example.test"

	leafDER, leafKey := ca.signLeaf(t, host, time.Now().Add(-time.Hour), time.Now().Add(48*time.Hour))
	leafPath := writeTestFile(t, tmp, "leaf.pem", certPEM(leafDER))
	leafKeyPath := writeTestFile(t, tmp, "leaf.key", keyPEM(t, leafKey))
	caPEMPath := writeTestFile(t, tmp, "ca.pem", certPEM(ca.der))
	caDERPath := writeTestFile(t, tmp, "ca.der", ca.der)

	t.Run("MissingKeyFile", func(t *testing.T) {
		_, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, KeyPath: missingDirPath(tmp, "nope.key")})
		assertErrContains(t, err, "failed to read key file")
	})

	t.Run("UnparsableKey", func(t *testing.T) {
		bad := writeTestFile(t, tmp, "garbage.key", []byte("this is not a key"))
		_, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, KeyPath: bad})
		assertErrContains(t, err, "failed to parse private key")
	})

	t.Run("MatchingKey", func(t *testing.T) {
		result, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, KeyPath: leafKeyPath})
		if err != nil {
			t.Fatal(err)
		}
		if !result.KeyChecked || !result.KeyMatches || !result.IsValid {
			t.Fatalf("expected key match, got checked=%v matches=%v valid=%v", result.KeyChecked, result.KeyMatches, result.IsValid)
		}
	})

	t.Run("MissingCAFile", func(t *testing.T) {
		_, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, CAPath: missingDirPath(tmp, "nope.pem")})
		assertErrContains(t, err, "failed to read CA file")
	})

	t.Run("GarbageCA", func(t *testing.T) {
		bad := writeTestFile(t, tmp, "garbage-ca.pem", []byte("neither PEM nor DER"))
		_, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, CAPath: bad})
		assertErrContains(t, err, "failed to parse CA certificate(s)")
	})

	t.Run("DERCASucceeds", func(t *testing.T) {
		result, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, CAPath: caDERPath})
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsValid {
			t.Fatalf("expected valid chain with DER CA, errors: %v", result.Errors)
		}
	})

	t.Run("HostnameWithCA", func(t *testing.T) {
		result, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, CAPath: caPEMPath, Hostname: host})
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsValid {
			t.Fatalf("expected valid, errors: %v", result.Errors)
		}

		result, err = VerifyWithOptions(VerifyOptions{CertPath: leafPath, CAPath: caPEMPath, Hostname: "other.example.test"})
		if err != nil {
			t.Fatal(err)
		}
		assertHasError(t, result, "Hostname verification failed")
		assertHasError(t, result, "Chain verification failed")
	})

	t.Run("UntrustedCA", func(t *testing.T) {
		other := newTestCA(t)
		otherPath := writeTestFile(t, tmp, "other-ca.pem", certPEM(other.der))
		result, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, CAPath: otherPath})
		if err != nil {
			t.Fatal(err)
		}
		assertHasError(t, result, "Chain verification failed")
	})

	t.Run("NotYetValid", func(t *testing.T) {
		der, _ := ca.signLeaf(t, host, time.Now().Add(time.Hour), time.Now().Add(48*time.Hour))
		path := writeTestFile(t, tmp, "future.pem", certPEM(der))
		result, err := VerifyWithOptions(VerifyOptions{CertPath: path})
		if err != nil {
			t.Fatal(err)
		}
		assertHasError(t, result, "Certificate is not yet valid")
	})

	t.Run("ExpiresInThreshold", func(t *testing.T) {
		result, err := VerifyWithOptions(VerifyOptions{CertPath: leafPath, ExpiresIn: 72 * time.Hour})
		if err != nil {
			t.Fatal(err)
		}
		assertHasError(t, result, "within the 3-day threshold")

		result, err = VerifyWithOptions(VerifyOptions{CertPath: leafPath, ExpiresIn: 24 * time.Hour})
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsValid {
			t.Fatalf("expected valid inside a 1-day window, errors: %v", result.Errors)
		}
	})

	t.Run("MissingCertificate", func(t *testing.T) {
		_, err := VerifyWithOptions(VerifyOptions{CertPath: missingDirPath(tmp, "nope.pem")})
		assertErrContains(t, err, "failed to read file")
	})
}

func TestSignCSRErrorPaths(t *testing.T) {
	tmp := t.TempDir()
	ca := newTestCA(t)
	caCertPath := writeTestFile(t, tmp, "ca.pem", certPEM(ca.der))
	caKeyPath := writeTestFile(t, tmp, "ca.key", keyPEM(t, ca.key))

	csrPath := filepath.Join(tmp, "server.csr")
	if err := GenerateCSR(CSROptions{CommonName: "server.example.test", KeySize: 2048, SANs: []string{"server.example.test"}}, csrPath, filepath.Join(tmp, "server.key")); err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}
	csrData, err := os.ReadFile(csrPath)
	if err != nil {
		t.Fatal(err)
	}

	good := func() SignOptions {
		return SignOptions{CSRPath: csrPath, CACert: caCertPath, CAKey: caKeyPath, Days: 1}
	}
	out := func(name string) string { return filepath.Join(tmp, name) }

	t.Run("MissingCSR", func(t *testing.T) {
		opts := good()
		opts.CSRPath = missingDirPath(tmp, "nope.csr")
		assertErrContains(t, SignCSR(opts, out("a.crt")), "failed to read CSR")
	})

	t.Run("CSRNotPEM", func(t *testing.T) {
		opts := good()
		opts.CSRPath = writeTestFile(t, tmp, "notpem.csr", []byte("plain text"))
		assertErrContains(t, SignCSR(opts, out("b.crt")), "failed to parse CSR PEM block")
	})

	t.Run("CSRBadDER", func(t *testing.T) {
		opts := good()
		opts.CSRPath = writeTestFile(t, tmp, "badder.csr", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: []byte{0x30, 0x01, 0xff}}))
		assertErrContains(t, SignCSR(opts, out("c.crt")), "failed to parse CSR")
	})

	t.Run("CSRBadSignature", func(t *testing.T) {
		block, _ := pem.Decode(csrData)
		if block == nil {
			t.Fatal("CSR fixture is not PEM")
		}
		tampered := append([]byte(nil), block.Bytes...)
		tampered[len(tampered)-1] ^= 0xff // signature is the trailing element
		opts := good()
		opts.CSRPath = writeTestFile(t, tmp, "tampered.csr", pem.EncodeToMemory(&pem.Block{Type: block.Type, Bytes: tampered}))
		assertErrContains(t, SignCSR(opts, out("d.crt")), "CSR signature verification failed")
	})

	t.Run("MissingCACert", func(t *testing.T) {
		opts := good()
		opts.CACert = missingDirPath(tmp, "nope.pem")
		assertErrContains(t, SignCSR(opts, out("e.crt")), "failed to read CA certificate")
	})

	t.Run("GarbageCACert", func(t *testing.T) {
		opts := good()
		opts.CACert = writeTestFile(t, tmp, "garbage-ca.pem", []byte("not a certificate"))
		assertErrContains(t, SignCSR(opts, out("f.crt")), "failed to parse CA certificate")
	})

	t.Run("MissingCAKey", func(t *testing.T) {
		opts := good()
		opts.CAKey = missingDirPath(tmp, "nope.key")
		assertErrContains(t, SignCSR(opts, out("g.crt")), "failed to read CA private key")
	})

	t.Run("GarbageCAKey", func(t *testing.T) {
		opts := good()
		opts.CAKey = writeTestFile(t, tmp, "garbage-ca.key", []byte("not a key"))
		assertErrContains(t, SignCSR(opts, out("h.crt")), "failed to parse CA private key")
	})

	t.Run("MismatchedCAKey", func(t *testing.T) {
		other := newTestCA(t)
		opts := good()
		opts.CAKey = writeTestFile(t, tmp, "other-ca.key", keyPEM(t, other.key))
		assertErrContains(t, SignCSR(opts, out("i.crt")), "failed to create certificate")
	})

	t.Run("UnwritableOutput", func(t *testing.T) {
		assertErrContains(t, SignCSR(good(), missingDirPath(tmp, "j.crt")), "failed to write certificate")
	})

	t.Run("SANOverride", func(t *testing.T) {
		opts := good()
		opts.SANs = []string{"override.example.test", "IP:10.0.0.1", "email:ops@example.test", "uri:https://example.test/id"}
		path := out("override.crt")
		if err := SignCSR(opts, path); err != nil {
			t.Fatalf("SignCSR: %v", err)
		}
		signed, err := InspectFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(signed.DNSNames) != 1 || signed.DNSNames[0] != "override.example.test" {
			t.Errorf("DNSNames = %v, want only the override", signed.DNSNames)
		}
		if len(signed.IPAddresses) != 1 || !signed.IPAddresses[0].Equal(net.ParseIP("10.0.0.1")) {
			t.Errorf("IPAddresses = %v, want [10.0.0.1]", signed.IPAddresses)
		}
		if len(signed.EmailAddresses) != 1 || signed.EmailAddresses[0] != "ops@example.test" {
			t.Errorf("EmailAddresses = %v, want [ops@example.test]", signed.EmailAddresses)
		}
		if len(signed.URIs) != 1 || signed.URIs[0].String() != "https://example.test/id" {
			t.Errorf("URIs = %v, want [https://example.test/id]", signed.URIs)
		}
	})
}

func TestGenerateErrorPaths(t *testing.T) {
	tmp := t.TempDir()

	// A directory placed where a file is expected makes the open fail on
	// every platform without needing permission tricks.
	mkdir := func(t *testing.T, path string) string {
		t.Helper()
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("InvalidKeySize", func(t *testing.T) {
		assertErrContains(t, Generate(GenerateOptions{CommonName: "x", Days: 1, KeySize: 1, OutputDir: tmp}), "failed to generate private key")
		assertErrContains(t, GenerateCSR(CSROptions{CommonName: "x", KeySize: 1}, filepath.Join(tmp, "x.csr"), filepath.Join(tmp, "x.key")), "failed to generate private key")
		assertErrContains(t, GenerateCA(CAOptions{CommonName: "x", Days: 1, KeySize: 1}, filepath.Join(tmp, "x.crt"), filepath.Join(tmp, "x.key")), "failed to generate private key")
	})

	t.Run("GenerateOutputDirIsFile", func(t *testing.T) {
		file := writeTestFile(t, tmp, "not-a-dir", []byte("x"))
		assertErrContains(t, Generate(GenerateOptions{CommonName: "x", Days: 1, KeySize: 2048, OutputDir: file}), "failed to create output directory")
	})

	t.Run("GenerateCertPathUnwritable", func(t *testing.T) {
		dir := mkdir(t, filepath.Join(tmp, "gen-cert"))
		mkdir(t, filepath.Join(dir, "blocked.crt"))
		assertErrContains(t, Generate(GenerateOptions{CommonName: "blocked", Days: 1, KeySize: 2048, OutputDir: dir}), "failed to write certificate")
	})

	t.Run("GenerateKeyPathUnwritable", func(t *testing.T) {
		dir := mkdir(t, filepath.Join(tmp, "gen-key"))
		mkdir(t, filepath.Join(dir, "blocked.key"))
		assertErrContains(t, Generate(GenerateOptions{CommonName: "blocked", Days: 1, KeySize: 2048, OutputDir: dir}), "failed to write private key")
		if _, err := os.Stat(filepath.Join(dir, "blocked.crt")); err != nil {
			t.Errorf("certificate should have been written before the key failed: %v", err)
		}
	})

	t.Run("GenerateCACertPathUnwritable", func(t *testing.T) {
		err := GenerateCA(CAOptions{CommonName: "x", Days: 1, KeySize: 2048}, missingDirPath(tmp, "ca.crt"), filepath.Join(tmp, "ca-unused.key"))
		assertErrContains(t, err, "failed to write certificate")
	})

	t.Run("GenerateCAKeyPathUnwritable", func(t *testing.T) {
		keyDir := mkdir(t, filepath.Join(tmp, "ca-blocked.key"))
		err := GenerateCA(CAOptions{CommonName: "x", Days: 1, KeySize: 2048}, filepath.Join(tmp, "ca-ok.crt"), keyDir)
		assertErrContains(t, err, "failed to write private key")
	})

	t.Run("GenerateCSRPathUnwritable", func(t *testing.T) {
		err := GenerateCSR(CSROptions{CommonName: "x", KeySize: 2048}, missingDirPath(tmp, "x.csr"), filepath.Join(tmp, "csr-unused.key"))
		assertErrContains(t, err, "failed to write CSR")
	})

	t.Run("GenerateCSRKeyPathUnwritable", func(t *testing.T) {
		keyDir := mkdir(t, filepath.Join(tmp, "csr-blocked.key"))
		err := GenerateCSR(CSROptions{CommonName: "x", KeySize: 2048}, filepath.Join(tmp, "csr-ok.csr"), keyDir)
		assertErrContains(t, err, "failed to write private key")
	})
}
