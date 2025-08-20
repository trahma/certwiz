package cmd

import (
	"certwiz/pkg/cert"
	"os"
	"path/filepath"
	"testing"
)

func TestSignCommand(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "certwiz-sign-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First, create a CA
	caOptions := cert.CAOptions{
		CommonName:   "Test CA",
		Organization: "Test Org",
		Country:      "US",
		Days:         365,
		KeySize:      2048,
	}

	caCertPath := filepath.Join(tmpDir, "ca.crt")
	caKeyPath := filepath.Join(tmpDir, "ca.key")

	err = cert.GenerateCA(caOptions, caCertPath, caKeyPath)
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	// Create a CSR
	csrOptions := cert.CSROptions{
		CommonName:   "test.example.com",
		Organization: "Test Company",
		Country:      "US",
		SANs:         []string{"test.example.com", "www.test.example.com"},
		KeySize:      2048,
	}

	csrPath := filepath.Join(tmpDir, "test.csr")
	keyPath := filepath.Join(tmpDir, "test.key")

	err = cert.GenerateCSR(csrOptions, csrPath, keyPath)
	if err != nil {
		t.Fatalf("Failed to generate CSR: %v", err)
	}

	// Test signing the CSR
	t.Run("SignCSR", func(t *testing.T) {
		signCSR = csrPath
		signCA = caCertPath
		signCAKey = caKeyPath
		signDays = 365
		signOutput = tmpDir

		err := signCmd.RunE(signCmd, []string{})
		if err != nil {
			t.Fatalf("Certificate signing failed: %v", err)
		}

		// Check if certificate was created
		certPath := filepath.Join(tmpDir, "test.crt")
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			t.Errorf("Signed certificate was not created: %s", certPath)
		}

		// Verify the certificate
		signedCert, err := cert.InspectFile(certPath)
		if err != nil {
			t.Fatalf("Failed to inspect signed certificate: %v", err)
		}

		// Check that it's not a CA
		if signedCert.IsCA {
			t.Error("Signed certificate should not be a CA")
		}

		// Check subject matches CSR
		if signedCert.Subject.CommonName != "test.example.com" {
			t.Errorf("Certificate CN mismatch: got %s, expected test.example.com",
				signedCert.Subject.CommonName)
		}

		// Check issuer matches CA
		if signedCert.Issuer.CommonName != "Test CA" {
			t.Errorf("Certificate issuer mismatch: got %s, expected Test CA",
				signedCert.Issuer.CommonName)
		}

		// Check SANs
		if len(signedCert.DNSNames) != 2 {
			t.Errorf("Certificate SAN count mismatch: got %d, expected 2",
				len(signedCert.DNSNames))
		}
	})

	// Test signing with SAN override
	t.Run("SignWithSANOverride", func(t *testing.T) {
		signCSR = csrPath
		signCA = caCertPath
		signCAKey = caKeyPath
		signDays = 365
		signOutput = tmpDir
		signSANs = []string{"override.example.com", "IP:10.0.0.1"}

		// Use a different output name to avoid overwriting
		certPath := filepath.Join(tmpDir, "test-override.crt")

		// Temporarily modify the sign command to output to a different file
		options := cert.SignOptions{
			CSRPath: csrPath,
			CACert:  caCertPath,
			CAKey:   caKeyPath,
			Days:    365,
			SANs:    signSANs,
		}

		err := cert.SignCSR(options, certPath)
		if err != nil {
			t.Fatalf("Certificate signing with SAN override failed: %v", err)
		}

		// Verify the certificate has overridden SANs
		signedCert, err := cert.InspectFile(certPath)
		if err != nil {
			t.Fatalf("Failed to inspect signed certificate: %v", err)
		}

		// Check that SANs were overridden
		if len(signedCert.DNSNames) != 1 || signedCert.DNSNames[0] != "override.example.com" {
			t.Errorf("Certificate DNS SAN override failed: got %v, expected [override.example.com]",
				signedCert.DNSNames)
		}

		if len(signedCert.IPAddresses) != 1 || signedCert.IPAddresses[0].String() != "10.0.0.1" {
			t.Errorf("Certificate IP SAN override failed: got %v, expected [10.0.0.1]",
				signedCert.IPAddresses)
		}
	})

	// Test missing required arguments
	t.Run("MissingArguments", func(t *testing.T) {
		// Test missing CSR
		signCSR = ""
		signCA = caCertPath
		signCAKey = caKeyPath

		err := signCmd.RunE(signCmd, []string{})
		if err == nil {
			t.Error("Expected error for missing CSR, but got none")
		}

		// Test missing CA cert
		signCSR = csrPath
		signCA = ""
		signCAKey = caKeyPath

		err = signCmd.RunE(signCmd, []string{})
		if err == nil {
			t.Error("Expected error for missing CA cert, but got none")
		}

		// Test missing CA key
		signCSR = csrPath
		signCA = caCertPath
		signCAKey = ""

		err = signCmd.RunE(signCmd, []string{})
		if err == nil {
			t.Error("Expected error for missing CA key, but got none")
		}
	})
}
