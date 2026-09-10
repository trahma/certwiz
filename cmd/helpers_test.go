package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"certwiz/internal/config"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// captureStdout runs fn and returns everything it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	return <-done
}

// withStdin points os.Stdin at data for the duration of the test.
func withStdin(t *testing.T, data []byte) {
	t.Helper()
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	go func() {
		_, _ = w.Write(data)
		_ = w.Close()
	}()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
}

// setOutputMode configures the global output flags and the UI package for a
// test, restoring the previous values on cleanup. PersistentPreRun is not run
// when RunE is invoked directly, so tests must do this themselves.
func setOutputMode(t *testing.T, json, plain bool) {
	t.Helper()
	oldJSON, oldPlain := jsonOutput, plainOutput
	jsonOutput, plainOutput = json, plain

	cfg := config.DefaultConfig()
	if plain {
		cfg.ApplyPlainMode()
	}
	ui.SetConfig(cfg)

	t.Cleanup(func() {
		jsonOutput, plainOutput = oldJSON, oldPlain
		ui.SetConfig(config.DefaultConfig())
	})
}

// resetInspectFlags restores the inspect command flags on cleanup and sets
// them to their defaults for the test.
func resetInspectFlags(t *testing.T) {
	t.Helper()
	full, port, chain, connect, timeout, sigAlg := inspectFull, inspectPort, inspectChain, inspectConnect, inspectTimeout, inspectSigAlg
	inspectFull, inspectPort, inspectChain, inspectConnect, inspectTimeout, inspectSigAlg = false, 443, false, "", 5*time.Second, "auto"
	t.Cleanup(func() {
		inspectFull, inspectPort, inspectChain, inspectConnect, inspectTimeout, inspectSigAlg = full, port, chain, connect, timeout, sigAlg
	})
}

// resetTLSFlags restores the tls command flags on cleanup.
func resetTLSFlags(t *testing.T) {
	t.Helper()
	port, timeout := tlsPort, tlsTimeout
	tlsPort, tlsTimeout = 443, 5*time.Second
	t.Cleanup(func() { tlsPort, tlsTimeout = port, timeout })
}

// testCert is a generated certificate with its key and DER encoding.
type testCert struct {
	cert *x509.Certificate
	der  []byte
	key  *rsa.PrivateKey
}

// newTestCA generates a self-signed CA certificate.
func newTestCA(t *testing.T) testCert {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA", Organization: []string{"certwiz tests"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
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
	return testCert{cert: parsed, der: der, key: key}
}

// newTestLeaf generates a server certificate for localhost/127.0.0.1. When
// issuer is nil the certificate is self-signed.
func newTestLeaf(t *testing.T, issuer *testCert) testCert {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	parent := &tmpl
	signer := key
	if issuer != nil {
		parent = issuer.cert
		signer = issuer.key
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, parent, &key.PublicKey, signer)
	if err != nil {
		t.Fatalf("create leaf cert: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse leaf cert: %v", err)
	}
	return testCert{cert: parsed, der: der, key: key}
}

// startTLSServer starts an in-process TLS listener presenting the given
// certificate chain (leaf first) and returns its host and port. The server
// completes the handshake and closes each connection.
func startTLSServer(t *testing.T, chain [][]byte, key *rsa.PrivateKey, minVersion, maxVersion uint16) (string, int) {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{{Certificate: chain, PrivateKey: key}},
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
				if tc, ok := c.(*tls.Conn); ok {
					_ = tc.Handshake()
				}
				_ = c.Close()
			}(conn)
		}
	}()

	return splitAddr(t, ln.Addr().String())
}

// startSilentListener accepts TCP connections but never speaks, so any TLS
// handshake against it must time out.
func startSilentListener(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	var conns []net.Conn
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				for _, c := range conns {
					_ = c.Close()
				}
				return
			}
			conns = append(conns, conn)
		}
	}()

	return splitAddr(t, ln.Addr().String())
}

func splitAddr(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, port
}

// resetCommandFlags restores every flag on cmd and its subcommands to its
// default and clears the Changed marker. Cobra keeps parsed flag values
// (including --help) between Execute calls, which otherwise leaks state from
// one table-driven case or -count iteration into the next.
func resetCommandFlags(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				if err := f.Value.Set(f.DefValue); err != nil {
					t.Fatalf("reset flag %s: %v", f.Name, err)
				}
				f.Changed = false
			}
		})
	}
	reset(cmd.Flags())
	reset(cmd.PersistentFlags())
	for _, sub := range cmd.Commands() {
		resetCommandFlags(t, sub)
	}
}
