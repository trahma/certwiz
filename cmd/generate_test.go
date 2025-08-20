package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		checkFiles     []string
	}{
		{
			name: "Generate basic certificate",
			args: []string{"generate", "--cn", "test.local"},
			wantErr: false,
			checkFiles: []string{"test.local.crt", "test.local.key"},
		},
		{
			name: "Generate with custom days",
			args: []string{"generate", "--cn", "test30.local", "--days", "30"},
			wantErr: false,
			checkFiles: []string{"test30.local.crt", "test30.local.key"},
		},
		{
			name: "Generate with SANs",
			args: []string{"generate", "--cn", "multi.local", "--san", "alt1.local", "--san", "alt2.local"},
			wantErr: false,
			checkFiles: []string{"multi.local.crt", "multi.local.key"},
		},
		{
			name: "Generate with custom key size",
			args: []string{"generate", "--cn", "strong.local", "--key-size", "4096"},
			wantErr: false,
			checkFiles: []string{"strong.local.crt", "strong.local.key"},
		},
		{
			name: "Generate with custom output directory",
			args: []string{"generate", "--cn", "custom.local", "--output", "custom_dir"},
			wantErr: false,
			checkFiles: []string{"custom_dir/custom.local.crt", "custom_dir/custom.local.key"},
		},
		// Skipping "Generate with no arguments" test because os.Exit(1) terminates test process
		{
			name: "Generate help",
			args: []string{"generate", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test outputs
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldDir)

			// Reset flags to ensure clean state
			generateCN = ""
			generateDays = 365
			generateKeySize = 2048
			generateSANs = []string{}
			generateOutput = "."

			// Reset the generateCmd flags
			generateCmd.Flags().Set("cn", "")
			
			// Create new root command for each test
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
			// So we only check files for successful operations

			// Check if expected files were created
			if !tt.wantErr && len(tt.checkFiles) > 0 {
				for _, file := range tt.checkFiles {
					path := filepath.Join(tempDir, file)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						t.Errorf("Expected file %s was not created", file)
					}
				}
			}
		})
	}
}

func TestGenerateCommandFlags(t *testing.T) {
	// Test that flags are properly defined
	generateCmd := generateCmd // Get the actual command

	// Check --days flag
	daysFlag := generateCmd.Flag("days")
	if daysFlag == nil {
		t.Error("--days flag not found")
	} else {
		if daysFlag.Value.Type() != "int" {
			t.Errorf("--days flag should be int, got %s", daysFlag.Value.Type())
		}
		if daysFlag.DefValue != "365" {
			t.Errorf("--days default should be 365, got %s", daysFlag.DefValue)
		}
	}

	// Check --key-size flag
	keySizeFlag := generateCmd.Flag("key-size")
	if keySizeFlag == nil {
		t.Error("--key-size flag not found")
	} else {
		if keySizeFlag.Value.Type() != "int" {
			t.Errorf("--key-size flag should be int, got %s", keySizeFlag.Value.Type())
		}
		if keySizeFlag.DefValue != "2048" {
			t.Errorf("--key-size default should be 2048, got %s", keySizeFlag.DefValue)
		}
	}

	// Check --san flag
	sanFlag := generateCmd.Flag("san")
	if sanFlag == nil {
		t.Error("--san flag not found")
	} else {
		if sanFlag.Value.Type() != "stringSlice" {
			t.Errorf("--san flag should be stringSlice, got %s", sanFlag.Value.Type())
		}
	}

	// Check --output flag
	outputFlag := generateCmd.Flag("output")
	if outputFlag == nil {
		t.Error("--output flag not found")
	} else {
		if outputFlag.Value.Type() != "string" {
			t.Errorf("--output flag should be string, got %s", outputFlag.Value.Type())
		}
		if outputFlag.DefValue != "." {
			t.Errorf("--output default should be '.', got %s", outputFlag.DefValue)
		}
	}
}