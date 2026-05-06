package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chesshell",
	Short: "A minimal, elegant, command-line chess puzzle tool.",
	Long: `chesshell is a CLI tool that fetches chess puzzles from Lichess
and lets users solve them directly in the terminal.

No account required, no configuration needed. All stats and history
are stored locally on your machine.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
