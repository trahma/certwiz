package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCSRCommand(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "certwiz-csr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test basic CSR generation
	t.Run("BasicCSR", func(t *testing.T) {
		csrCN = "test.example.com"
		csrOutput = tmpDir
		csrKeySize = 2048
		csrSANs = []string{}
		
		// Run the command
		err := csrCmd.RunE(csrCmd, []string{})
		if err != nil {
			t.Fatalf("CSR generation failed: %v", err)
		}

		// Check if files were created
		csrPath := filepath.Join(tmpDir, "test.example.com.csr")
		keyPath := filepath.Join(tmpDir, "test.example.com.key")

		if _, err := os.Stat(csrPath); os.IsNotExist(err) {
			t.Errorf("CSR file was not created: %s", csrPath)
		}

		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			t.Errorf("Key file was not created: %s", keyPath)
		}
	})

	// Test CSR with SANs
	t.Run("CSRWithSANs", func(t *testing.T) {
		csrCN = "multi.example.com"
		csrOutput = tmpDir
		csrKeySize = 2048
		csrSANs = []string{"multi.example.com", "www.multi.example.com", "IP:192.168.1.1"}
		
		err := csrCmd.RunE(csrCmd, []string{})
		if err != nil {
			t.Fatalf("CSR generation with SANs failed: %v", err)
		}

		// Check if files were created
		csrPath := filepath.Join(tmpDir, "multi.example.com.csr")
		if _, err := os.Stat(csrPath); os.IsNotExist(err) {
			t.Errorf("CSR file with SANs was not created: %s", csrPath)
		}
	})

	// Test CSR with organization details
	t.Run("CSRWithOrgDetails", func(t *testing.T) {
		csrCN = "org.example.com"
		csrOrg = "Test Organization"
		csrCountry = "US"
		csrState = "California"
		csrOutput = tmpDir
		csrKeySize = 2048
		
		err := csrCmd.RunE(csrCmd, []string{})
		if err != nil {
			t.Fatalf("CSR generation with org details failed: %v", err)
		}

		// Check if files were created
		csrPath := filepath.Join(tmpDir, "org.example.com.csr")
		if _, err := os.Stat(csrPath); os.IsNotExist(err) {
			t.Errorf("CSR file with org details was not created: %s", csrPath)
		}
	})

	// Test missing common name
	t.Run("MissingCN", func(t *testing.T) {
		csrCN = ""
		csrOutput = tmpDir
		
		err := csrCmd.RunE(csrCmd, []string{})
		if err == nil {
			t.Error("Expected error for missing common name, but got none")
		}
	})
}