package cmd

import (
	"fmt"
	"net"
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

		// Strip URL scheme and path if present (e.g. https://example.com/path)
		port := tlsPort
		host := target
		if i := strings.Index(host, "://"); i != -1 {
			host = host[i+3:]
		}
		if i := strings.Index(host, "/"); i != -1 {
			host = host[:i]
		}

		// Extract port from target if specified (handles IPv6 like [::1]:443)
		if h, p, err := net.SplitHostPort(host); err == nil {
			if pn, err := strconv.Atoi(p); err == nil {
				host = h
				port = pn
			}
		}

		// Determine timeout
		timeout := 5 * time.Second
		if tlsTimeout != "" {
			d, err := time.ParseDuration(tlsTimeout)
			if err != nil {
				return fmt.Errorf("invalid --timeout value %q: %w", tlsTimeout, err)
			}
			timeout = d
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
