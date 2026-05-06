package cmd

import (
	"bufio"
	"fmt"
	"os"
	"pawned-cli/internal/api"
	"pawned-cli/internal/engine"
	"pawned-cli/internal/puzzle"
	"pawned-cli/internal/store"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Fetches a puzzle or starts an AI game.",
	Long:  `Starts an interactive solving session for a Lichess puzzle or plays a full game against a local AI.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		ai, _ := cmd.Flags().GetBool("ai")

		if id != "" && ai {
			fmt.Println("Error: Cannot use --id and --ai at the same time.")
			os.Exit(1)
		}

		if ai {
			runAIGame()
			return
		}

		runPuzzle(id)
	},
}

func runAIGame() {
	fmt.Println("Select AI Difficulty:")
	fmt.Println("1. Beginner")
	fmt.Println("2. Casual")
	fmt.Println("3. Intermediate")
	fmt.Println("4. Advanced")
	fmt.Println("5. Expert")
	
	reader := bufio.NewReader(os.Stdin)
	var difficulty int
	for {
		fmt.Print("\nEnter choice (1-5): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		val, err := strconv.Atoi(input)
		if err == nil && val >= 1 && val <= 5 {
			difficulty = val
			break
		}
		fmt.Println("Invalid input. Please enter a number between 1 and 5.")
	}

	// Map 1-5 to Stockfish skill levels (0-20)
	skillLevel := (difficulty - 1) * 5
	fmt.Printf("Starting AI game with Skill Level %d (Difficulty %d)...\n", skillLevel, difficulty)

	path, err := engine.GetEnginePath()
	if err != nil {
		fmt.Printf("Error setting up AI engine: %v\n", err)
		os.Exit(1)
	}

	eng, err := engine.Start(path)
	if err != nil {
		fmt.Printf("Error starting engine: %v\n", err)
		os.Exit(1)
	}
	defer eng.Close()

	if err := eng.Configure(skillLevel); err != nil {
		fmt.Printf("Error configuring engine: %v\n", err)
		os.Exit(1)
	}

	session := puzzle.NewAISession(eng, os.Stdin, os.Stdout)
	result, err := session.Run()
	if err != nil {
		fmt.Printf("Error during game: %v\n", err)
		os.Exit(1)
	}

	if result == "Abandoned" {
		return
	}

	// 5. Load, Update, and Save Data
	data, err := store.Load()
	if err != nil && err != store.ErrCorruptedFile {
		fmt.Printf("Error loading stats: %v\n", err)
		os.Exit(1)
	}
	if err == store.ErrCorruptedFile {
		fmt.Println("Warning: Corrupted data file was found and backed up.")
	}

	switch session.Board.Outcome() {
	case "1-0": // White won (User is always White in v2 AI games)
		data.AIGames.Wins++
	case "0-1": // Black won
		data.AIGames.Losses++
	case "1/2-1/2": // Draw
		data.AIGames.Draws++
	}

	if err := store.Save(data); err != nil {
		fmt.Printf("Error saving stats: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("AI game stats have been saved.")
}

func runPuzzle(id string) {
	// 1. Fetch Puzzle
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
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().String("id", "", "Play a specific puzzle by its Lichess ID.")
	playCmd.Flags().Bool("ai", false, "Play a full game against a local AI.")
}
