package environ

import (
  "os"
  "strings"
)

// IsCI reports whether the process is running in a CI environment.
func IsCI() bool {
  ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS", "CIRCLECI"}
  for _, v := range ciVars {
    if os.Getenv(v) != "" {
      return true
    }
  }
  return false
}

// SupportsUnicode heuristically determines if the terminal supports Unicode.
func SupportsUnicode() bool {
  term := os.Getenv("TERM")
  if term == "dumb" || term == "" {
    return false
  }
  lang := os.Getenv("LANG")
  if lang == "" {
    lang = os.Getenv("LC_ALL")
  }
  if lang != "" && !strings.Contains(strings.ToLower(lang), "utf") {
    return false
  }
  return true
}

