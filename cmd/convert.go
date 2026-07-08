package cmd

import (
    "fmt"
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
  cert convert cert.pem cert.der --format der
  cert convert cert.der cert.pem --format pem
  cert convert server.crt server.der --format der`,
	Args: cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        inputPath := args[0]
        outputPath := args[1]

		// Check if input file exists
        if _, err := os.Stat(inputPath); os.IsNotExist(err) {
            err := fmt.Errorf("input file does not exist: %s", inputPath)
            if jsonOutput {
                printJSONError(err)
            } else {
                ui.ShowError(err.Error())
            }
            return err
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

		if !jsonOutput {
			ui.ShowInfo("Converting certificate format...")
		}

        if err := cert.Convert(inputPath, outputPath, convertFormat); err != nil {
            if jsonOutput {
                printJSONError(err)
            } else {
                ui.ShowError(err.Error())
            }
            return err
        }

        if jsonOutput {
            printJSON(cert.JSONOperationResult{
                Success: true,
                Message: fmt.Sprintf("Converted from %s to %s", strings.ToUpper(inputFormat), strings.ToUpper(convertFormat)),
                Files:   []string{outputPath},
            })
        } else {
            ui.DisplayConversionResult(inputPath, outputPath, inputFormat, convertFormat)
        }
        return nil
    },
}

func init() {
	convertCmd.Flags().StringVar(&convertFormat, "format", "pem", "Output format (pem or der)")
}
