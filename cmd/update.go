package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const installerTimeout = 30 * time.Second

// installerURL is where the installer script is fetched from. It is a
// variable so tests can point it at a local server.
var installerURL = "https://raw.githubusercontent.com/trahma/certwiz/main/install.sh"

var forceUpdate bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update cert to the latest version",
	Long: `Update cert to the latest version by downloading and running the installer.

This command will:
1. Download the installer script
2. Run it, which checks the latest release against your current version
3. If an update is available, the installer upgrades cert in place`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "windows" {
			fmt.Println("Please download the latest version from:")
			fmt.Println("  https://github.com/trahma/certwiz/releases")
			return fmt.Errorf("auto-update is not supported on Windows")
		}

		currentVersion := strings.TrimPrefix(version, "v")
		fmt.Printf("Current version: v%s\n", currentVersion)

		fmt.Println("Downloading installer...")
		installerPath, err := downloadInstaller()
		if err != nil {
			return err
		}

		// Clear extended attributes on macOS
		if runtime.GOOS == "darwin" {
			_ = exec.Command("xattr", "-cr", installerPath).Run() // xattr might not be available
		}

		fmt.Println("Running installer...")

		bashPath, err := exec.LookPath("bash")
		if err != nil {
			_ = os.Remove(installerPath)
			return fmt.Errorf("error finding bash: %w", err)
		}

		argv := installerArgv(installerPath, forceUpdate)

		// Replace the current process with bash running the installer so it
		// runs in a clean context; the wrapper removes the script when it
		// exits. Exec only returns on failure.
		if err := syscall.Exec(bashPath, argv, os.Environ()); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing installer: %v\n", err)
			// Fallback to a child process; skip argv[0]
			child := exec.Command(bashPath, argv[1:]...)
			child.Stdin = os.Stdin
			child.Stdout = os.Stdout
			child.Stderr = os.Stderr
			runErr := child.Run()
			_ = os.Remove(installerPath)
			if runErr != nil {
				return fmt.Errorf("error running installer: %w", runErr)
			}
		}
		return nil
	},
}

// installerWrapper is the bash -c script that runs the downloaded installer
// (passed as $1) with any remaining arguments and removes it on exit, so the
// temp file is cleaned up even though the installer replaces this process.
const installerWrapper = `script="$1"; shift; trap 'rm -f "$script"' EXIT; bash "$script" "$@"`

// installerArgv builds the argv for exec'ing bash with installerWrapper.
// argv[0] is the program name as required by syscall.Exec; $0 inside the
// wrapper is "cert-update" and $1 is the installer path.
func installerArgv(installerPath string, force bool) []string {
	argv := []string{"bash", "-c", installerWrapper, "cert-update", installerPath}
	if force {
		argv = append(argv, "--force")
	}
	return argv
}

// downloadInstaller fetches the installer script into a private temporary
// file and returns its path. The caller is responsible for removing it.
func downloadInstaller() (string, error) {
	client := &http.Client{Timeout: installerTimeout}
	resp, err := client.Get(installerURL)
	if err != nil {
		return "", fmt.Errorf("error downloading installer: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error downloading installer: unexpected status %s", resp.Status)
	}

	installerFile, err := os.CreateTemp("", "certwiz-installer-*.sh")
	if err != nil {
		return "", fmt.Errorf("error creating installer file: %w", err)
	}
	installerPath := installerFile.Name()

	_, copyErr := io.Copy(installerFile, resp.Body)
	closeErr := installerFile.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(installerPath)
		if copyErr != nil {
			return "", fmt.Errorf("error writing installer: %w", copyErr)
		}
		return "", fmt.Errorf("error writing installer: %w", closeErr)
	}

	if err := os.Chmod(installerPath, 0755); err != nil {
		_ = os.Remove(installerPath)
		return "", fmt.Errorf("error making installer executable: %w", err)
	}

	return installerPath, nil
}

func init() {
	updateCmd.Flags().BoolVar(&forceUpdate, "force", false, "Force update even if already on latest version")
	rootCmd.AddCommand(updateCmd)
}
