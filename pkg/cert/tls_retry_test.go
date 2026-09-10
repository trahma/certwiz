package cert

import (
	"crypto/tls"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

// startRejectingTLSServer starts a TLS server for the given version range
// that closes the first rejectFirst accepted connections at the TCP level
// before any handshake, simulating a per-client connection cap. It returns
// the host, port, and a function reporting how many connections were accepted.
func startRejectingTLSServer(t *testing.T, rejectFirst int, minVersion, maxVersion uint16) (string, int, func() int) {
	t.Helper()
	_, der, key := newTestKeyPair(t)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		MinVersion:   minVersion,
		MaxVersion:   maxVersion,
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	var mu sync.Mutex
	accepted := 0
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			accepted++
			reject := accepted <= rejectFirst
			mu.Unlock()
			if reject {
				_ = conn.Close()
				continue
			}
			go func(c net.Conn) {
				tc := tls.Server(c, cfg)
				_ = tc.Handshake()
				_ = tc.Close()
			}(conn)
		}
	}()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, _ := strconv.Atoi(portStr)
	return host, port, func() int {
		mu.Lock()
		defer mu.Unlock()
		return accepted
	}
}

// TestCheckTLSVersionsRetriesAfterSpuriousReject proves the sequential retry
// pass recovers a version whose parallel probe was reset at the TCP level.
func TestCheckTLSVersionsRetriesAfterSpuriousReject(t *testing.T) {
	// Reject exactly one connection: at most one of TLS 1.2/1.3 can be hit,
	// so the parallel round always has a success and the retry pass runs.
	host, port, accepted := startRejectingTLSServer(t, 1, tls.VersionTLS12, tls.VersionTLS13)

	result, err := CheckTLSVersions(host, port, 5*time.Second)
	if err != nil {
		t.Fatalf("CheckTLSVersions: %v", err)
	}

	supported := map[TLSVersion]bool{}
	for _, v := range result.Versions {
		supported[v.Version] = v.Supported
	}
	if !supported[TLSVersionTLS12] || !supported[TLSVersionTLS13] {
		t.Errorf("TLS 1.2 and 1.3 should be supported after the retry pass, got %+v", result.Versions)
	}
	if supported[TLSVersionTLS10] || supported[TLSVersionTLS11] {
		t.Errorf("TLS 1.0/1.1 should stay unsupported, got %+v", result.Versions)
	}
	if result.MinSupported != TLSVersionTLS12 || result.MaxSupported != TLSVersionTLS13 {
		t.Errorf("min/max = %v/%v, want TLS 1.2/1.3", result.MinSupported, result.MaxSupported)
	}

	// Four parallel probes plus one retry per version that failed in the
	// first round. TLS 1.0 and 1.1 always fail at the handshake, so at least
	// two retries happen (three when the rejected probe was 1.2 or 1.3),
	// giving at least six accepted connections; four would mean no retry.
	if n := accepted(); n < 6 {
		t.Errorf("accepted %d connections, want at least 6 (retry pass should have run)", n)
	}
}

// TestCheckTLSVersionsNoRetryWhenUnreachable proves a host that never
// answers is not retried, so the command still fails in about one timeout.
func TestCheckTLSVersionsNoRetryWhenUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	// Never accept: connections queue in the backlog and handshakes hang.

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	const timeout = 300 * time.Millisecond
	start := time.Now()
	result, err := CheckTLSVersions(host, port, timeout)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("CheckTLSVersions: %v", err)
	}
	for _, v := range result.Versions {
		if v.Supported {
			t.Errorf("%s should be unsupported against a silent listener", v.Name)
		}
	}
	// One parallel round only; allow generous slack for the race detector.
	if elapsed > 3*timeout {
		t.Errorf("took %v, want about one timeout (%v) with no retry pass", elapsed, timeout)
	}
}
