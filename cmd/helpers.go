package cmd

import "os"

// isCI checks if we're running in a CI environment
func isCI() bool {
	// Check common CI environment variables
	ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS", "CIRCLECI"}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

// getEmoji returns an emoji or ASCII equivalent based on environment
func getEmoji(emoji, ascii string) string {
	if isCI() {
		return ascii
	}
	return emoji
}
