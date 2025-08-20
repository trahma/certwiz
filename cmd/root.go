package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.8"

var (
	versionFlag bool
	jsonOutput  bool
)

var rootCmd = &cobra.Command{
	Use:   "cert",
	Short: "A user-friendly CLI tool for certificate management",
	Long: `cert (from certwiz) is a user-friendly CLI tool for certificate management.
Similar to HTTPie but for certificates.`,
	Example: `  # Inspect a certificate file
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
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("cert version %s\n", version)
			return
		}
		
		// Check if help was explicitly requested
		helpRequested := false
		for _, arg := range os.Args[1:] {
			if arg == "help" || arg == "--help" || arg == "-h" {
				helpRequested = true
				break
			}
		}
		
		// If running without arguments (no help requested), show simplified output
		if !helpRequested && len(os.Args) == 1 {
			// Show simplified usage without examples
			fmt.Println("cert (from certwiz) - A user-friendly CLI tool for certificate management")
			fmt.Println("\nUsage:")
			fmt.Println("  cert [command]")
			fmt.Println("\nAvailable Commands:")
			fmt.Println("  ca          Create a Certificate Authority (CA) certificate")
			fmt.Println("  completion  Generate the autocompletion script for the specified shell")
			fmt.Println("  convert     Convert certificates between different formats")
			fmt.Println("  csr         Generate a Certificate Signing Request (CSR)")
			fmt.Println("  generate    Generate a self-signed certificate")
			fmt.Println("  help        Help about any command")
			fmt.Println("  inspect     Inspect a certificate from a file or URL")
			fmt.Println("  sign        Sign a Certificate Signing Request (CSR) with a CA")
			fmt.Println("  update      Update cert to the latest version")
			fmt.Println("  verify      Verify a certificate")
			fmt.Println("  version     Print the version of cert")
			fmt.Println("\nFlags:")
			fmt.Println("  -h, --help      help for cert")
			fmt.Println("      --json      Output in JSON format")
			fmt.Println("  -v, --version   Print version information")
			fmt.Println("\nUse \"cert [command] --help\" for more information about a command.")
			fmt.Println("Use \"cert help\" to see examples.")
			return
		}
		
		// Otherwise show full help with examples
		_ = cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// Add version flag
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")

	// Add subcommands
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of cert",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cert version %s\n", version)
	},
}
