package cmd

import (
	"strconv"
	"strings"
	"time"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	tlsPort    int
	tlsTimeout string
)

var tlsCmd = &cobra.Command{
	Use:   "tls [hostname]",
	Short: "Test supported TLS versions for a hostname",
	Long: `Test which TLS versions are supported by a remote server.

This command attempts to connect to the specified hostname using each
TLS version (1.0, 1.1, 1.2, and 1.3) and reports which versions are
supported by the server.

Examples:
  cert tls google.com
  cert tls example.com:443
  cert tls 192.168.1.1 --port 443
  cert tls localhost --timeout 2s`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		// Extract port from target if specified
		port := tlsPort
		host := target
		if strings.Contains(target, ":") && !strings.HasPrefix(target, "http") {
			parts := strings.Split(target, ":")
			if len(parts) == 2 {
				if p, err := strconv.Atoi(parts[1]); err == nil {
					host = parts[0]
					port = p
				}
			}
		}

		// Determine timeout
		timeout := 5 * time.Second
		if tlsTimeout != "" {
			if d, err := time.ParseDuration(tlsTimeout); err == nil {
				timeout = d
			}
		}

		// Test TLS versions
		result, err := cert.CheckTLSVersions(host, port, timeout)
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
			ui.DisplayTLSVersionResults(result)
		}

		return nil
	},
}

func init() {
	tlsCmd.Flags().IntVar(&tlsPort, "port", 443, "Port for TLS testing")
	tlsCmd.Flags().StringVar(&tlsTimeout, "timeout", "5s", "Network timeout (e.g., 5s, 2s)")

	rootCmd.AddCommand(tlsCmd)
}
