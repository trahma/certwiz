package cmd

import (
	"encoding/json"
	"fmt"

	env "certwiz/internal/environ"
	"certwiz/pkg/cert"
)

// getEmoji returns an emoji or ASCII equivalent based on config and environment
func getEmoji(emoji, ascii string) string {
	// Check config first (if loaded)
	if AppConfig != nil && !AppConfig.ShouldShowEmojis() {
		return ascii
	}
	// Fall back to environment check
	if env.IsCI() {
		return ascii
	}
	return emoji
}

// printJSON pretty-prints a value as JSON
func printJSON(v interface{}) {
    data, _ := json.MarshalIndent(v, "", "  ")
    fmt.Println(string(data))
}

// printJSONError prints a standardized JSON error payload
func printJSONError(err error) {
    printJSON(cert.JSONOperationResult{Success: false, Error: err.Error()})
}
