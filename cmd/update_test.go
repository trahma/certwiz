package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDownloadInstaller(t *testing.T) {
	const body = "#!/bin/bash\necho installer\n"

	mux := http.NewServeMux()
	mux.HandleFunc("/install.sh", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	oldURL := installerURL
	t.Cleanup(func() { installerURL = oldURL })

	t.Run("Success", func(t *testing.T) {
		installerURL = server.URL + "/install.sh"

		path, err := downloadInstaller()
		if err != nil {
			t.Fatalf("downloadInstaller: %v", err)
		}
		t.Cleanup(func() { _ = os.Remove(path) })

		if !strings.HasPrefix(filepath.Base(path), "certwiz-installer-") || !strings.HasSuffix(path, ".sh") {
			t.Errorf("unexpected installer path %q", path)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read installer: %v", err)
		}
		if string(got) != body {
			t.Errorf("installer content = %q, want %q", got, body)
		}
		if runtime.GOOS != "windows" {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat installer: %v", err)
			}
			if info.Mode().Perm()&0100 == 0 {
				t.Errorf("installer should be owner-executable, mode %v", info.Mode().Perm())
			}
		}
	})

	t.Run("NotFoundLeavesNoFile", func(t *testing.T) {
		installerURL = server.URL + "/missing.sh"

		before := countInstallerFiles(t)
		path, err := downloadInstaller()
		if err == nil {
			_ = os.Remove(path)
			t.Fatal("expected an error for a 404 response")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error should mention the status, got: %v", err)
		}
		if after := countInstallerFiles(t); after != before {
			t.Errorf("installer temp files changed from %d to %d on failure", before, after)
		}
	})

	t.Run("UnreachableServer", func(t *testing.T) {
		unreachable := httptest.NewServer(http.NotFoundHandler())
		unreachable.Close()
		installerURL = unreachable.URL + "/install.sh"

		if _, err := downloadInstaller(); err == nil {
			t.Fatal("expected an error when the server is unreachable")
		}
	})
}

// countInstallerFiles counts leftover installer temp files in the temp dir.
func countInstallerFiles(t *testing.T) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "certwiz-installer-*.sh"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	return len(matches)
}

func TestInstallerArgv(t *testing.T) {
	const path = "/tmp/certwiz-installer-123.sh"

	argv := installerArgv(path, false)
	if argv[0] != "bash" || argv[1] != "-c" {
		t.Fatalf("argv should exec bash -c, got %v", argv)
	}
	wrapper := argv[2]
	if !strings.Contains(wrapper, "trap 'rm -f \"$script\"' EXIT") {
		t.Errorf("wrapper should trap EXIT to remove the script, got %q", wrapper)
	}
	if !strings.Contains(wrapper, `script="$1"`) {
		t.Errorf("wrapper should read the script path from $1, got %q", wrapper)
	}
	// argv[3] is $0 inside the wrapper; argv[4] is $1, the installer path.
	if argv[4] != path {
		t.Errorf("installer path should be $1, got argv[4]=%q", argv[4])
	}
	for _, a := range argv {
		if a == "--force" {
			t.Errorf("--force should not be present when force is false: %v", argv)
		}
	}

	forced := installerArgv(path, true)
	if forced[len(forced)-1] != "--force" {
		t.Errorf("--force should be the last argument when force is true: %v", forced)
	}
	if len(forced) != len(argv)+1 {
		t.Errorf("force should add exactly one argument: %v vs %v", forced, argv)
	}
}

// TestInstallerWrapperRemovesScript runs the wrapper through a real bash
// (as a child, not exec) to prove the installer receives its arguments and
// the script is removed afterwards.
func TestInstallerWrapperRemovesScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash wrapper is not used on Windows")
	}
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	dir := t.TempDir()
	script := filepath.Join(dir, "installer.sh")
	argsFile := filepath.Join(dir, "args.txt")
	body := "#!/bin/bash\nprintf '%s\\n' \"$@\" > " + argsFile + "\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	argv := installerArgv(script, true)
	cmd := exec.Command(bashPath, argv[1:]...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("wrapper failed: %v\n%s", err, out)
	}

	got, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("installer did not run: %v", err)
	}
	if strings.TrimSpace(string(got)) != "--force" {
		t.Errorf("installer args = %q, want --force", got)
	}
	if _, err := os.Stat(script); !os.IsNotExist(err) {
		t.Errorf("installer script should be removed after the wrapper exits, stat err=%v", err)
	}
}
