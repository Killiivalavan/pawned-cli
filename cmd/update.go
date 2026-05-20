package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"chesshell-cli/internal/update"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update chesshell to the latest version.",
	Long:  `Checks for new releases on GitHub and updates the application to the latest version if one is available.`,
	Run: func(cmd *cobra.Command, args []string) {
		v := version
		if v == "" {
			v = "v0.1.0-dev"
		}
		
		fmt.Printf("Checking for updates (current version: %s)...\n", v)
		
		checker := update.NewChecker(v, "Killiivalavan", "chesshell-cli")
		latest, isNewer, err := checker.CheckLatest()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\nCheck your internet connection and try again.\n", err)
			return
		}

		if !isNewer {
			fmt.Println("chesshell is already up to date.")
			return
		}

		fmt.Printf("Update available: %s → %s. Apply? (y/N): ", v, latest)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			fmt.Println("Update cancelled.")
			return
		}

		fmt.Printf("Downloading %s for %s/%s...\n", latest, runtime.GOOS, runtime.GOARCH)
		
		assetURL := checker.GetAssetURL(latest, runtime.GOOS, runtime.GOARCH)
		
		err = applyUpdateWithRetry(checker, assetURL, 1)
		if err != nil {
			fmt.Printf("\nFailed to apply update: %v\n", err)
			if strings.Contains(err.Error(), "permission denied") {
				fmt.Println("Hint: Try running the command with 'sudo' or install manually.")
			}
			return
		}

		fmt.Println("\nUpdate applied successfully! You are now running", latest)
	},
}

var updateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a new version is available",
	Run: func(cmd *cobra.Command, args []string) {
		v := version
		if v == "" {
			v = "v0.1.0-dev"
		}
		
		fmt.Printf("Current version: %s\n", v)
		fmt.Println("Checking for latest release...")
		
		checker := update.NewChecker(v, "Killiivalavan", "chesshell-cli")
		latest, isNewer, err := checker.CheckLatest()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		fmt.Printf("Latest version:  %s\n", latest)
		
		if isNewer {
			fmt.Printf("\nA new update is available! Run 'chesshell update' to install.\n")
		} else {
			fmt.Println("\nYou have the latest version.")
		}
	},
}

func applyUpdateWithRetry(checker *update.Checker, assetURL string, retries int) error {
	var err error
	for i := 0; i <= retries; i++ {
		if i > 0 {
			fmt.Println("Retrying download...")
		}
		err = checker.ApplyUpdate(assetURL)
		if err == nil {
			return nil
		}
	}
	return err
}

func init() {
	updateCmd.AddCommand(updateCheckCmd)
	rootCmd.AddCommand(updateCmd)
}
