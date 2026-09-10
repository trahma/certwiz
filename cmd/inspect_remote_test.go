package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"certwiz/internal/testutil"
)

func TestInspectRemote(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, &ca)
	host, port := startTLSServer(t, [][]byte{leaf.der, ca.der}, leaf.key, tls.VersionTLS12, tls.VersionTLS13)

	t.Run("HostPortTargetPlain", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, false, true)

		var err error
		out := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{fmt.Sprintf("%s:%d", host, port)})
		})
		if err != nil {
			t.Fatalf("inspect failed: %v", err)
		}
		for _, want := range []string{"Certificate for https://" + host, "CN=localhost", "TLS Version", "Cipher Suite"} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "Certificate Chain") {
			t.Error("chain should not be shown without --chain")
		}
	})

	t.Run("PortFlag", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)
		inspectPort = port

		var err error
		out := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{host})
		})
		if err != nil {
			t.Fatalf("inspect failed: %v", err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v\n%s", err, out)
		}
		if result["tls_version"] == "" || result["tls_version"] == nil {
			t.Error("JSON should include tls_version")
		}
		if result["cipher_suite"] == "" || result["cipher_suite"] == nil {
			t.Error("JSON should include cipher_suite")
		}
		if _, ok := result["chain"]; ok {
			t.Error("chain should be absent without --chain")
		}
	})

	t.Run("ConnectFlagWithUnresolvableTarget", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)
		inspectConnect = fmt.Sprintf("%s:%d", host, port)
		inspectChain = true

		var err error
		out := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{"example.invalid"})
		})
		if err != nil {
			t.Fatalf("inspect via --connect failed: %v", err)
		}
		var result struct {
			Source string `json:"source"`
			Chain  []struct {
				Subject string `json:"subject"`
				Issuer  string `json:"issuer"`
			} `json:"chain"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v\n%s", err, out)
		}
		if result.Source != "https://example.invalid" {
			t.Errorf("source = %q, want https://example.invalid", result.Source)
		}
		if len(result.Chain) != 1 {
			t.Fatalf("expected chain of length 1, got %d", len(result.Chain))
		}
		if !strings.Contains(result.Chain[0].Subject, "Test Root CA") {
			t.Errorf("chain[0].subject = %q, want the test CA", result.Chain[0].Subject)
		}
	})

	t.Run("ChainPlain", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, false, true)
		inspectChain = true

		var err error
		out := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{fmt.Sprintf("%s:%d", host, port)})
		})
		if err != nil {
			t.Fatalf("inspect failed: %v", err)
		}
		for _, want := range []string{"Certificate Chain", "Chain[1]", "CN=Test Root CA"} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
	})

	t.Run("SigAlgRSA", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)
		inspectPort = port
		inspectSigAlg = "rsa"

		var err error
		out := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{host})
		})
		if err != nil {
			t.Fatalf("inspect with --sig-alg rsa failed: %v", err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if result["tls_version"] != "TLS 1.2" {
			t.Errorf("--sig-alg rsa should cap at TLS 1.2, got %v", result["tls_version"])
		}
	})

	t.Run("SilentServerFailsWithJSONError", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)

		// A silent listener never completes the handshake; with a short
		// timeout the command must fail and report a JSON error payload.
		silentHost, silentPort := startSilentListener(t)
		inspectPort = silentPort
		inspectTimeout = 200 * time.Millisecond

		var err error
		out := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{silentHost})
		})
		if err == nil {
			t.Fatal("expected an error from a silent server")
		}
		var result map[string]interface{}
		if jerr := json.Unmarshal([]byte(out), &result); jerr != nil {
			t.Fatalf("expected a JSON error payload, got: %s", out)
		}
		if result["success"] != false {
			t.Errorf("expected success:false, got %v", result["success"])
		}
	})
}

func TestInspectLocalBundle(t *testing.T) {
	bundlePath := testutil.TestdataPath("fullchain.pem")
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read fullchain.pem: %v", err)
	}

	t.Run("StdinWithChain", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, false, true)
		inspectChain = true
		withStdin(t, data)

		var runErr error
		out := captureStdout(t, func() {
			runErr = inspectCmd.RunE(inspectCmd, []string{"-"})
		})
		if runErr != nil {
			t.Fatalf("inspect - failed: %v", runErr)
		}
		for _, want := range []string{"Certificate from stdin", "Certificate Chain", "Chain[1]", "Chain[2]"} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
	})

	t.Run("FileWithoutChainShowsHint", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, false, true)

		var runErr error
		out := captureStdout(t, func() {
			runErr = inspectCmd.RunE(inspectCmd, []string{bundlePath})
		})
		if runErr != nil {
			t.Fatalf("inspect failed: %v", runErr)
		}
		if !strings.Contains(out, "Contains 3 certificates; showing the first") {
			t.Errorf("expected bundle hint, got:\n%s", out)
		}
		if strings.Contains(out, "Certificate Chain") {
			t.Error("chain should not be rendered without --chain")
		}
	})

	t.Run("FileJSONWithChain", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)
		inspectChain = true

		var runErr error
		out := captureStdout(t, func() {
			runErr = inspectCmd.RunE(inspectCmd, []string{bundlePath})
		})
		if runErr != nil {
			t.Fatalf("inspect failed: %v", runErr)
		}
		var result struct {
			Chain []map[string]interface{} `json:"chain"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(result.Chain) != 2 {
			t.Errorf("expected 2 chain entries, got %d", len(result.Chain))
		}
	})

	t.Run("StdinInvalidData", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)
		withStdin(t, []byte("not a certificate"))

		var runErr error
		out := captureStdout(t, func() {
			runErr = inspectCmd.RunE(inspectCmd, []string{"-"})
		})
		if runErr == nil {
			t.Fatal("expected an error for invalid stdin data")
		}
		if !strings.Contains(out, `"success": false`) {
			t.Errorf("expected JSON error payload, got:\n%s", out)
		}
	})
}
