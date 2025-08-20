package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	signCSR    string
	signCA     string
	signCAKey  string
	signDays   int
	signOutput string
	signSANs   []string
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a Certificate Signing Request (CSR) with a CA",
	Long: `Sign a Certificate Signing Request (CSR) using a Certificate Authority (CA).

This command takes a CSR file and signs it with the specified CA certificate and key,
producing a signed certificate that can be used for TLS/SSL or other purposes.

Examples:
  # Sign a CSR with a CA
  cert sign --csr server.csr --ca ca.crt --ca-key ca.key
  
  # Sign with custom validity period (1 year)
  cert sign --csr server.csr --ca ca.crt --ca-key ca.key --days 365
  
  # Sign and output to specific directory
  cert sign --csr server.csr --ca ca.crt --ca-key ca.key --output /etc/ssl/certs/
  
  # Sign with additional SANs (overrides CSR SANs)
  cert sign --csr server.csr --ca ca.crt --ca-key ca.key --san server.local --san *.server.local`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required arguments
		if signCSR == "" {
			return fmt.Errorf("CSR file (--csr) is required")
		}
		if signCA == "" {
			return fmt.Errorf("CA certificate (--ca) is required")
		}
		if signCAKey == "" {
			return fmt.Errorf("CA private key (--ca-key) is required")
		}

		// Prepare options
		options := cert.SignOptions{
			CSRPath: signCSR,
			CACert:  signCA,
			CAKey:   signCAKey,
			Days:    signDays,
			SANs:    processSANs(signSANs),
		}

		// Set output path
		if signOutput == "" {
			signOutput = "."
		}

		// Extract base name from CSR for output filename
		csrBase := filepath.Base(signCSR)
		csrBase = strings.TrimSuffix(csrBase, ".csr")
		csrBase = strings.TrimSuffix(csrBase, ".req")

		certPath := filepath.Join(signOutput, csrBase+".crt")

		// Sign the CSR
		fmt.Println("🖊️  Signing Certificate Signing Request...")

		err := cert.SignCSR(options, certPath)
		if err != nil {
			return fmt.Errorf("failed to sign CSR: %w", err)
		}

		// Display success message
		ui.ShowSuccess("Certificate signed successfully!")
		fmt.Println()
		fmt.Println("📁 Certificate created:")
		fmt.Printf("  📜 Certificate: %s\n", certPath)
		fmt.Println()
		fmt.Println("📋 Next steps:")
		fmt.Println("  1. Deliver the signed certificate to the requester")
		fmt.Println("  2. The certificate should be used with the original private key from the CSR")
		fmt.Println("  3. Install the certificate along with the CA certificate chain")

		// Display the signed certificate details
		fmt.Println()
		fmt.Println("🔍 Signed Certificate Details:")
		signedCert, err := cert.InspectFile(certPath)
		if err != nil {
			ui.ShowInfo(fmt.Sprintf("Could not display certificate details: %v", err))
		} else {
			ui.DisplayCertificate(signedCert, false)
		}

		return nil
	},
}

func init() {
	signCmd.Flags().StringVar(&signCSR, "csr", "", "Path to the CSR file to sign (required)")
	signCmd.Flags().StringVar(&signCA, "ca", "", "Path to the CA certificate (required)")
	signCmd.Flags().StringVar(&signCAKey, "ca-key", "", "Path to the CA private key (required)")
	signCmd.Flags().IntVarP(&signDays, "days", "d", 365, "Validity period in days")
	signCmd.Flags().StringVarP(&signOutput, "output", "o", "", "Output directory for signed certificate")
	signCmd.Flags().StringSliceVar(&signSANs, "san", []string{}, "Subject Alternative Name (overrides CSR SANs if specified)")

	rootCmd.AddCommand(signCmd)
}
