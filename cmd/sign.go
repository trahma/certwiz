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
		var validationErr error
		switch {
		case signCSR == "":
			validationErr = fmt.Errorf("CSR file (--csr) is required")
		case signCA == "":
			validationErr = fmt.Errorf("CA certificate (--ca) is required")
		case signCAKey == "":
			validationErr = fmt.Errorf("CA private key (--ca-key) is required")
		}
		if validationErr != nil {
			if jsonOutput {
				printJSONError(validationErr)
			}
			return validationErr
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
		if !jsonOutput {
			fmt.Printf("%s Signing Certificate Signing Request...\n", getEmoji("🖊️", "[SIGN]"))
		}

		err := cert.SignCSR(options, certPath)
		if err != nil {
			err = fmt.Errorf("failed to sign CSR: %w", err)
			if jsonOutput {
				printJSONError(err)
			}
			return err
		}

		if jsonOutput {
			printJSON(cert.JSONOperationResult{
				Success: true,
				Message: "Certificate signed successfully",
				Files:   []string{certPath},
			})
			return nil
		}

		// Display success message
		ui.ShowSuccess("Certificate signed successfully!")
		fmt.Println()
		fmt.Printf("%s Certificate created:\n", getEmoji("📁", "[FILES]"))
		fmt.Printf("  %s Certificate: %s\n", getEmoji("📜", "[CERT]"), certPath)
		fmt.Println()
		fmt.Printf("%s Next steps:\n", getEmoji("📋", "[NEXT]"))
		fmt.Println("  1. Deliver the signed certificate to the requester")
		fmt.Println("  2. The certificate should be used with the original private key from the CSR")
		fmt.Println("  3. Install the certificate along with the CA certificate chain")

		// Display the signed certificate details
		fmt.Println()
		fmt.Printf("%s Signed Certificate Details:\n", getEmoji("🔍", "[INFO]"))
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
