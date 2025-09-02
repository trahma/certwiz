package cmd

import (
    "os"
    "strconv"
    "strings"
    "time"

    "certwiz/pkg/cert"
    "certwiz/pkg/ui"

    "github.com/spf13/cobra"
)

var (
    inspectFull    bool
    inspectPort    int
    inspectChain   bool
    inspectConnect string
    inspectTimeout string
    inspectSigAlg  string
)

var inspectCmd = &cobra.Command{
    Use:   "inspect [file|url]",
    Short: "Inspect a certificate from a file or URL",
	Long: `Inspect a certificate from a file or URL and display its information.

If the argument is a valid file path, it will read and parse the certificate file.
If the argument looks like a URL or domain name, it will connect to the remote
server and retrieve its certificate.

Examples:
  cert inspect cert.pem
  cert inspect cert.der --full  
  cert inspect google.com
  cert inspect https://example.com:8443 --port 8443
  cert inspect 192.168.1.1:443
  cert inspect google.com --connect localhost:8080
  cert inspect api.example.com --connect tunnel.local --port 443
  cert inspect cloudflare.com --sig-alg ecdsa
  cert inspect cloudflare.com --sig-alg rsa`,
	Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        target := args[0]

		// Determine if target is a file or URL
		if _, err := os.Stat(target); err == nil {
			// It's a file
            certificate, err := cert.InspectFile(target)
            if err != nil {
                if jsonOutput {
                    printJSONError(err)
                } else {
                    ui.ShowError(err.Error())
                }
                return err
            }

            if jsonOutput {
                printJSON(certificate.ToJSON())
            } else {
                ui.DisplayCertificate(certificate, inspectFull)
            }
        } else {
			// It's a URL/hostname
			port := inspectPort
			connectHost := ""

			// Extract port from target if specified
			if strings.Contains(target, ":") && !strings.HasPrefix(target, "http") {
				parts := strings.Split(target, ":")
				if len(parts) == 2 {
					if p, err := strconv.Atoi(parts[1]); err == nil {
						target = parts[0]
						port = p
					}
				}
			}

			// Handle --connect flag
			if inspectConnect != "" {
				connectHost = inspectConnect
				// Check if connect has a port specified
				if strings.Contains(connectHost, ":") {
					parts := strings.Split(connectHost, ":")
					if len(parts) == 2 {
						if p, err := strconv.Atoi(parts[1]); err == nil {
							connectHost = parts[0]
							port = p // Override port with the one from --connect
						}
					}
				}
			}

			// Determine timeout
            timeout := 5 * time.Second
            if inspectTimeout != "" {
                if d, err := time.ParseDuration(inspectTimeout); err == nil {
                    timeout = d
                }
            }

			// Use the enhanced function that supports connect host, timeout, and signature algorithm preference
            certificate, chain, err := cert.InspectURLWithOptions(target, port, connectHost, timeout, inspectSigAlg)
            if err != nil {
                if jsonOutput {
                    printJSONError(err)
                } else {
                    ui.ShowError(err.Error())
                }
                return err
            }

            if jsonOutput {
                jsonCert := certificate.ToJSON()

				// Add chain if requested
				if inspectChain && len(chain) > 0 {
					for _, c := range chain {
						jsonCert.Chain = append(jsonCert.Chain, cert.JSONCertSummary{
							Subject:      c.Subject.String(),
							Issuer:       c.Issuer.String(),
							NotBefore:    c.NotBefore,
							NotAfter:     c.NotAfter,
							IsExpired:    c.IsExpired,
							SerialNumber: c.SerialNumber.Text(16),
						})
					}
				}

                printJSON(jsonCert)
            } else {
                ui.DisplayCertificate(certificate, inspectFull)

                // Display chain if requested
                if inspectChain && len(chain) > 0 {
                    ui.DisplayCertificateChain(chain)
                }
            }
        }
        return nil
    },
}

func init() {
    inspectCmd.Flags().BoolVar(&inspectFull, "full", false, "Show full certificate details including extensions")
    inspectCmd.Flags().IntVar(&inspectPort, "port", 443, "Port for remote inspection")
    inspectCmd.Flags().BoolVar(&inspectChain, "chain", false, "Show certificate chain")
    inspectCmd.Flags().StringVar(&inspectConnect, "connect", "", "Connect to a different host (e.g., localhost:8080) while validating the cert for the target hostname")
    inspectCmd.Flags().StringVar(&inspectTimeout, "timeout", "5s", "Network timeout for remote inspection (e.g., 5s, 2s)")
    inspectCmd.Flags().StringVar(&inspectSigAlg, "sig-alg", "auto", "Preferred signature algorithm: auto, ecdsa, or rsa (TLS 1.2 only)")
}
