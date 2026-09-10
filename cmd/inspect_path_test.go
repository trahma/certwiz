package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// useErrBuffer routes the inspect command's error writer into a buffer for
// the test. reportError writes through cmd.ErrOrStderr(), which other tests
// in this package may have pointed at their own writers on the root command.
func useErrBuffer(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	inspectCmd.SetErr(&buf)
	t.Cleanup(func() { inspectCmd.SetErr(nil) })
	return &buf
}

func TestLooksLikeFilePath(t *testing.T) {
	cases := []struct {
		target string
		want   bool
	}{
		// File paths
		{"cert.pem", true},
		{"./cert.pem", true},
		{"../x/cert.pem", true},
		{"/etc/ssl/cert.pem", true},
		{"/nonexistent", true},
		{`C:\certs\cert.pem`, true},
		{"C:/certs/cert.pem", true},
		{"server.key", true},
		{"bundle.P7B", true},
		{"missing.crt", true},
		{"certs/server", true},
		{"~/certs/server.crt", true},
		{"~", true},

		// Hosts and URLs
		{"", false},
		{"GOOGLE.COM", false},
		{"google.com", false},
		{"google.com:443", false},
		{"example.com:8443", false},
		{"https://google.com", false},
		{"https://google.com/x.pem", false},
		{"https://example.com/path", false},
		{"example.com/path", false},
		{"host:8443/x", false},
		{"[::1]:443", false},
		{"192.168.1.1:443", false},
		{"localhost", false},
		{"c:443", false},
	}

	for _, tc := range cases {
		if got := looksLikeFilePath(tc.target); got != tc.want {
			t.Errorf("looksLikeFilePath(%q) = %v, want %v", tc.target, got, tc.want)
		}
	}
}

func TestInspectMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.pem")

	t.Run("Plain", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, false, true)
		errBuf := useErrBuffer(t)

		var err error
		stdout := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{missing})
		})
		if err == nil {
			t.Fatal("expected an error for a missing file")
		}
		for _, s := range []string{err.Error(), errBuf.String()} {
			if !strings.Contains(s, "certificate file does not exist") || !strings.Contains(s, missing) {
				t.Errorf("expected a missing-file message naming %q, got %q", missing, s)
			}
			if strings.Contains(s, "dial") || strings.Contains(s, "lookup") {
				t.Errorf("missing file was treated as a hostname: %q", s)
			}
		}
		if stdout != "" {
			t.Errorf("expected empty stdout, got %q", stdout)
		}
	})

	t.Run("JSON", func(t *testing.T) {
		resetInspectFlags(t)
		setOutputMode(t, true, false)
		errBuf := useErrBuffer(t)

		var err error
		stdout := captureStdout(t, func() {
			err = inspectCmd.RunE(inspectCmd, []string{missing})
		})
		if err == nil {
			t.Fatal("expected an error for a missing file")
		}
		if errBuf.Len() != 0 {
			t.Errorf("expected nothing on the error stream in JSON mode, got %q", errBuf.String())
		}
		var payload struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		if jerr := json.Unmarshal([]byte(stdout), &payload); jerr != nil {
			t.Fatalf("stdout is not JSON: %v\n%s", jerr, stdout)
		}
		if payload.Success {
			t.Error("expected success=false")
		}
		if !strings.Contains(payload.Error, "certificate file does not exist") || !strings.Contains(payload.Error, missing) {
			t.Errorf("unexpected JSON error %q", payload.Error)
		}
		if strings.Contains(payload.Error, "dial") || strings.Contains(payload.Error, "lookup") {
			t.Errorf("missing file was treated as a hostname: %q", payload.Error)
		}
	})
}
