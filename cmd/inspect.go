package cmd

import (
    "fmt"
    "io"
    "net"
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
	Long: `Inspect a certificate from a file, URL, or stdin and display its information.

If the argument is a valid file path, it will read and parse the certificate file.
Files containing multiple certificates (e.g. fullchain.pem) are supported; use
--chain to display all of them. Use "-" to read from stdin.
If the argument looks like a URL or domain name, it will connect to the remote
server and retrieve its certificate.

Examples:
  cert inspect cert.pem
  cert inspect cert.der --full
  cert inspect fullchain.pem --chain
  openssl s_client -connect example.com:443 </dev/null | cert inspect -
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

		// Read from stdin when the target is "-"
		if target == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				err = fmt.Errorf("failed to read from stdin: %w", err)
				if jsonOutput {
					printJSONError(err)
				} else {
					ui.ShowError(err.Error())
				}
				return err
			}

			certs, err := cert.InspectData(data, "stdin")
			if err != nil {
				if jsonOutput {
					printJSONError(err)
				} else {
					ui.ShowError(err.Error())
				}
				return err
			}

			displayLocalCertificates(certs)
			return nil
		}

		// Determine if target is a file or URL
		if _, err := os.Stat(target); err == nil {
			// It's a file (possibly a bundle with multiple certificates)
            certs, err := cert.InspectFileAll(target)
            if err != nil {
                if jsonOutput {
                    printJSONError(err)
                } else {
                    ui.ShowError(err.Error())
                }
                return err
            }

            displayLocalCertificates(certs)
        } else {
			// It's a URL/hostname
			port := inspectPort
			connectHost := ""

			// Extract port from target if specified (URLs with a scheme are
			// parsed later; handles IPv6 like [::1]:443)
			if !strings.Contains(target, "://") {
				if h, p, err := net.SplitHostPort(target); err == nil {
					if pn, err := strconv.Atoi(p); err == nil {
						target = h
						port = pn
					}
				}
			}

			// Handle --connect flag
			if inspectConnect != "" {
				connectHost = inspectConnect
				// Check if connect has a port specified
				if h, p, err := net.SplitHostPort(connectHost); err == nil {
					if pn, err := strconv.Atoi(p); err == nil {
						connectHost = h
						port = pn // Override port with the one from --connect
					}
				}
			}

			// Determine timeout
            timeout := 5 * time.Second
            if inspectTimeout != "" {
                d, err := time.ParseDuration(inspectTimeout)
                if err != nil {
                    return fmt.Errorf("invalid --timeout value %q: %w", inspectTimeout, err)
                }
                timeout = d
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
					jsonCert.Chain = chainSummaries(chain)
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

// chainSummaries converts chain certificates to their JSON summary form
func chainSummaries(chain []*cert.Certificate) []cert.JSONCertSummary {
	summaries := make([]cert.JSONCertSummary, 0, len(chain))
	for _, c := range chain {
		summaries = append(summaries, cert.JSONCertSummary{
			Subject:      c.Subject.String(),
			Issuer:       c.Issuer.String(),
			NotBefore:    c.NotBefore,
			NotAfter:     c.NotAfter,
			IsExpired:    c.IsExpired,
			SerialNumber: c.SerialNumber.Text(16),
		})
	}
	return summaries
}

// displayLocalCertificates renders certificates parsed from a file or stdin.
// The first certificate is shown in full; any additional bundle certificates
// are shown with --chain or surfaced via a hint so they aren't silently hidden.
func displayLocalCertificates(certs []*cert.Certificate) {
	certificate := certs[0]
	rest := certs[1:]

	if jsonOutput {
		jsonCert := certificate.ToJSON()
		if inspectChain && len(rest) > 0 {
			jsonCert.Chain = chainSummaries(rest)
		}
		printJSON(jsonCert)
		return
	}

	ui.DisplayCertificate(certificate, inspectFull)

	if len(rest) == 0 {
		return
	}
	if inspectChain {
		ui.DisplayCertificateChain(rest)
	} else {
		fmt.Println()
		ui.ShowInfo(fmt.Sprintf("Contains %d certificates; showing the first. Use --chain to see the rest.", len(certs)))
	}
}

func init() {
    inspectCmd.Flags().BoolVar(&inspectFull, "full", false, "Show full certificate details including extensions")
    inspectCmd.Flags().IntVar(&inspectPort, "port", 443, "Port for remote inspection")
    inspectCmd.Flags().BoolVar(&inspectChain, "chain", false, "Show certificate chain")
    inspectCmd.Flags().StringVar(&inspectConnect, "connect", "", "Connect to a different host (e.g., localhost:8080) while validating the cert for the target hostname")
    inspectCmd.Flags().StringVar(&inspectTimeout, "timeout", "5s", "Network timeout for remote inspection (e.g., 5s, 2s)")
    inspectCmd.Flags().StringVar(&inspectSigAlg, "sig-alg", "auto", "Preferred signature algorithm: auto, ecdsa, or rsa (TLS 1.2 only)")
}
