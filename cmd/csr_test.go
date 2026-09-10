package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCSRCommand(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "certwiz-csr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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

		info, err := os.Stat(keyPath)
		if os.IsNotExist(err) {
			t.Fatalf("Key file was not created: %s", keyPath)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
			t.Errorf("Key file permissions = %v, want 0600", info.Mode().Perm())
		}
	})

	// Test that the generated CSR is read back and displayed
	t.Run("DisplaysCSRDetails", func(t *testing.T) {
		csrCN = "display.example.com"
		csrOrg = "Display Org"
		csrCountry = ""
		csrState = ""
		csrOutput = tmpDir
		csrKeySize = 2048
		csrSANs = []string{"display.example.com", "IP:10.0.0.1"}
		setOutputMode(t, false, true)

		var err error
		out := captureStdout(t, func() {
			err = csrCmd.RunE(csrCmd, []string{})
		})
		if err != nil {
			t.Fatalf("CSR generation failed: %v", err)
		}
		for _, want := range []string{"CSR Details", "CN=display.example.com", "O=Display Org", "RSA 2048 bits", "10.0.0.1"} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
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
