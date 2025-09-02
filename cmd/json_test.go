package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONOutput(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "certwiz-json-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test generate command with JSON output
	t.Run("GenerateJSON", func(t *testing.T) {
		// Set up command parameters
		generateCN = "test.example.com"
		generateOutput = tmpDir
		generateKeySize = 2048
		generateDays = 365
		jsonOutput = true

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

        // Run the command
        _ = generateCmd.RunE(generateCmd, []string{})

		// Restore stdout and read output
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)

		// Parse JSON output
		var result map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		// Verify JSON structure
		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected success: true in JSON output")
		}

		if files, ok := result["files"].([]interface{}); !ok || len(files) != 2 {
			t.Error("Expected 2 files in JSON output")
		}

		// Reset for other tests
		jsonOutput = false
	})

	// Test inspect command with JSON output
	t.Run("InspectJSON", func(t *testing.T) {
		// First generate a certificate
		generateCN = "inspect-test.local"
		generateOutput = tmpDir
		generateKeySize = 2048
		generateDays = 365
		jsonOutput = false

			_ = generateCmd.RunE(generateCmd, []string{})

		// Now inspect it with JSON output
		certPath := filepath.Join(tmpDir, "inspect-test.local.crt")
		jsonOutput = true

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

        // Run inspect command
        _ = inspectCmd.RunE(inspectCmd, []string{certPath})

		// Restore stdout and read output
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)

		// Parse JSON output
		var cert map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &cert)
		if err != nil {
			t.Fatalf("Failed to parse certificate JSON: %v", err)
		}

		// Verify certificate fields
		if subject, ok := cert["subject"].(map[string]interface{}); ok {
			if cn, ok := subject["common_name"].(string); !ok || cn != "inspect-test.local" {
				t.Errorf("Expected common_name to be inspect-test.local, got %v", cn)
			}
		} else {
			t.Error("Missing subject in certificate JSON")
		}

		if _, ok := cert["serial_number"].(string); !ok {
			t.Error("Missing serial_number in certificate JSON")
		}

		if _, ok := cert["not_before"].(string); !ok {
			t.Error("Missing not_before in certificate JSON")
		}

		if _, ok := cert["not_after"].(string); !ok {
			t.Error("Missing not_after in certificate JSON")
		}

		// Reset for other tests
		jsonOutput = false
	})

	// Test verify command with JSON output
	t.Run("VerifyJSON", func(t *testing.T) {
		// Use the certificate generated in previous test
		certPath := filepath.Join(tmpDir, "inspect-test.local.crt")
		jsonOutput = true
		verifyHost = ""
		verifyCA = ""

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

        // Run verify command
        _ = verifyCmd.RunE(verifyCmd, []string{certPath})

		// Restore stdout and read output
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)

		// Parse JSON output
		var result map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			t.Fatalf("Failed to parse verification JSON: %v", err)
		}

		// Check verification result
		if _, ok := result["is_valid"].(bool); !ok {
			t.Error("Missing is_valid in verification JSON")
		}

		if cert, ok := result["certificate"].(map[string]interface{}); !ok {
			t.Error("Missing certificate in verification JSON")
		} else {
			if subject, ok := cert["subject"].(map[string]interface{}); ok {
				if cn, ok := subject["common_name"].(string); !ok || cn != "inspect-test.local" {
					t.Errorf("Expected certificate CN to be inspect-test.local, got %v", cn)
				}
			}
		}

		// Reset for other tests
		jsonOutput = false
	})
}
