package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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
	inspectTimeout time.Duration
	inspectSigAlg  string
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [file|url]",
	Short: "Inspect a certificate from a file or URL",
	Long: `Inspect a certificate from a file, URL, or stdin and display its information.

If the argument is an existing file, it will read and parse the certificate file.
Files containing multiple certificates (e.g. fullchain.pem) are supported; use
--chain to display all of them. Use "-" to read from stdin.
If the argument looks like a URL or domain name, it will connect to the remote
server and retrieve its certificate. Arguments that look like file paths (a
directory separator, a leading "." or "~", or a certificate extension such as
.pem or .crt) are always treated as files, so a typo in a path is reported as
a missing file rather than a failed connection.

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
				reportError(cmd, err)
				return err
			}

			certs, err := cert.InspectData(data, "stdin")
			if err != nil {
				reportError(cmd, err)
				return err
			}

			displayLocalCertificates(certs)
			return nil
		}

		// A readable path is a file (possibly a bundle with multiple certificates)
		if _, err := os.Stat(target); err == nil {
			certs, err := cert.InspectFileAll(target)
			if err != nil {
				reportError(cmd, err)
				return err
			}

			displayLocalCertificates(certs)
			return nil
		} else if looksLikeFilePath(target) {
			// Do not fall through to a hostname lookup for an obvious path typo
			if os.IsNotExist(err) {
				err = fmt.Errorf("certificate file does not exist: %s", target)
			} else {
				err = fmt.Errorf("cannot access certificate file: %w", err)
			}
			reportError(cmd, err)
			return err
		}

		// Otherwise treat the target as a URL/hostname
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

		certificate, chain, err := cert.InspectURLWithOptions(target, port, connectHost, inspectTimeout, inspectSigAlg)
		if err != nil {
			reportError(cmd, err)
			return err
		}

		if jsonOutput {
			jsonCert := certificate.ToJSON()
			if inspectChain && len(chain) > 0 {
				jsonCert.Chain = chainSummaries(chain)
			}
			printJSON(jsonCert)
			return nil
		}

		ui.DisplayCertificate(certificate, inspectFull)
		if inspectChain && len(chain) > 0 {
			ui.DisplayCertificateChain(chain)
		}
		return nil
	},
}

// certFileExtensions lists file extensions that identify a target as a
// certificate or key file rather than a hostname.
var certFileExtensions = map[string]bool{
	".pem": true, ".crt": true, ".cer": true, ".der": true, ".key": true,
	".p7b": true, ".p7c": true, ".pfx": true, ".p12": true, ".csr": true,
}

// looksLikeFilePath reports whether a non-existent inspect target should be
// treated as a file path rather than a hostname. It is true when the target:
//   - starts with "." or "~", or a Windows drive prefix such as "C:";
//   - ends with a certificate-ish extension (.pem, .crt, .key, ...); or
//   - contains a "/" or "\" whose first segment is not a hostname. A first
//     segment containing a dot or a colon (e.g. "example.com/path" or
//     "host:8443/x") is taken to be a host, since URLs without a scheme are
//     accepted and get "https://" prepended.
//
// Anything containing "://" is a URL and never a file path.
func looksLikeFilePath(target string) bool {
	if target == "" || strings.Contains(target, "://") {
		return false
	}

	if strings.HasPrefix(target, ".") || strings.HasPrefix(target, "~") {
		return true
	}

	// Windows drive prefix: a letter, a colon, then a separator
	if len(target) >= 3 && target[1] == ':' && (target[2] == '\\' || target[2] == '/') {
		c := target[0]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}

	if certFileExtensions[strings.ToLower(filepath.Ext(target))] {
		return true
	}

	if i := strings.IndexAny(target, "/\\"); i >= 0 {
		first := target[:i]
		if first == "" {
			return true // absolute path
		}
		return !strings.ContainsAny(first, ".:")
	}

	return false
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
	inspectCmd.Flags().DurationVar(&inspectTimeout, "timeout", 5*time.Second, "Network timeout for remote inspection (e.g., 5s, 2s)")
	inspectCmd.Flags().StringVar(&inspectSigAlg, "sig-alg", "auto", "Preferred signature algorithm: auto, ecdsa, or rsa (TLS 1.2 only)")
}
