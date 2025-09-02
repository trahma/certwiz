package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var version = "0.1.10"

var (
	versionFlag bool
	jsonOutput  bool
)

var rootCmd = &cobra.Command{
    Use:   "cert",
    Short: "A user-friendly CLI tool for certificate management",
    Long:  `cert (from certwiz) is a user-friendly CLI tool for certificate management. Similar to HTTPie but for certificates.`,
    Example: `  cert inspect cert.pem
  cert inspect google.com --chain
  cert generate --cn example.com
  cert convert cert.pem cert.der --format der
  cert verify cert.pem --host example.com`,
    RunE: func(cmd *cobra.Command, args []string) error {
        if versionFlag {
            fmt.Printf("cert version %s\n", version)
            return nil
        }
        // Defer to Cobra's help when no subcommand provided
        return cmd.Help()
    },
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
    // Prefer Cobra-managed help/errors
    rootCmd.SilenceUsage = true
    rootCmd.SilenceErrors = false

    // Add global flags
    rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// Add version flag
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")

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
