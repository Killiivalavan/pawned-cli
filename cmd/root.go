package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"chesshell-cli/internal/store"
	"chesshell-cli/internal/update"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip update check if this is the update command itself
		if cmd.Use == "update" {
			return
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			data, err := store.Load()
			if err != nil {
				return
			}

			// 1. Read data.LastUpdateCheck — skip if < 24h ago
			if data.Stats.LastUpdateCheck != nil && time.Since(*data.Stats.LastUpdateCheck) < 24*time.Hour {
				return
			}

			v := version
			if v == "" {
				v = "v0.1.0-dev"
			}

			// 2. Call update.Checker.CheckLatest()
			checker := update.NewChecker(v, "Killiivalavan", "chesshell-cli")
			latest, isNewer, err := checker.CheckLatest()
			if err != nil {
				return
			}

			// 4. Update timestamps asynchronously
			now := time.Now()
			data.Stats.LastUpdateCheck = &now
			
			// 3. If newer version found and not already acknowledged
			if isNewer && latest != data.Stats.LastUpdateAcknowledged {
				fmt.Printf("\n⭐ A new version %s is available. Run 'chesshell update' to upgrade. ⭐\n\n", latest)
				data.Stats.LastUpdateAcknowledged = latest
			}
			
			store.Save(data)
		}()

		// Wait for the goroutine to finish or timeout after 3 seconds
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
