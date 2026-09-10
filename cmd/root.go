package cmd

import (
	"fmt"

	"certwiz/internal/config"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var version = "0.4.1"

var (
	versionFlag bool
	jsonOutput  bool
	plainOutput bool
)

// AppConfig holds the loaded configuration
var AppConfig *config.Config

var rootCmd = &cobra.Command{
	Use:   "cert",
	Short: "A user-friendly CLI tool for certificate management",
	Long:  `cert (from certwiz) is a user-friendly CLI tool for certificate management. Similar to HTTPie but for certificates.`,
	Example: `  cert inspect cert.pem
  cert inspect google.com --chain
  cert generate --cn example.com
  cert convert cert.pem cert.der --format der
  cert verify cert.pem --host example.com
  cert csr --cn server.example.com
  cert ca --cn "My Root CA"
  cert sign --csr server.csr --ca ca.crt --ca-key ca.key
  cert tls example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag {
			fmt.Printf("cert version %s\n", version)
			return nil
		}
		// Defer to Cobra's help when no subcommand provided
		return cmd.Help()
	},
}

// Execute runs the root command. Cobra's own error printing is disabled so
// each failure is reported exactly once: commands print their own errors via
// reportError, and anything else (flag or argument errors, verification
// failures) is printed here on stderr.
func Execute() error {
	errorReported = false
	err := rootCmd.Execute()
	if err != nil && !errorReported {
		_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "Error:", err)
	}
	return err
}

func init() {
	// Usage and errors are printed by Execute and reportError, not by cobra
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&plainOutput, "plain", false, "Output in plain format (no borders, colors, or emojis)")

	// Add version flag
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")

	// Load config and apply flag overrides before any command runs
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		AppConfig = config.Load()
		if plainOutput {
			AppConfig.ApplyPlainMode()
		}
		// Pass config to UI package
		ui.SetConfig(AppConfig)
	}

	// Add subcommands
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of cert",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cert version %s\n", version)
	},
}
