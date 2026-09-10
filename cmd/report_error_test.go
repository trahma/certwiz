package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"certwiz/internal/testutil"
	"certwiz/pkg/cert"
)

// captureStreams runs fn and returns everything written to os.Stdout and
// os.Stderr, so tests can prove which stream a message landed on.
func captureStreams(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout, os.Stderr = outW, errW

	read := func(r io.Reader) chan string {
		ch := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			ch <- buf.String()
		}()
		return ch
	}
	outCh, errCh := read(outR), read(errR)

	fn()

	_ = outW.Close()
	_ = errW.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	return <-outCh, <-errCh
}

// runRoot executes the CLI with args through the package Execute function
// (PersistentPreRun and the unreported-error printing included), with cobra.s
// writers pointed at the real process streams.
func runRoot(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	resetCommandFlags(t, rootCmd)
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		resetCommandFlags(t, rootCmd)
		rootCmd.SetArgs(nil)
		jsonOutput, plainOutput = false, false
	})

	stdout, stderr = captureStreams(t, func() {
		err = Execute()
	})
	return stdout, stderr, err
}

func TestReportErrorStreams(t *testing.T) {
	t.Run("PlainErrorGoesToStderrOnce", func(t *testing.T) {
		stdout, stderr, err := runRoot(t, "ca", "--plain")
		if err == nil {
			t.Fatal("expected an error for a missing --cn")
		}
		if stdout != "" {
			t.Errorf("stdout should be empty when the command fails, got %q", stdout)
		}
		const msg = "common name (--cn) is required"
		if n := strings.Count(stderr, msg); n != 1 {
			t.Errorf("stderr should carry the error exactly once, found %d in %q", n, stderr)
		}
		if !strings.Contains(stderr, "Error: "+msg) {
			t.Errorf("stderr should use the styled Error: prefix, got %q", stderr)
		}
	})

	t.Run("JSONErrorGoesToStdoutOnly", func(t *testing.T) {
		stdout, stderr, err := runRoot(t, "ca", "--json")
		if err == nil {
			t.Fatal("expected an error for a missing --cn")
		}
		if stderr != "" {
			t.Errorf("stderr should be empty in JSON mode, got %q", stderr)
		}
		var payload cert.JSONOperationResult
		if jsonErr := json.Unmarshal([]byte(stdout), &payload); jsonErr != nil {
			t.Fatalf("stdout should be a JSON error payload, got %q: %v", stdout, jsonErr)
		}
		if payload.Success || payload.Error != "common name (--cn) is required" {
			t.Errorf("unexpected payload %+v", payload)
		}
	})

	t.Run("GenerateUsesSameWording", func(t *testing.T) {
		// Cobra.s required-flag check fires before RunE, so exercise the
		// command.s own guard directly.
		setOutputMode(t, false, true)
		old := generateCN
		generateCN = ""
		t.Cleanup(func() { generateCN = old })

		var err error
		_, stderr := captureStreams(t, func() {
			err = generateCmd.RunE(generateCmd, nil)
		})
		if err == nil {
			t.Fatal("expected an error for a missing --cn")
		}
		if !strings.Contains(stderr, "common name (--cn) is required") {
			t.Errorf("generate should use the shared --cn wording, got %q", stderr)
		}
	})

	t.Run("CobraErrorsStillPrintedAfterReportedRun", func(t *testing.T) {
		// A reported failure must not mute a later flag-parse error, which
		// happens before PersistentPreRun ever runs.
		_, _, _ = runRoot(t, "ca", "--plain")
		stdout, stderr, err := runRoot(t, "ca", "--no-such-flag")
		if err == nil {
			t.Fatal("expected a flag parse error")
		}
		if stdout != "" {
			t.Errorf("stdout should be empty, got %q", stdout)
		}
		if n := strings.Count(stderr, "unknown flag"); n != 1 {
			t.Errorf("flag error should reach stderr exactly once, found %d in %q", n, stderr)
		}
	})

	t.Run("UnreportedFailurePrintedOnceOnStderr", func(t *testing.T) {
		// verify prints its results itself and returns a bare error; Execute
		// must print that error once, after the results, on stderr.
		stdout, stderr, err := runRoot(t, "verify", testutil.TestdataPath("expired.pem"), "--plain")
		if err == nil {
			t.Fatal("expected verification of an expired certificate to fail")
		}
		if !strings.Contains(stdout, "Certificate has expired") {
			t.Errorf("verification results should stay on stdout, got %q", stdout)
		}
		if n := strings.Count(stderr, "Error: verification failed"); n != 1 {
			t.Errorf("stderr should carry the failure exactly once, found %d in %q", n, stderr)
		}
		if strings.Contains(stdout, "verification failed") {
			t.Errorf("the bare error must not be duplicated on stdout, got %q", stdout)
		}
	})
}
