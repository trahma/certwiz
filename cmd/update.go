package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

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
		
		// Use bash to download and execute the installer
		installerURL := "https://raw.githubusercontent.com/trahma/certwiz/main/install.sh"
		
		// Prepare the bash command
		var bashCmd *exec.Cmd
		if forceUpdate {
			// Force reinstall even if version is the same
			bashCmd = exec.Command("bash", "-c", 
				fmt.Sprintf("curl -sSL %s | bash -s -- --force", installerURL))
		} else {
			// Normal update - will check version and only update if newer
			bashCmd = exec.Command("bash", "-c", 
				fmt.Sprintf("curl -sSL %s | bash", installerURL))
		}
		
		// Connect stdin, stdout, stderr so user can interact if needed
		bashCmd.Stdin = os.Stdin
		bashCmd.Stdout = os.Stdout
		bashCmd.Stderr = os.Stderr
		
		// Run the installer
		if err := bashCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running update: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	updateCmd.Flags().BoolVar(&forceUpdate, "force", false, "Force update even if already on latest version")
	rootCmd.AddCommand(updateCmd)
}// Force CI refresh
