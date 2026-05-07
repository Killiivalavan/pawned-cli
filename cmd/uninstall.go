package cmd

import (
	"bufio"
	"chesshell-cli/internal/store"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Permanently remove chesshell from your system.",
	Run: func(cmd *cobra.Command, args []string) {
		// Step 1: Resolve paths
		dataDir, err := store.GetDataDir()
		if err != nil {
			fmt.Printf("Error resolving data directory: %v\n", err)
			os.Exit(1)
		}

		binaryPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error resolving binary path: %v\n", err)
			os.Exit(1)
		}

		// Show confirmation
		fmt.Println("This will permanently remove chesshell from your system.")
		fmt.Println("\nThe following will be deleted:")
		fmt.Printf("  • chesshell binary     %s\n", binaryPath)
		fmt.Printf("  • All local data       %s\n", dataDir)
		fmt.Println("\nYour puzzle history and stats will be permanently deleted and cannot be recovered.")
		fmt.Print("\nAre you sure? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) != "y" {
			fmt.Println("Uninstall cancelled.")
			os.Exit(0)
		}

		// Step 2: Delete data directory
		fmt.Printf("Deleting data directory: %s...\n", dataDir)
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Printf("Error deleting data directory: %v\n", err)
			os.Exit(1)
		}

		// Step 4: Print goodbye and delete binary
		fmt.Println("We are sad to see you go. chesshell has been removed. Goodbye.")
		if err := os.Remove(binaryPath); err != nil {
			fmt.Printf("\nError deleting binary: %v\n", err)
			fmt.Printf("Please manually delete the binary at: %s\n", binaryPath)
		}

		// Step 5: Exit
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
