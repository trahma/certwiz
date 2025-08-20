package cmd

import (
	"os"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	verifyCA   string
	verifyHost string
)

var verifyCmd = &cobra.Command{
	Use:   "verify [certificate]",
	Short: "Verify a certificate",
	Long: `Verify a certificate's validity, expiration, and optionally check
hostname matching and CA chain validation.

Examples:
  certwiz verify cert.pem
  certwiz verify server.crt --host example.com
  certwiz verify cert.pem --ca ca.pem --host myserver.local`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certPath := args[0]

		// Check if certificate file exists
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			ui.ShowError("Certificate file does not exist: " + certPath)
			os.Exit(1)
		}

		ui.ShowInfo("Verifying certificate...")

		result, err := cert.Verify(certPath, verifyCA, verifyHost)
		if err != nil {
			ui.ShowError(err.Error())
			os.Exit(1)
		}

		ui.DisplayVerificationResult(result)

		// Exit with error code if verification failed
		if !result.IsValid {
			os.Exit(1)
		}
	},
}

func init() {
	verifyCmd.Flags().StringVar(&verifyCA, "ca", "", "CA certificate file for chain verification")
	verifyCmd.Flags().StringVar(&verifyHost, "host", "", "Hostname to verify against the certificate")
}
