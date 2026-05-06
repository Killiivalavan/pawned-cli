package cmd

import (
	"chesshell-cli/internal/store"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show the last puzzles the user attempted.",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")

		data, err := store.Load()
		if err != nil && err != store.ErrCorruptedFile {
			fmt.Printf("Error loading history: %v\n", err)
			os.Exit(1)
		}
		if err == store.ErrCorruptedFile {
			fmt.Println("Warning: Corrupted data file was found and backed up.")
		}

		if len(data.History) == 0 {
			fmt.Println("No history available. Play some puzzles first!")
			return
		}

		// Initialize tabwriter for table formatting
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		headerColor := color.New(color.Bold)

		fmt.Println()
		headerColor.Fprintln(w, "Date\tPuzzle ID\tRating\tThemes\tResult\tAttempts")
		headerColor.Fprintln(w, "----\t---------\t------\t------\t------\t--------")

		// Determine bounds
		startIdx := len(data.History) - 1
		endIdx := startIdx - limit + 1
		if endIdx < 0 {
			endIdx = 0
		}

		successColor := color.New(color.FgGreen)
		failColor := color.New(color.FgRed)

		// Iterate backwards to show newest first
		for i := startIdx; i >= endIdx; i-- {
			item := data.History[i]

			dateStr := item.AttemptedAt.Format("2006-01-02 15:04")

			// Format themes (truncate if too long for the table)
			themesStr := strings.Join(item.Themes, ", ")
			if len(themesStr) > 30 {
				themesStr = themesStr[:27] + "..."
			}

			resultStr := "Failed"
			resultFmt := failColor.Sprint(resultStr)
			if item.Solved {
				resultStr = "Solved"
				resultFmt = successColor.Sprint(resultStr)
			}

			// Add the lichess link format as requested in PRD
			idFmt := fmt.Sprintf("%s (lichess.org/training/%s)", item.PuzzleID, item.PuzzleID)

			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%d\n",
				dateStr,
				idFmt,
				item.Rating,
				themesStr,
				resultFmt,
				item.Attempts,
			)
		}

		w.Flush()
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.Flags().Int("limit", 10, "Specify the number of history entries to show.")
}
