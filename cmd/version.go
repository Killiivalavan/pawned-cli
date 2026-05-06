package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// This will be set during the build process
var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of chesshell.",
	Run: func(cmd *cobra.Command, args []string) {
		if version == "" {
			version = "v0.1.0-dev" // Default for local development
		}
		fmt.Printf("chesshell version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
