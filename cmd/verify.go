package cmd

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "certwiz/pkg/cert"
    "certwiz/pkg/ui"

    "github.com/spf13/cobra"
)

var (
	verifyCA        string
	verifyHost      string
	verifyKey       string
	verifyExpiresIn string
)

var verifyCmd = &cobra.Command{
    Use:   "verify [certificate]",
    Short: "Verify a certificate",
	Long: `Verify a certificate's validity, expiration, and optionally check
hostname matching, CA chain validation, private key matching, and
upcoming expiry.

Examples:
  cert verify cert.pem
  cert verify server.crt --host example.com
  cert verify cert.pem --ca ca.pem --host myserver.local
  cert verify server.crt --key server.key
  cert verify cert.pem --expires-in 30d`,
	Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        certPath := args[0]

		// Check if certificate file exists
        if _, err := os.Stat(certPath); os.IsNotExist(err) {
            err := fmt.Errorf("certificate file does not exist: %s", certPath)
            if jsonOutput {
                printJSONError(err)
            } else {
                ui.ShowError(err.Error())
            }
            return err
        }

        expiresIn, err := parseExpiryWindow(verifyExpiresIn)
        if err != nil {
            if jsonOutput {
                printJSONError(err)
            } else {
                ui.ShowError(err.Error())
            }
            return err
        }

		if !jsonOutput {
			ui.ShowInfo("Verifying certificate...")
		}

        result, err := cert.VerifyWithOptions(cert.VerifyOptions{
            CertPath:  certPath,
            CAPath:    verifyCA,
            Hostname:  verifyHost,
            KeyPath:   verifyKey,
            ExpiresIn: expiresIn,
        })
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

// parseExpiryWindow parses an expiry threshold like "30d", "30" (days),
// or any Go duration such as "720h". An empty string means no threshold.
func parseExpiryWindow(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	days := ""
	if strings.HasSuffix(s, "d") {
		days = strings.TrimSuffix(s, "d")
	} else if _, err := strconv.Atoi(s); err == nil {
		days = s
	}

	if days != "" {
		n, err := strconv.Atoi(days)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid --expires-in value %q (use e.g. 30d, 30, or 720h)", s)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil || d < 0 {
		return 0, fmt.Errorf("invalid --expires-in value %q (use e.g. 30d, 30, or 720h)", s)
	}
	return d, nil
}

func init() {
	verifyCmd.Flags().StringVar(&verifyCA, "ca", "", "CA certificate file for chain verification")
	verifyCmd.Flags().StringVar(&verifyHost, "host", "", "Hostname to verify against the certificate")
	verifyCmd.Flags().StringVar(&verifyKey, "key", "", "Private key file to check against the certificate")
	verifyCmd.Flags().StringVar(&verifyExpiresIn, "expires-in", "", "Fail if the certificate expires within this window (e.g. 30d, 720h)")
}
