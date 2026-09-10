package cmd

import (
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"certwiz/internal/testutil"
)

func TestConvertCommandRun(t *testing.T) {
	tmpDir := t.TempDir()

	resetConvertFlags := func(t *testing.T) {
		t.Helper()
		old := convertFormat
		convertFormat = "pem"
		t.Cleanup(func() { convertFormat = old })
	}

	t.Run("DERToPEM", func(t *testing.T) {
		resetConvertFlags(t)
		setOutputMode(t, false, true)
		outPath := filepath.Join(tmpDir, "from-der.pem")

		var err error
		out := captureStdout(t, func() {
			err = convertCmd.RunE(convertCmd, []string{testutil.TestdataPath("valid.der"), outPath})
		})
		if err != nil {
			t.Fatalf("convert failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "Converted from DER to PEM") {
			t.Errorf("expected conversion message, got:\n%s", out)
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		block, _ := pem.Decode(data)
		if block == nil || block.Type != "CERTIFICATE" {
			t.Errorf("output is not a PEM certificate")
		}
	})

	t.Run("PEMToDER", func(t *testing.T) {
		resetConvertFlags(t)
		setOutputMode(t, false, true)
		convertFormat = "der"
		outPath := filepath.Join(tmpDir, "from-pem.der")

		var err error
		out := captureStdout(t, func() {
			err = convertCmd.RunE(convertCmd, []string{testutil.TestdataPath("valid.pem"), outPath})
		})
		if err != nil {
			t.Fatalf("convert failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "Converted from PEM to DER") {
			t.Errorf("expected conversion message, got:\n%s", out)
		}
		want, _ := os.ReadFile(testutil.TestdataPath("valid.der"))
		got, _ := os.ReadFile(outPath)
		if string(got) != string(want) {
			t.Error("DER output should match the reference valid.der")
		}
	})

	t.Run("BundleToPEMKeepsAllCertificates", func(t *testing.T) {
		resetConvertFlags(t)
		setOutputMode(t, true, false)
		outPath := filepath.Join(tmpDir, "bundle.pem")

		var err error
		captureStdout(t, func() {
			err = convertCmd.RunE(convertCmd, []string{testutil.TestdataPath("fullchain.pem"), outPath})
		})
		if err != nil {
			t.Fatalf("convert failed: %v", err)
		}
		data, _ := os.ReadFile(outPath)
		if n := strings.Count(string(data), "BEGIN CERTIFICATE"); n != 3 {
			t.Errorf("expected 3 certificates in output, got %d", n)
		}
	})

	t.Run("BundleToDERFails", func(t *testing.T) {
		resetConvertFlags(t)
		setOutputMode(t, true, false)
		convertFormat = "der"
		outPath := filepath.Join(tmpDir, "bundle.der")

		var err error
		out := captureStdout(t, func() {
			err = convertCmd.RunE(convertCmd, []string{testutil.TestdataPath("fullchain.pem"), outPath})
		})
		if err == nil {
			t.Fatal("expected an error converting a bundle to DER")
		}
		if !strings.Contains(out, `"success": false`) {
			t.Errorf("expected JSON error payload, got:\n%s", out)
		}
		if _, statErr := os.Stat(outPath); statErr == nil {
			t.Error("no output file should be written on failure")
		}
	})

	t.Run("UnsupportedFormat", func(t *testing.T) {
		resetConvertFlags(t)
		setOutputMode(t, false, true)
		convertFormat = "p7b"

		var err error
		captureStdout(t, func() {
			err = convertCmd.RunE(convertCmd, []string{testutil.TestdataPath("valid.pem"), filepath.Join(tmpDir, "x.p7b")})
		})
		if err == nil || !strings.Contains(err.Error(), "unsupported format") {
			t.Fatalf("expected unsupported format error, got %v", err)
		}
	})

	t.Run("MissingInput", func(t *testing.T) {
		resetConvertFlags(t)
		setOutputMode(t, false, true)

		var err error
		captureStdout(t, func() {
			err = convertCmd.RunE(convertCmd, []string{filepath.Join(tmpDir, "nope.pem"), filepath.Join(tmpDir, "out.pem")})
		})
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("expected does-not-exist error, got %v", err)
		}
	})
}
