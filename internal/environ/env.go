package environ

import (
	"os"
	"strings"
	"sync"
)

// ciVariables are environment variables whose presence indicates a CI environment.
var ciVariables = []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS", "CIRCLECI"}

var (
	ciOnce    sync.Once
	ciResult  bool
	uniOnce   sync.Once
	uniResult bool
)

// IsCI reports whether the process is running in a CI environment.
// The result is computed once per process.
func IsCI() bool {
	ciOnce.Do(func() {
		ciResult = detectCI(os.Getenv)
	})
	return ciResult
}

// SupportsUnicode heuristically determines if the terminal supports Unicode.
// The result is computed once per process.
func SupportsUnicode() bool {
	uniOnce.Do(func() {
		uniResult = detectUnicode(os.Getenv)
	})
	return uniResult
}

// detectCI reports whether any known CI variable is set, using getenv for lookups.
func detectCI(getenv func(string) string) bool {
	for _, v := range ciVariables {
		if getenv(v) != "" {
			return true
		}
	}
	return false
}

// detectUnicode is a TERM/locale heuristic; getenv is injected so tests do
// not depend on the process environment.
func detectUnicode(getenv func(string) string) bool {
	term := getenv("TERM")
	if term == "dumb" || term == "" {
		return false
	}
	lang := getenv("LANG")
	if lang == "" {
		lang = getenv("LC_ALL")
	}
	if lang != "" && !strings.Contains(strings.ToLower(lang), "utf") {
		return false
	}
	return true
}
