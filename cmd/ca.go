package cmd

import (
	"fmt"
	"path/filepath"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	caCN      string
	caOrg     string
	caCountry string
	caDays    int
	caKeySize int
	caOutput  string
)

var caCmd = &cobra.Command{
	Use:   "ca",
	Short: "Create a Certificate Authority (CA) certificate",
	Long: `Create a self-signed Certificate Authority (CA) certificate and private key.

A CA certificate can be used to sign other certificates, creating a chain of trust.
This is useful for internal PKI, development environments, or testing.

Examples:
  # Create a basic CA certificate
  cert ca --cn "My Company CA"
  
  # Create a CA with organization details
  cert ca --cn "Example Corp Root CA" --org "Example Corporation" --country US
  
  # Create a CA with custom validity period (10 years)
  cert ca --cn "Internal CA" --days 3650
  
  # Create a CA with larger key size for extra security
  cert ca --cn "Secure CA" --key-size 4096 --output /etc/pki/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if caCN == "" {
			return fmt.Errorf("common name (--cn) is required")
		}

		// Prepare options
		options := cert.CAOptions{
			CommonName:   caCN,
			Organization: caOrg,
			Country:      caCountry,
			Days:         caDays,
			KeySize:      caKeySize,
		}

		// Set output path
		if caOutput == "" {
			caOutput = "."
		}

		// Generate CA certificate
		fmt.Println("🔐 Generating Certificate Authority...")

		certPath := filepath.Join(caOutput, sanitizeCAFilename(caCN)+"-ca.crt")
		keyPath := filepath.Join(caOutput, sanitizeCAFilename(caCN)+"-ca.key")

		err := cert.GenerateCA(options, certPath, keyPath)
		if err != nil {
			return fmt.Errorf("failed to generate CA: %w", err)
		}

		// Display success message
		ui.ShowSuccess("Certificate Authority generated successfully!")
		fmt.Println()
		fmt.Println("📁 Files created:")
		fmt.Printf("  🏛️  CA Certificate: %s\n", certPath)
		fmt.Printf("  🔑 CA Private Key: %s\n", keyPath)
		fmt.Println()
		fmt.Println("⚠️  Security Notes:")
		fmt.Println("  • Keep the CA private key extremely secure")
		fmt.Println("  • Never share the CA private key")
		fmt.Println("  • Consider storing the key offline or in an HSM")
		fmt.Println()
		fmt.Println("📋 Next steps:")
		fmt.Println("  1. Distribute the CA certificate to clients that need to trust it")
		fmt.Println("  2. Use 'cert sign' command to sign CSRs with this CA")
		fmt.Println("  3. Keep the CA key secure and backed up")

		// Display the CA certificate details
		fmt.Println()
		fmt.Println("🔍 CA Certificate Details:")
		caCert, err := cert.InspectFile(certPath)
		if err != nil {
			ui.ShowInfo(fmt.Sprintf("Could not display CA details: %v", err))
		} else {
			ui.DisplayCertificate(caCert, false)
		}

		return nil
	},
}

func init() {
	caCmd.Flags().StringVar(&caCN, "cn", "", "Common Name for the CA (required)")
	caCmd.Flags().StringVar(&caOrg, "org", "", "Organization name")
	caCmd.Flags().StringVar(&caCountry, "country", "", "Country (2-letter code)")
	caCmd.Flags().IntVarP(&caDays, "days", "d", 3650, "Validity period in days (default 10 years)")
	caCmd.Flags().IntVarP(&caKeySize, "key-size", "k", 4096, "RSA key size in bits")
	caCmd.Flags().StringVarP(&caOutput, "output", "o", "", "Output directory for CA files")

	rootCmd.AddCommand(caCmd)
}

func sanitizeCAFilename(name string) string {
	// Reuse the sanitize function from csr.go
	return sanitizeFilename(name)
}
