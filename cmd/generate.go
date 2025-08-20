package cmd

import (
	"os"
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
  certwiz generate --cn example.com
  certwiz generate --cn myserver --days 730 --key-size 4096
  certwiz generate --cn example.com --san *.example.com --san www.example.com
  certwiz generate --cn server --san IP:192.168.1.100 --san localhost`,
	Run: func(cmd *cobra.Command, args []string) {
		if generateCN == "" {
			ui.ShowError("Common Name (--cn) is required")
			os.Exit(1)
		}

		opts := cert.GenerateOptions{
			CommonName: generateCN,
			Days:       generateDays,
			KeySize:    generateKeySize,
			SANs:       generateSANs,
			OutputDir:  generateOutput,
		}

		ui.ShowInfo("Generating RSA private key...")
		ui.ShowInfo("Creating self-signed certificate...")

		err := cert.Generate(opts)
		if err != nil {
			ui.ShowError(err.Error())
			os.Exit(1)
		}

		certPath := filepath.Join(generateOutput, generateCN+".crt")
		keyPath := filepath.Join(generateOutput, generateCN+".key")

		ui.DisplayGenerationResult(certPath, keyPath)

		// Also display the generated certificate
		generatedCert, err := cert.InspectFile(certPath)
		if err == nil {
			ui.DisplayCertificate(generatedCert, false)
		}
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
