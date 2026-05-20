package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// This will be set during the build process
var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the currently installed version.",
	Long:  `Prints the version number of the currently installed chesshell binary.`,
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
