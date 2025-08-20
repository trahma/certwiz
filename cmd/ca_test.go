package cmd

import (
	"certwiz/pkg/cert"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCACommand(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "certwiz-ca-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test basic CA generation
	t.Run("BasicCA", func(t *testing.T) {
		caCN = "Test CA"
		caOutput = tmpDir
		caKeySize = 2048 // Use smaller key for faster tests
		caDays = 365

		// Run the command
		err := caCmd.RunE(caCmd, []string{})
		if err != nil {
			t.Fatalf("CA generation failed: %v", err)
		}

		// Check if files were created
		certPath := filepath.Join(tmpDir, "Test_CA-ca.crt")
		keyPath := filepath.Join(tmpDir, "Test_CA-ca.key")

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			t.Errorf("CA certificate file was not created: %s", certPath)
		}

		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			t.Errorf("CA key file was not created: %s", keyPath)
		}

		// Verify the certificate is a CA
		caCert, err := cert.InspectFile(certPath)
		if err != nil {
			t.Fatalf("Failed to inspect CA certificate: %v", err)
		}

		if !caCert.IsCA {
			t.Error("Generated certificate is not marked as CA")
		}

		// Check key permissions (only on Unix-like systems)
		// Windows has different permission semantics
		if runtime.GOOS != "windows" {
			info, err := os.Stat(keyPath)
			if err != nil {
				t.Fatalf("Failed to stat key file: %v", err)
			}

			// On Unix systems, check that permissions are restricted
			if info.Mode().Perm() != 0600 {
				t.Errorf("CA key file has incorrect permissions: %v, expected 0600", info.Mode().Perm())
			}
		}
	})

	// Test CA with organization details
	t.Run("CAWithOrgDetails", func(t *testing.T) {
		caCN = "Company Root CA"
		caOrg = "Test Company"
		caCountry = "US"
		caOutput = tmpDir
		caKeySize = 2048
		caDays = 3650

		err := caCmd.RunE(caCmd, []string{})
		if err != nil {
			t.Fatalf("CA generation with org details failed: %v", err)
		}

		certPath := filepath.Join(tmpDir, "Company_Root_CA-ca.crt")
		caCert, err := cert.InspectFile(certPath)
		if err != nil {
			t.Fatalf("Failed to inspect CA certificate: %v", err)
		}

		// Check organization in subject
		if len(caCert.Subject.Organization) == 0 || caCert.Subject.Organization[0] != "Test Company" {
			t.Errorf("CA certificate organization mismatch: got %v, expected [Test Company]",
				caCert.Subject.Organization)
		}

		// Check country in subject
		if len(caCert.Subject.Country) == 0 || caCert.Subject.Country[0] != "US" {
			t.Errorf("CA certificate country mismatch: got %v, expected [US]",
				caCert.Subject.Country)
		}
	})

	// Test missing common name
	t.Run("MissingCN", func(t *testing.T) {
		caCN = ""
		caOutput = tmpDir

		err := caCmd.RunE(caCmd, []string{})
		if err == nil {
			t.Error("Expected error for missing common name, but got none")
		}
	})
}
