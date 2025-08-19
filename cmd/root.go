package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cert",
	Short: "A user-friendly CLI tool for certificate management",
	Long: `cert (from certwiz) is a user-friendly CLI tool for certificate management.
Similar to HTTPie but for certificates.

Examples:
  # Inspect a certificate file
  cert inspect cert.pem
  cert inspect cert.der --full
  
  # Inspect a website's certificate  
  cert inspect google.com
  cert inspect https://example.com:8443
  
  # Generate a self-signed certificate
  cert generate --cn example.com
  cert generate --cn myserver --san example.com --san *.example.com --san IP:192.168.1.1
  
  # Convert certificate formats
  cert convert cert.pem cert.der --format der
  cert convert cert.der cert.pem --format pem
  
  # Verify a certificate
  cert verify cert.pem
  cert verify cert.pem --host example.com`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(verifyCmd)
}

// checkError is a helper function to handle errors consistently
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}