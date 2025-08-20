package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var forceUpdate bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update cert to the latest version",
	Long: `Update cert to the latest version by downloading and running the installer.

This command will:
1. Check for the latest available version
2. Compare with your current version
3. If an update is available, download and run the installer
4. The installer will upgrade cert in place`,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS == "windows" {
			fmt.Println("Auto-update is not supported on Windows.")
			fmt.Println("Please download the latest version from:")
			fmt.Println("  https://github.com/trahma/certwiz/releases")
			os.Exit(1)
		}

		fmt.Println("Checking for updates...")

		// Check current version
		currentVersion := strings.TrimPrefix(version, "v")
		fmt.Printf("Current version: v%s\n", currentVersion)

		// Download and run the installer script
		fmt.Println("\nFetching latest version information...")

		// Download the installer script to a temp file
		installerURL := "https://raw.githubusercontent.com/trahma/certwiz/main/install.sh"
		
		// Create temp file for installer script
		tempDir := os.TempDir()
		installerPath := filepath.Join(tempDir, "certwiz-installer.sh")
		
		// Download the installer
		fmt.Println("Downloading installer...")
		resp, err := http.Get(installerURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading installer: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		
		// Create the installer file
		installerFile, err := os.Create(installerPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating installer file: %v\n", err)
			os.Exit(1)
		}
		
		// Write the installer content
		_, err = io.Copy(installerFile, resp.Body)
		installerFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing installer: %v\n", err)
			os.Exit(1)
		}
		
		// Make installer executable
		if err := os.Chmod(installerPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error making installer executable: %v\n", err)
			os.Exit(1)
		}
		
		// Clear extended attributes on macOS
		if runtime.GOOS == "darwin" {
			xattrCmd := exec.Command("xattr", "-cr", installerPath)
			_ = xattrCmd.Run() // Ignore errors, xattr might not be available
		}
		
		// Prepare arguments for the installer
		// For syscall.Exec, the first argument must be the program name itself
		installerArgs := []string{"bash", installerPath}
		if forceUpdate {
			installerArgs = append(installerArgs, "--force")
		}
		
		fmt.Println("Running installer...")
		
		// Use syscall.Exec to replace the current process with the installer
		// This breaks the inheritance chain that might be causing issues
		env := os.Environ()
		
		// Find bash executable
		bashPath, err := exec.LookPath("bash")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding bash: %v\n", err)
			os.Exit(1)
		}
		
		// Replace current process with bash running the installer
		// This ensures the installer runs in a clean context
		if err := syscall.Exec(bashPath, installerArgs, env); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing installer: %v\n", err)
			// Fallback to regular exec if syscall.Exec fails
			// Skip the first "bash" argument for exec.Command
			cmd := exec.Command("bash", installerArgs[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running installer: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	updateCmd.Flags().BoolVar(&forceUpdate, "force", false, "Force update even if already on latest version")
	rootCmd.AddCommand(updateCmd)
}
