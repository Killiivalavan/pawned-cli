package cmd

import (
	"fmt"
	"os"
	"pawned-cli/internal/store"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display the user's local stats summary.",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := store.Load()
		if err != nil && err != store.ErrCorruptedFile {
			fmt.Printf("Error loading stats: %v\n", err)
			os.Exit(1)
		}
		if err == store.ErrCorruptedFile {
			fmt.Println("Warning: Corrupted data file was found and backed up.")
		}

		stats := data.Stats

		header := color.New(color.Bold, color.Underline)
		header.Println("Your Stats")

		if stats.TotalAttempted == 0 {
			fmt.Println("\nNo puzzles played yet. Use 'pawned play' to start!")
			return
		}

		// Calculate accuracy
		accuracy := 0.0
		if stats.TotalAttempted > 0 {
			accuracy = (float64(stats.TotalSolved) / float64(stats.TotalAttempted)) * 100
		}

		fmt.Printf("\n%-20s %d\n", "Puzzles Attempted:", stats.TotalAttempted)
		fmt.Printf("%-20s %d\n", "Puzzles Solved:", stats.TotalSolved)
		fmt.Printf("%-20s %.2f%%\n", "Accuracy:", accuracy)

		fmt.Println()
		// TODO: Implement streak calculation
		fmt.Printf("%-20s %d days\n", "Current Streak:", stats.CurrentStreak)
		fmt.Printf("%-20s %d days\n", "Best Streak:", stats.BestStreak)

		fmt.Println()
		if stats.FirstPlayedAt != nil {
			fmt.Printf("First puzzle played on %s.\n", stats.FirstPlayedAt.Format("Jan 2, 2006"))
		}

		// AI Games Stats
		fmt.Println()
		header.Println("AI Games")
		if data.AIGames.Wins == 0 && data.AIGames.Losses == 0 && data.AIGames.Draws == 0 {
			fmt.Println("\nNo AI games played yet. Use 'pawned play --ai' to start!")
		} else {
			totalAIGames := data.AIGames.Wins + data.AIGames.Losses + data.AIGames.Draws
			winRate := 0.0
			if totalAIGames > 0 {
				winRate = (float64(data.AIGames.Wins) / float64(totalAIGames)) * 100
			}
			fmt.Printf("\n%-20s %d\n", "Games Played:", totalAIGames)
			fmt.Printf("%-20s %d\n", "Wins:", data.AIGames.Wins)
			fmt.Printf("%-20s %d\n", "Losses:", data.AIGames.Losses)
			fmt.Printf("%-20s %d\n", "Draws:", data.AIGames.Draws)
			fmt.Printf("%-20s %.2f%%\n", "Win Rate:", winRate)
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
