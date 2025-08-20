package cmd

import (
	"os"
	"strconv"
	"strings"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	inspectFull  bool
	inspectPort  int
	inspectChain bool
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
  cert inspect 192.168.1.1:443`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]

		// Determine if target is a file or URL
		if _, err := os.Stat(target); err == nil {
			// It's a file
			cert, err := cert.InspectFile(target)
			if err != nil {
				ui.ShowError(err.Error())
				os.Exit(1)
			}
			ui.DisplayCertificate(cert, inspectFull)
		} else {
			// It's a URL/hostname
			port := inspectPort

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

			cert, chain, err := cert.InspectURLWithChain(target, port)
			if err != nil {
				ui.ShowError(err.Error())
				os.Exit(1)
			}
			ui.DisplayCertificate(cert, inspectFull)

			// Display chain if requested
			if inspectChain && len(chain) > 0 {
				ui.DisplayCertificateChain(chain)
			}
		}
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectFull, "full", false, "Show full certificate details including extensions")
	inspectCmd.Flags().IntVar(&inspectPort, "port", 443, "Port for remote inspection")
	inspectCmd.Flags().BoolVar(&inspectChain, "chain", false, "Show certificate chain")
}
