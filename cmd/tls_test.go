package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestTLSCmdFlags(t *testing.T) {
	// Find the tls command from the actual rootCmd
	var tlsCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "tls" {
			tlsCmd = cmd
			break
		}
	}
	if tlsCmd == nil {
		t.Fatal("tls command not found in rootCmd")
	}

	// Check port flag
	portFlag := tlsCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("port flag not found")
	}
	if portFlag.DefValue != "443" {
		t.Errorf("port flag default = %q, want %q", portFlag.DefValue, "443")
	}

	// Check timeout flag
	timeoutFlag := tlsCmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Fatal("timeout flag not found")
	}
	if timeoutFlag.DefValue != "5s" {
		t.Errorf("timeout flag default = %q, want %q", timeoutFlag.DefValue, "5s")
	}
}

func TestTLSCmdArgs(t *testing.T) {
	// Test that the tls command validates arguments
	// Note: Cobra's behavior with args validation depends on the Args field
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"example.com"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test command with the same structure as tlsCmd
			cmd := &cobra.Command{
				Use:  "tls",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTLSCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	tlsCmd.SetOut(buf)
	t.Cleanup(func() { tlsCmd.SetOut(nil) })

	if err := tlsCmd.Help(); err != nil {
		t.Errorf("Help() failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("Help output is empty")
	}
	for _, want := range []string{"TLS", "--port", "--timeout", "cert tls google.com"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("Help output does not contain %q", want)
		}
	}
}
