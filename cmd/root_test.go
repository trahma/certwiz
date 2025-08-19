package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		expectedOutput []string
	}{
		{
			name:    "No arguments shows help",
			args:    []string{},
			wantErr: false,
			expectedOutput: []string{
				"cert",
				"A user-friendly CLI tool for certificate management",
				"Available Commands:",
				"inspect",
				"generate",
				"convert",
				"verify",
			},
		},
		{
			name:    "Help flag",
			args:    []string{"--help"},
			wantErr: false,
			expectedOutput: []string{
				"cert",
				"A user-friendly CLI tool for certificate management",
			},
		},
		{
			name:    "Version command",
			args:    []string{"version"},
			wantErr: false,
			expectedOutput: []string{
				"cert version 0.1.0",
			},
		},
		{
			name:    "Invalid command",
			args:    []string{"invalid"},
			wantErr: true,
			expectedOutput: []string{
				"unknown command",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command for testing
			rootCmd.SetArgs(tt.args)

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			output := stdout.String() + stderr.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	// Verify all expected commands are registered
	expectedCommands := []string{
		"inspect",
		"generate",
		"convert",
		"verify",
		"version",
	}

	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected command %q not found", expected)
		}
	}

	// Check that we don't have extra unexpected commands
	if len(commands) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(commands))
	}
}