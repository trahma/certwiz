package cmd

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"certwiz/internal/testutil"
)

func TestVerifyCommandRun(t *testing.T) {
	validCert := testutil.TestdataPath("valid.pem")
	validKey := testutil.TestdataPath("valid.key")
	otherKey := testutil.TestdataPath("strong.key")

	resetVerifyFlags := func(t *testing.T) {
		t.Helper()
		ca, host, key, expires := verifyCA, verifyHost, verifyKey, verifyExpiresIn
		verifyCA, verifyHost, verifyKey, verifyExpiresIn = "", "", "", ""
		t.Cleanup(func() { verifyCA, verifyHost, verifyKey, verifyExpiresIn = ca, host, key, expires })
	}

	t.Run("ValidCertificatePasses", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, false, true)

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{validCert})
		})
		if err != nil {
			t.Fatalf("verify failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "Certificate is valid") {
			t.Errorf("expected success message, got:\n%s", out)
		}
	})

	t.Run("ExpiresInThresholdFails", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, false, true)
		verifyExpiresIn = "36500d" // far beyond any test certificate's validity

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{validCert})
		})
		if err == nil || err.Error() != "verification failed" {
			t.Fatalf("expected 'verification failed', got %v", err)
		}
		if !strings.Contains(out, "within the 36500-day threshold") {
			t.Errorf("expected threshold message, got:\n%s", out)
		}
	})

	t.Run("InvalidExpiresInValue", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, true, false)
		verifyExpiresIn = "soon"

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{validCert})
		})
		if err == nil {
			t.Fatal("expected an error for an invalid --expires-in value")
		}
		if !strings.Contains(out, `"success": false`) {
			t.Errorf("expected JSON error payload, got:\n%s", out)
		}
	})

	t.Run("MatchingKeyPasses", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, true, false)
		verifyKey = validKey

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{validCert})
		})
		if err != nil {
			t.Fatalf("verify with matching key failed: %v\n%s", err, out)
		}
		var result struct {
			IsValid    bool  `json:"is_valid"`
			KeyMatches *bool `json:"key_matches"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if !result.IsValid || result.KeyMatches == nil || !*result.KeyMatches {
			t.Errorf("expected is_valid and key_matches true, got %s", out)
		}
	})

	t.Run("MismatchedKeyFails", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, false, true)
		verifyKey = otherKey

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{validCert})
		})
		if err == nil || err.Error() != "verification failed" {
			t.Fatalf("expected 'verification failed', got %v", err)
		}
		if !strings.Contains(out, "Private key does not match certificate") {
			t.Errorf("expected key mismatch message, got:\n%s", out)
		}
		if !strings.Contains(out, "Private key match: FAIL") {
			t.Errorf("expected key match check to fail, got:\n%s", out)
		}
	})

	t.Run("HostnameMismatch", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, false, true)
		verifyHost = "wrong.example.org"

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{validCert})
		})
		if err == nil {
			t.Fatal("expected hostname verification to fail")
		}
		if !strings.Contains(out, "Hostname verification failed") {
			t.Errorf("expected hostname failure message, got:\n%s", out)
		}
	})

	t.Run("ChainWithCA", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, true, false)
		verifyCA = testutil.TestdataPath("ca.pem")

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{testutil.TestdataPath("chain-server.pem")})
		})
		// The chain-server certificate is signed by the intermediate, so a
		// root-only bundle fails chain verification; a full bundle succeeds.
		if err == nil {
			t.Fatalf("root-only CA verification should fail:\n%s", out)
		}
		if !strings.Contains(out, "Chain verification failed") {
			t.Errorf("expected chain failure message, got:\n%s", out)
		}

		resetVerifyFlags(t)
		verifyCA = testutil.TestdataPath("fullchain.pem")
		out = captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{testutil.TestdataPath("chain-server.pem")})
		})
		if err != nil {
			t.Fatalf("verify against full chain failed: %v\n%s", err, out)
		}
	})

	t.Run("MissingFile", func(t *testing.T) {
		resetVerifyFlags(t)
		setOutputMode(t, true, false)

		var err error
		out := captureStdout(t, func() {
			err = verifyCmd.RunE(verifyCmd, []string{filepath.Join(t.TempDir(), "missing.pem")})
		})
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("expected a does-not-exist error, got %v", err)
		}
		if !strings.Contains(out, `"success": false`) {
			t.Errorf("expected JSON error payload, got:\n%s", out)
		}
	})
}
