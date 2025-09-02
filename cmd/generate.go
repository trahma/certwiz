package cmd

import (
    "fmt"
    "path/filepath"

    "certwiz/pkg/cert"
    "certwiz/pkg/ui"

    "github.com/spf13/cobra"
)

var (
	generateCN      string
	generateDays    int
	generateKeySize int
	generateSANs    []string
	generateOutput  string
)

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate a self-signed certificate",
	Long: `Generate a self-signed certificate with the specified parameters.

The certificate and private key will be saved in the output directory
with filenames based on the common name.

Examples:
  cert generate --cn example.com
  cert generate --cn myserver --days 730 --key-size 4096
  cert generate --cn example.com --san *.example.com --san www.example.com
  cert generate --cn server --san IP:192.168.1.100 --san localhost`,
    RunE: func(cmd *cobra.Command, args []string) error {
        if generateCN == "" {
            ui.ShowError("Common Name (--cn) is required")
            return fmt.Errorf("missing required flag: --cn")
        }

		opts := cert.GenerateOptions{
			CommonName: generateCN,
			Days:       generateDays,
			KeySize:    generateKeySize,
			SANs:       generateSANs,
			OutputDir:  generateOutput,
		}

		if !jsonOutput {
			ui.ShowInfo("Generating RSA private key...")
			ui.ShowInfo("Creating self-signed certificate...")
		}

        if err := cert.Generate(opts); err != nil {
            if jsonOutput { printJSONError(err) } else { ui.ShowError(err.Error()) }
            return err
        }

		certPath := filepath.Join(generateOutput, generateCN+".crt")
		keyPath := filepath.Join(generateOutput, generateCN+".key")

        if jsonOutput {
            printJSON(cert.JSONOperationResult{
                Success: true,
                Message: "Certificate generated successfully",
                Files:   []string{certPath, keyPath},
            })
        } else {
            ui.DisplayGenerationResult(certPath, keyPath)

			// Also display the generated certificate
			generatedCert, err := cert.InspectFile(certPath)
			if err == nil {
				ui.DisplayCertificate(generatedCert, false)
			}
        }
        return nil
    },
}

func init() {
	generateCmd.Flags().StringVar(&generateCN, "cn", "", "Common Name for the certificate (required)")
	generateCmd.Flags().IntVar(&generateDays, "days", 365, "Validity period in days")
	generateCmd.Flags().IntVar(&generateKeySize, "key-size", 2048, "RSA key size in bits")
	generateCmd.Flags().StringSliceVar(&generateSANs, "san", []string{}, "Subject Alternative Name (can be used multiple times)")
	generateCmd.Flags().StringVar(&generateOutput, "output", ".", "Output directory")

	_ = generateCmd.MarkFlagRequired("cn")
}
