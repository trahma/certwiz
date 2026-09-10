package environ

import "testing"

// mapEnv returns a getenv function backed by a fixed map.
func mapEnv(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestDetectCI(t *testing.T) {
	for _, name := range ciVariables {
		t.Run(name, func(t *testing.T) {
			if !detectCI(mapEnv(map[string]string{name: "1"})) {
				t.Errorf("detectCI should be true when %s is set", name)
			}
		})
	}

	t.Run("NoneSet", func(t *testing.T) {
		if detectCI(mapEnv(map[string]string{"HOME": "/tmp", "TERM": "xterm"})) {
			t.Error("detectCI should be false when no CI variable is set")
		}
	})

	t.Run("EmptyValueDoesNotCount", func(t *testing.T) {
		if detectCI(mapEnv(map[string]string{"CI": ""})) {
			t.Error("detectCI should be false when the CI variable is empty")
		}
	})
}

func TestDetectUnicode(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"DumbTerm", map[string]string{"TERM": "dumb", "LANG": "en_US.UTF-8"}, false},
		{"EmptyTerm", map[string]string{"TERM": "", "LANG": "en_US.UTF-8"}, false},
		{"LangWithoutUTF", map[string]string{"TERM": "xterm", "LANG": "C"}, false},
		{"LangWithUTF8", map[string]string{"TERM": "xterm-256color", "LANG": "en_US.UTF-8"}, true},
		{"LangLowercaseUTF", map[string]string{"TERM": "xterm", "LANG": "en_GB.utf8"}, true},
		{"LCAllFallbackUTF", map[string]string{"TERM": "xterm", "LANG": "", "LC_ALL": "C.UTF-8"}, true},
		{"LCAllFallbackNonUTF", map[string]string{"TERM": "xterm", "LANG": "", "LC_ALL": "POSIX"}, false},
		{"LangTakesPrecedenceOverLCAll", map[string]string{"TERM": "xterm", "LANG": "C", "LC_ALL": "en_US.UTF-8"}, false},
		{"NoLangOrLCAll", map[string]string{"TERM": "xterm"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectUnicode(mapEnv(tt.env)); got != tt.want {
				t.Errorf("detectUnicode(%v) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}

func TestCachedResultsAreStable(t *testing.T) {
	ci := IsCI()
	uni := SupportsUnicode()

	// The cached values must not change across calls, even if the environment
	// changes after the first evaluation.
	t.Setenv("CI", "1")
	t.Setenv("TERM", "dumb")

	for i := 0; i < 3; i++ {
		if got := IsCI(); got != ci {
			t.Fatalf("IsCI changed between calls: first %v, then %v", ci, got)
		}
		if got := SupportsUnicode(); got != uni {
			t.Fatalf("SupportsUnicode changed between calls: first %v, then %v", uni, got)
		}
	}
}
