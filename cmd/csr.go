package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	csrCN       string
	csrOrg      string
	csrOrgUnit  string
	csrCountry  string
	csrState    string
	csrLocality string
	csrEmail    string
	csrSANs     []string
	csrKeySize  int
	csrOutput   string
)

var csrCmd = &cobra.Command{
	Use:   "csr",
	Short: "Generate a Certificate Signing Request (CSR)",
	Long: `Generate a Certificate Signing Request (CSR) and private key.

A CSR is used to request a certificate from a Certificate Authority (CA).
It contains your public key and identity information.

Examples:
  # Basic CSR generation
  cert csr --cn example.com
  
  # CSR with organization details
  cert csr --cn example.com --org "Example Inc" --country US --state CA
  
  # CSR with Subject Alternative Names
  cert csr --cn example.com --san example.com --san www.example.com --san api.example.com
  
  # CSR with custom output directory and key size
  cert csr --cn secure.example.com --key-size 4096 --output /etc/ssl/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if csrCN == "" {
			return fmt.Errorf("common name (--cn) is required")
		}

		// Prepare options
		options := cert.CSROptions{
			CommonName:         csrCN,
			Organization:       csrOrg,
			OrganizationalUnit: csrOrgUnit,
			Country:            csrCountry,
			Province:           csrState,
			Locality:           csrLocality,
			EmailAddress:       csrEmail,
			SANs:               processSANs(csrSANs),
			KeySize:            csrKeySize,
		}

		// Set output path
		if csrOutput == "" {
			csrOutput = "."
		}

		// Generate CSR
		fmt.Printf("%s Generating Certificate Signing Request...\n", getEmoji("🔐", "[CSR]"))

		csrPath := filepath.Join(csrOutput, sanitizeFilename(csrCN)+".csr")
		keyPath := filepath.Join(csrOutput, sanitizeFilename(csrCN)+".key")

		err := cert.GenerateCSR(options, csrPath, keyPath)
		if err != nil {
			return fmt.Errorf("failed to generate CSR: %w", err)
		}

		// Display success message
		ui.ShowSuccess("Certificate Signing Request generated successfully!")
		fmt.Println()
		fmt.Printf("%s Files created:\n", getEmoji("📁", "[FILES]"))
		fmt.Printf("  %s CSR:         %s\n", getEmoji("📄", "[CSR]"), csrPath)
		fmt.Printf("  %s Private Key: %s\n", getEmoji("🔑", "[KEY]"), keyPath)
		fmt.Println()
		fmt.Printf("%s Next steps:\n", getEmoji("📋", "[NEXT]"))
		fmt.Println("  1. Submit the CSR to your Certificate Authority")
		fmt.Println("  2. Keep the private key secure - you'll need it with the signed certificate")
		fmt.Println("  3. Once you receive the signed certificate, install it with the private key")

		// Optionally display the CSR details
		fmt.Println()
		fmt.Printf("%s CSR Details:\n", getEmoji("🔍", "[INFO]"))
		if err := displayCSRInfo(csrPath); err != nil {
			ui.ShowInfo(fmt.Sprintf("Could not display CSR details: %v", err))
		}

		return nil
	},
}

func init() {
	csrCmd.Flags().StringVar(&csrCN, "cn", "", "Common Name (required)")
	csrCmd.Flags().StringVar(&csrOrg, "org", "", "Organization")
	csrCmd.Flags().StringVar(&csrOrgUnit, "org-unit", "", "Organizational Unit")
	csrCmd.Flags().StringVar(&csrCountry, "country", "", "Country (2-letter code)")
	csrCmd.Flags().StringVar(&csrState, "state", "", "State or Province")
	csrCmd.Flags().StringVar(&csrLocality, "locality", "", "Locality or City")
	csrCmd.Flags().StringVar(&csrEmail, "email", "", "Email Address")
	csrCmd.Flags().StringSliceVar(&csrSANs, "san", []string{}, "Subject Alternative Name (can be used multiple times)")
	csrCmd.Flags().IntVarP(&csrKeySize, "key-size", "k", 2048, "RSA key size in bits")
	csrCmd.Flags().StringVarP(&csrOutput, "output", "o", "", "Output directory for CSR and key files")

	rootCmd.AddCommand(csrCmd)
}

func displayCSRInfo(csrPath string) error {
	data, err := os.ReadFile(csrPath)
	if err != nil {
		return err
	}

	// Parse and display CSR info
	info, err := cert.ParseCSR(data)
	if err != nil {
		return err
	}

	ui.DisplayCSRInfo(info)
	return nil
}

func sanitizeFilename(name string) string {
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

func processSANs(sans []string) []string {
	// Just return the SANs as-is, they'll be processed in the cert package
	return sans
}
