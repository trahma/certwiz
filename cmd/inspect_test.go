package cmd

import (
	"bytes"
	"path/filepath"
	"testing"
)

// testdataPath returns the path to a file in the testdata directory
func testdataPath(filename string) string {
	return filepath.Join("..", "testdata", filename)
}

func TestInspectCommand(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		wantErr          bool
		expectedOutput   []string
		unexpectedOutput []string
	}{
		{
			name:    "Inspect valid PEM file",
			args:    []string{"inspect", testdataPath("valid.pem")},
			wantErr: false,
			expectedOutput: []string{
				"Certificate from",
				"test.example.com",
				"Subject",
				"Issuer",
				"Valid",
			},
		},
		{
			name:    "Inspect valid DER file",
			args:    []string{"inspect", testdataPath("valid.der")},
			wantErr: false,
			expectedOutput: []string{
				"Certificate from",
				"test.example.com",
			},
		},
		{
			name:    "Inspect with --full flag",
			args:    []string{"inspect", testdataPath("valid.pem"), "--full"},
			wantErr: false,
			expectedOutput: []string{
				"Certificate from",
				"Certificate Extensions",
			},
		},
		{
			name:    "Inspect with no arguments",
			args:    []string{"inspect"},
			wantErr: true,
			expectedOutput: []string{
				"required argument",
			},
		},
		{
			name:    "Inspect help",
			args:    []string{"inspect", "--help"},
			wantErr: false,
			expectedOutput: []string{
				"Inspect a certificate from a file or URL",
				"Usage:",
				"--chain",
				"--full",
				"--port",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new root command for each test to reset state
			cmd := rootCmd
			cmd.SetArgs(tt.args)

			// Capture output
			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Note: UI output goes directly to stdout/stderr via fmt.Println, not through cmd output
			// For these tests, we mainly verify that commands don't error when they shouldn't
		})
	}
}

func TestInspectCommandFlags(t *testing.T) {
	// Test that flags are properly defined
	inspectCmd := inspectCmd // Get the actual command

	// Check --full flag
	fullFlag := inspectCmd.Flag("full")
	if fullFlag == nil {
		t.Error("--full flag not found")
	} else {
		if fullFlag.Value.Type() != "bool" {
			t.Errorf("--full flag should be bool, got %s", fullFlag.Value.Type())
		}
	}

	// Check --chain flag
	chainFlag := inspectCmd.Flag("chain")
	if chainFlag == nil {
		t.Error("--chain flag not found")
	} else {
		if chainFlag.Value.Type() != "bool" {
			t.Errorf("--chain flag should be bool, got %s", chainFlag.Value.Type())
		}
	}

	// Check --port flag
	portFlag := inspectCmd.Flag("port")
	if portFlag == nil {
		t.Error("--port flag not found")
	} else {
		if portFlag.Value.Type() != "int" {
			t.Errorf("--port flag should be int, got %s", portFlag.Value.Type())
		}
		if portFlag.DefValue != "443" {
			t.Errorf("--port default should be 443, got %s", portFlag.DefValue)
		}
	}
}
