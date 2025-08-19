package cmd

import (
	"os"
	"strings"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	convertFormat string
)

var convertCmd = &cobra.Command{
	Use:   "convert [input] [output]",
	Short: "Convert certificate between formats",
	Long: `Convert a certificate file between PEM and DER formats.

The input format is automatically detected. The output format is specified
using the --format flag.

Examples:
  certwiz convert cert.pem cert.der --format der
  certwiz convert cert.der cert.pem --format pem
  certwiz convert server.crt server.der --format der`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]
		outputPath := args[1]

		// Check if input file exists
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			ui.ShowError("Input file does not exist: " + inputPath)
			os.Exit(1)
		}

		// Detect input format for display purposes
		var inputFormat string
		if data, err := os.ReadFile(inputPath); err == nil {
			if strings.Contains(string(data), "-----BEGIN CERTIFICATE-----") {
				inputFormat = "pem"
			} else {
				inputFormat = "der"
			}
		} else {
			inputFormat = "unknown"
		}

		ui.ShowInfo("Converting certificate format...")

		err := cert.Convert(inputPath, outputPath, convertFormat)
		if err != nil {
			ui.ShowError(err.Error())
			os.Exit(1)
		}

		ui.DisplayConversionResult(inputPath, outputPath, inputFormat, convertFormat)
	},
}

func init() {
	convertCmd.Flags().StringVar(&convertFormat, "format", "pem", "Output format (pem or der)")
}