package cert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"certwiz/internal/testutil"
)

func TestFingerprints(t *testing.T) {
	cert, err := InspectFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatalf("InspectFile failed: %v", err)
	}

	sha256fp := cert.FingerprintSHA256()
	sha1fp := cert.FingerprintSHA1()

	// SHA-256: 32 bytes as colon-separated hex = 32*2 + 31 separators
	if len(sha256fp) != 95 {
		t.Errorf("SHA-256 fingerprint length = %d, want 95: %q", len(sha256fp), sha256fp)
	}
	// SHA-1: 20 bytes as colon-separated hex = 20*2 + 19 separators
	if len(sha1fp) != 59 {
		t.Errorf("SHA-1 fingerprint length = %d, want 59: %q", len(sha1fp), sha1fp)
	}

	for _, fp := range []string{sha256fp, sha1fp} {
		for _, part := range strings.Split(fp, ":") {
			if len(part) != 2 {
				t.Errorf("Fingerprint has malformed byte group %q in %q", part, fp)
			}
		}
		if fp != strings.ToUpper(fp) {
			t.Errorf("Fingerprint should be uppercase: %q", fp)
		}
	}

	// Fingerprints must be stable across calls
	if cert.FingerprintSHA256() != sha256fp {
		t.Error("SHA-256 fingerprint is not stable across calls")
	}

	// JSON output must carry the fingerprints
	jc := cert.ToJSON()
	if jc.FingerprintSHA256 != sha256fp || jc.FingerprintSHA1 != sha1fp {
		t.Error("JSON output fingerprints do not match certificate fingerprints")
	}
}

func TestInspectFileAllBundle(t *testing.T) {
	certs, err := InspectFileAll(testutil.TestdataPath("fullchain.pem"))
	if err != nil {
		t.Fatalf("InspectFileAll failed: %v", err)
	}
	if len(certs) != 3 {
		t.Fatalf("Expected 3 certificates in fullchain.pem, got %d", len(certs))
	}
	for i, c := range certs {
		if c.Format != FormatPEM {
			t.Errorf("cert[%d] format = %s, want PEM", i, c.Format)
		}
		if c.Source != testutil.TestdataPath("fullchain.pem") {
			t.Errorf("cert[%d] source = %s", i, c.Source)
		}
	}

	// Single-cert files still return exactly one certificate
	single, err := InspectFileAll(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatalf("InspectFileAll on single cert failed: %v", err)
	}
	if len(single) != 1 {
		t.Errorf("Expected 1 certificate in valid.pem, got %d", len(single))
	}

	// InspectFile keeps returning the first certificate of a bundle
	first, err := InspectFile(testutil.TestdataPath("fullchain.pem"))
	if err != nil {
		t.Fatalf("InspectFile on bundle failed: %v", err)
	}
	if first.Subject.CommonName != certs[0].Subject.CommonName {
		t.Error("InspectFile should return the first certificate of a bundle")
	}
}

func TestInspectData(t *testing.T) {
	data, err := os.ReadFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}

	certs, err := InspectData(data, "stdin")
	if err != nil {
		t.Fatalf("InspectData failed: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("Expected 1 certificate, got %d", len(certs))
	}
	if certs[0].Source != "stdin" {
		t.Errorf("Source = %s, want stdin", certs[0].Source)
	}

	if _, err := InspectData([]byte("not a certificate"), "stdin"); err == nil {
		t.Error("Expected error for invalid data")
	}

	// PEM data without certificate blocks (e.g. a key file) should error clearly
	keyData, err := os.ReadFile(testutil.TestdataPath("valid.key"))
	if err != nil {
		t.Fatalf("Failed to read key testdata: %v", err)
	}
	if _, err := InspectData(keyData, "stdin"); err == nil {
		t.Error("Expected error for PEM data without certificates")
	}
}

func TestVerifyWithKey(t *testing.T) {
	tests := []struct {
		name        string
		certFile    string
		keyFile     string
		wantMatch   bool
		expectError bool
	}{
		{
			name:      "Matching key",
			certFile:  "valid.pem",
			keyFile:   "valid.key",
			wantMatch: true,
		},
		{
			name:      "Mismatched key",
			certFile:  "valid.pem",
			keyFile:   "strong.key",
			wantMatch: false,
		},
		{
			name:        "Invalid key file",
			certFile:    "valid.pem",
			keyFile:     "invalid.pem",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyWithOptions(VerifyOptions{
				CertPath: testutil.TestdataPath(tt.certFile),
				KeyPath:  testutil.TestdataPath(tt.keyFile),
			})
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("VerifyWithOptions failed: %v", err)
			}
			if !result.KeyChecked {
				t.Error("Expected KeyChecked to be true")
			}
			if result.KeyMatches != tt.wantMatch {
				t.Errorf("KeyMatches = %v, want %v", result.KeyMatches, tt.wantMatch)
			}
			if tt.wantMatch && !result.IsValid {
				t.Errorf("Expected valid result, errors: %v", result.Errors)
			}
			if !tt.wantMatch && result.IsValid {
				t.Error("Expected invalid result for mismatched key")
			}

			// JSON output should carry the key match result
			jr := result.ToJSON()
			if jr.KeyMatches == nil {
				t.Error("Expected key_matches in JSON output")
			} else if *jr.KeyMatches != tt.wantMatch {
				t.Errorf("JSON key_matches = %v, want %v", *jr.KeyMatches, tt.wantMatch)
			}
		})
	}

	// Without --key, no key check should be reported
	result, err := VerifyWithOptions(VerifyOptions{CertPath: testutil.TestdataPath("valid.pem")})
	if err != nil {
		t.Fatalf("VerifyWithOptions failed: %v", err)
	}
	if result.KeyChecked {
		t.Error("KeyChecked should be false when no key is provided")
	}
	if jr := result.ToJSON(); jr.KeyMatches != nil {
		t.Error("key_matches should be omitted when no key is provided")
	}
}

func TestVerifyExpiresIn(t *testing.T) {
	// Generate a certificate valid for 10 days
	tmpDir := t.TempDir()
	err := Generate(GenerateOptions{
		CommonName: "expiring.local",
		Days:       10,
		KeySize:    2048,
		OutputDir:  tmpDir,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	certPath := filepath.Join(tmpDir, "expiring.local.crt")

	// Threshold larger than remaining validity: should fail
	result, err := VerifyWithOptions(VerifyOptions{
		CertPath:  certPath,
		ExpiresIn: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("VerifyWithOptions failed: %v", err)
	}
	if result.IsValid {
		t.Error("Expected invalid result for certificate expiring within threshold")
	}

	// Threshold smaller than remaining validity: should pass
	result, err = VerifyWithOptions(VerifyOptions{
		CertPath:  certPath,
		ExpiresIn: 5 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("VerifyWithOptions failed: %v", err)
	}
	if !result.IsValid {
		t.Errorf("Expected valid result, errors: %v", result.Errors)
	}

	// Already-expired certificates report expiry, not the threshold
	result, err = VerifyWithOptions(VerifyOptions{
		CertPath:  testutil.TestdataPath("expired.pem"),
		ExpiresIn: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("VerifyWithOptions failed: %v", err)
	}
	if result.IsValid {
		t.Error("Expected invalid result for expired certificate")
	}
	if len(result.Errors) != 1 || result.Errors[0] != "Certificate has expired" {
		t.Errorf("Expected single expiry error, got %v", result.Errors)
	}
}
