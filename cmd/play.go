package cmd

import (
	"fmt"
	"os"
	"pawned-cli/internal/api"
	"pawned-cli/internal/puzzle"
	"pawned-cli/internal/store"
	"time"

	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Fetches a puzzle and starts an interactive solving session.",
	Long:  `Fetches today's daily puzzle from Lichess and starts an interactive solving session. You can also provide a specific puzzle ID to play.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Fetch Puzzle
		id, _ := cmd.Flags().GetString("id")
		apiClient := api.NewClient()
		var lichessPuzzle *api.LichessPuzzle
		var err error
		if id != "" {
			fmt.Printf("Fetching puzzle with ID: %s...\n", id)
			lichessPuzzle, err = apiClient.FetchByID(id)
		} else {
			fmt.Println("Fetching daily puzzle...")
			lichessPuzzle, err = apiClient.FetchDaily()
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// 2. Run Session
		session, err := puzzle.NewSession(lichessPuzzle, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Printf("Error creating session: %v\n", err)
			os.Exit(1)
		}
		solved, attempts, err := session.Run()
		if err != nil {
			fmt.Printf("Error during session: %v\n", err)
			os.Exit(1)
		}

		// A puzzle is only recorded if at least one move was attempted.
		// Quitting immediately does not count.
		if attempts == 0 && !solved {
			return
		}

		// 3. Load, Update, and Save Data
		data, err := store.Load()
		if err != nil && err != store.ErrCorruptedFile {
			fmt.Printf("Error loading stats: %v\n", err)
			os.Exit(1)
		}
		if err == store.ErrCorruptedFile {
			fmt.Println("Warning: Corrupted data file was found and backed up.")
		}

		// Create history item
		historyItem := store.HistoryItem{
			PuzzleID:    lichessPuzzle.Puzzle.ID,
			Rating:      lichessPuzzle.Puzzle.Rating,
			Themes:      lichessPuzzle.Puzzle.Themes,
			AttemptedAt: time.Now(),
			Solved:      solved,
			Attempts:    attempts,
		}
		data.History = append(data.History, historyItem)

		// Update stats
		data.Stats.TotalAttempted++
		if solved {
			data.Stats.TotalSolved++
		}
		// TODO: Streak calculation
		now := time.Now()
		if data.Stats.FirstPlayedAt == nil {
			data.Stats.FirstPlayedAt = &now
		}
		data.Stats.LastPlayedAt = &now

		// Limit history
		if len(data.History) > 200 {
			data.History = data.History[len(data.History)-200:]
		}

		if err := store.Save(data); err != nil {
			fmt.Printf("Error saving stats: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Your progress has been saved.")
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().String("id", "", "Play a specific puzzle by its Lichess ID.")
}
