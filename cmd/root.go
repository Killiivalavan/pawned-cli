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
	Long: `A beautiful, command-line chess experience.

chesshell is a modern CLI for chess enthusiasts. Play against a local AI, 
challenge friends to a multiplayer match, or solve daily puzzles from Lichess—all 
from the comfort of your terminal.

FEATURES
  • Play Modes:       Challenge a local Stockfish AI with adjustable difficulty,
                      play online against a friend, or solve daily puzzles.
  • Multiplayer:      Create a game with a simple join code and play against
                      anyone, anywhere.
  • Self-Updating:    Notifies you of new versions and lets you upgrade with a
                      single command ('chesshell update').
  • Graphical Board:  A high-fidelity Unicode board that renders beautifully in
                      modern terminals.
  • Local-First:      All your stats and game history are stored locally.
                      No accounts, no cloud sync, no nonsense.`,
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
