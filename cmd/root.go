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
and lets you play games against a local AI directly in the terminal.

Key Features:
  - Graphical Board: Optional high-fidelity rendering for modern terminals.
  - Game Resume: Pick up your AI games exactly where you left off.
  - Zero Config: Works out of the box with local-first stats and history.
  - Privacy Focused: No accounts required; all data stays on your machine.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
