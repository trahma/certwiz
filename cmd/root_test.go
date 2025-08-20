package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "No arguments shows help",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "Help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "Version command",
			args:    []string{"version"},
			wantErr: false,
		},
		{
			name:    "Invalid command",
			args:    []string{"invalid"},
			wantErr: true,
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

			// Note: Cobra output goes to the configured streams,
			// so we don't check specific output content in these tests
		})
	}
}

func TestCommandStructure(t *testing.T) {
	// Verify all expected commands are registered
	expectedCommands := []string{
		"completion", // Added by Cobra
		"convert",
		"generate",
		"help", // Added by Cobra
		"inspect",
		"update",
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
		t.Logf("Actual commands: %v", func() []string {
			var names []string
			for _, cmd := range commands {
				names = append(names, cmd.Name())
			}
			return names
		}())
	}
}
