package cmd

import (
    "fmt"
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
  cert verify cert.pem
  cert verify server.crt --host example.com
  cert verify cert.pem --ca ca.pem --host myserver.local`,
	Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        certPath := args[0]

		// Check if certificate file exists
        if _, err := os.Stat(certPath); os.IsNotExist(err) {
            ui.ShowError("Certificate file does not exist: " + certPath)
            return fmt.Errorf("certificate file does not exist: %s", certPath)
        }

		if !jsonOutput {
			ui.ShowInfo("Verifying certificate...")
		}

        result, err := cert.Verify(certPath, verifyCA, verifyHost)
        if err != nil {
            if jsonOutput {
                printJSONError(err)
            } else {
                ui.ShowError(err.Error())
            }
            return err
        }

        if jsonOutput {
            printJSON(result.ToJSON())
        } else {
            ui.DisplayVerificationResult(result)
        }

        // Surface failure as an error to drive non-zero exit via main
        if !result.IsValid {
            return fmt.Errorf("verification failed")
        }
        return nil
    },
}

func init() {
	verifyCmd.Flags().StringVar(&verifyCA, "ca", "", "CA certificate file for chain verification")
	verifyCmd.Flags().StringVar(&verifyHost, "host", "", "Hostname to verify against the certificate")
}
