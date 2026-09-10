package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTLSCommandRun(t *testing.T) {
	leaf := newTestLeaf(t, nil)
	host, port := startTLSServer(t, [][]byte{leaf.der}, leaf.key, tls.VersionTLS12, tls.VersionTLS13)

	t.Run("Plain", func(t *testing.T) {
		resetTLSFlags(t)
		setOutputMode(t, false, true)

		var err error
		out := captureStdout(t, func() {
			err = tlsCmd.RunE(tlsCmd, []string{fmt.Sprintf("%s:%d", host, port)})
		})
		if err != nil {
			t.Fatalf("tls command failed: %v", err)
		}
		for _, want := range []string{
			fmt.Sprintf("TLS Version Support for %s:%d", host, port),
			"TLS 1.0", "TLS 1.1", "TLS 1.2", "TLS 1.3",
			"Minimum supported version: TLS 1.2",
			"Maximum supported version: TLS 1.3",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "Security Warning") {
			t.Error("no deprecated-version warning expected for a TLS 1.2+ server")
		}
	})

	t.Run("JSONWithPortFlag", func(t *testing.T) {
		resetTLSFlags(t)
		setOutputMode(t, true, false)
		tlsPort = port

		var err error
		out := captureStdout(t, func() {
			err = tlsCmd.RunE(tlsCmd, []string{host})
		})
		if err != nil {
			t.Fatalf("tls command failed: %v", err)
		}
		var result struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Versions []struct {
				Name        string `json:"name"`
				Supported   bool   `json:"supported"`
				CipherSuite string `json:"cipher_suite"`
			} `json:"versions"`
			MinSupported string `json:"min_supported"`
			MaxSupported string `json:"max_supported"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v\n%s", err, out)
		}
		if result.Host != host || result.Port != port {
			t.Errorf("host/port = %s:%d, want %s:%d", result.Host, result.Port, host, port)
		}
		if len(result.Versions) != 4 {
			t.Fatalf("expected 4 versions, got %d", len(result.Versions))
		}
		want := map[string]bool{"TLS 1.0": false, "TLS 1.1": false, "TLS 1.2": true, "TLS 1.3": true}
		for _, v := range result.Versions {
			if v.Supported != want[v.Name] {
				t.Errorf("%s supported = %v, want %v", v.Name, v.Supported, want[v.Name])
			}
			if v.Supported && v.CipherSuite == "" {
				t.Errorf("%s is supported but has no cipher suite", v.Name)
			}
		}
		if result.MinSupported != "TLS 1.2" || result.MaxSupported != "TLS 1.3" {
			t.Errorf("min/max = %q/%q, want TLS 1.2/TLS 1.3", result.MinSupported, result.MaxSupported)
		}
	})

	t.Run("URLTargetStripsSchemeAndPath", func(t *testing.T) {
		resetTLSFlags(t)
		setOutputMode(t, true, false)

		var err error
		out := captureStdout(t, func() {
			err = tlsCmd.RunE(tlsCmd, []string{fmt.Sprintf("https://%s:%d/some/path", host, port)})
		})
		if err != nil {
			t.Fatalf("tls command failed: %v", err)
		}
		var result struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if result.Host != host || result.Port != port {
			t.Errorf("host/port = %s:%d, want %s:%d", result.Host, result.Port, host, port)
		}
	})

	t.Run("DeprecatedVersionsWarn", func(t *testing.T) {
		// Go refuses to serve TLS 1.0/1.1 by default in recent versions, so
		// this only checks the warning when the server actually accepts them.
		oldHost, oldPort := startTLSServer(t, [][]byte{leaf.der}, leaf.key, tls.VersionTLS10, tls.VersionTLS13)
		resetTLSFlags(t)
		setOutputMode(t, false, true)
		tlsPort = oldPort

		var err error
		out := captureStdout(t, func() {
			err = tlsCmd.RunE(tlsCmd, []string{oldHost})
		})
		if err != nil {
			t.Fatalf("tls command failed: %v", err)
		}
		if strings.Contains(out, "TLS 1.0 is enabled") != strings.Contains(out, "Security Warning") {
			t.Errorf("deprecated-version warning inconsistent:\n%s", out)
		}
	})

	t.Run("SilentServerTimesOut", func(t *testing.T) {
		silentHost, silentPort := startSilentListener(t)
		resetTLSFlags(t)
		setOutputMode(t, true, false)
		tlsPort = silentPort
		tlsTimeout = 200 * time.Millisecond

		start := time.Now()
		var err error
		out := captureStdout(t, func() {
			err = tlsCmd.RunE(tlsCmd, []string{silentHost})
		})
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("tls command should not error on unsupported versions: %v", err)
		}
		if elapsed > 5*time.Second {
			t.Errorf("probes should time out promptly and run concurrently, took %v", elapsed)
		}
		var result struct {
			Versions []struct {
				Supported bool   `json:"supported"`
				Error     string `json:"error"`
			} `json:"versions"`
			MinSupported string `json:"min_supported"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		for i, v := range result.Versions {
			if v.Supported || v.Error == "" {
				t.Errorf("version %d should be unsupported with an error, got %+v", i, v)
			}
		}
		if result.MinSupported != "" {
			t.Errorf("min_supported should be empty, got %q", result.MinSupported)
		}
	})
}
