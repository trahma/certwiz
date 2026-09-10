package cmd

import (
	"encoding/json"
	"fmt"

	"certwiz/pkg/cert"
	"certwiz/pkg/ui"

	"github.com/spf13/cobra"
)

// printJSON pretty-prints a value as JSON
func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

// printJSONError prints a standardized JSON error payload
func printJSONError(err error) {
	printJSON(cert.JSONOperationResult{Success: false, Error: err.Error()})
}

// errorReported records that reportError already printed the failure, so
// Execute does not print it a second time.
var errorReported bool

// reportError prints an error once: as JSON on stdout, or styled on stderr.
func reportError(cmd *cobra.Command, err error) {
	errorReported = true
	if jsonOutput {
		printJSONError(err)
		return
	}
	ui.ShowErrorTo(cmd.ErrOrStderr(), err.Error())
}
