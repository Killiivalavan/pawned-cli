package cmd

import (
	"bufio"
	"chesshell-cli/internal/api"
	"chesshell-cli/internal/board"
	"chesshell-cli/internal/engine"
	"chesshell-cli/internal/puzzle"
	"chesshell-cli/internal/relay"
	"chesshell-cli/internal/store"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/corentings/chess"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Fetches a puzzle or starts an AI game.",
	Long:  `Starts an interactive solving session for a Lichess puzzle or plays a full game against a local AI.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		ai, _ := cmd.Flags().GetBool("ai")
		multiplayer, _ := cmd.Flags().GetBool("multiplayer")
		joinCode, _ := cmd.Flags().GetString("join")
		unicodeFlag, _ := cmd.Flags().GetBool("unicode")

		count := 0
		if id != "" { count++ }
		if ai { count++ }
		if multiplayer { count++ }
		if joinCode != "" { count++ }

		if count > 1 {
			fmt.Println("Error: Cannot use --id, --ai, --multiplayer, or --join simultaneously.")
			os.Exit(1)
		}

		// 1. Load Data to get config
		data, err := store.Load()
		if err != nil && err != store.ErrCorruptedFile {
			fmt.Printf("Error loading stats: %v\n", err)
			os.Exit(1)
		}
		if err == store.ErrCorruptedFile {
			fmt.Println("Warning: Corrupted data file was found and backed up.")
		}

		// 2. Resolve Graphical Board (Unicode) preference
		useUnicode := board.IsUnicodeSupported()
		configChanged := false

		if data.Config.Unicode == nil {
			// First time setup or migration
			if useUnicode {
				fmt.Print("Graphical Board support detected! Would you like to enable high-fidelity rendering? (Y/n): ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.ToLower(strings.TrimSpace(input))
				
				choice := true
				if input == "n" || input == "no" {
					choice = false
				}
				data.Config.Unicode = &choice
				useUnicode = choice
				configChanged = true
			} else {
				f := false
				data.Config.Unicode = &f
				useUnicode = false
				configChanged = true
			}
		} else if !*data.Config.Unicode && useUnicode {
			// Supported but disabled in config
			fmt.Print("Graphical Board (Unicode) is supported but disabled in your config. Enable it now? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.ToLower(strings.TrimSpace(input))

			if input == "y" || input == "yes" {
				t := true
				data.Config.Unicode = &t
				useUnicode = true
				configChanged = true
			} else {
				useUnicode = false
			}
		} else {
			useUnicode = *data.Config.Unicode
		}

		if unicodeFlag {
			useUnicode = true
		}

		if configChanged {
			if err := store.Save(data); err != nil {
				fmt.Printf("Warning: Could not save config: %v\n", err)
			}
		}

		if ai {
			runAIGame(data, useUnicode)
			return
		}
		if multiplayer {
			runMultiplayerCreate(data, useUnicode)
			return
		}
		if joinCode != "" {
			runMultiplayerJoin(data, joinCode, useUnicode)
			return
		}

		runPuzzle(data, id, useUnicode)
	},
}

func runMultiplayerCreate(data *store.Data, useUnicode bool) {
	fmt.Println("Creating multiplayer game...")
	client := relay.NewClient("")
	code, err := client.CreateGame()
	if err != nil {
		fmt.Printf("Failed to create game: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nYour game code is: %s\n", code)
	fmt.Println("Share this with your opponent.")
	fmt.Println("Waiting for opponent to connect...")

	session := relay.NewSession(client)
	if err := session.Connect(code); err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	colorStr, err := session.WaitForOpponent()
	if err != nil {
		fmt.Printf("Error waiting for opponent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nOpponent joined! You are playing as %s.\n", colorStr)

	gameSession := puzzle.NewMultiplayerSession(session, os.Stdin, os.Stdout, colorStr)
	gameSession.Board.Unicode = useUnicode

	result, err := gameSession.Run()
	if err != nil {
		fmt.Printf("Error during game: %v\n", err)
		os.Exit(1)
	}

	if result == "Abandoned" {
		return
	}

	// Update stats (requires Phase 3.5 updates to models.go, assuming MultiplayerGames exists)
	// For now, we will add the struct fields in Phase 3.5 and uncomment this, or do it now.
	updateMultiplayerStats(data, gameSession.UserColor, gameSession.Board.Outcome())
}

func runMultiplayerJoin(data *store.Data, code string, useUnicode bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	fmt.Printf("Joining game %s...\n", code)
	
	client := relay.NewClient("")
	exists, err := client.ValidateCode(code)
	if err != nil {
		fmt.Printf("Error validating code: %v\n", err)
		os.Exit(1)
	}
	if !exists {
		fmt.Printf("No game found with code %s. Check the code and try again.\n", code)
		os.Exit(1)
	}

	session := relay.NewSession(client)
	if err := session.Connect(code); err != nil {
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "1008") {
			fmt.Printf("Game %s is already full.\n", code)
		} else {
			fmt.Printf("Connection error: %v\n", err)
		}
		os.Exit(1)
	}
	defer session.Close()

	colorStr, err := session.WaitForOpponent()
	if err != nil {
		fmt.Printf("Error waiting for game start: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nJoined successfully! You are playing as %s.\n", colorStr)
	
	gameSession := puzzle.NewMultiplayerSession(session, os.Stdin, os.Stdout, colorStr)
	gameSession.Board.Unicode = useUnicode

	result, err := gameSession.Run()
	if err != nil {
		fmt.Printf("Error during game: %v\n", err)
		os.Exit(1)
	}

	if result == "Abandoned" {
		return
	}

	updateMultiplayerStats(data, gameSession.UserColor, gameSession.Board.Outcome())
}

func updateMultiplayerStats(data *store.Data, userColor chess.Color, outcome chess.Outcome) {
	switch outcome {
	case chess.WhiteWon:
		if userColor == chess.White {
			data.MultiplayerGames.Wins++
		} else {
			data.MultiplayerGames.Losses++
		}
	case chess.BlackWon:
		if userColor == chess.Black {
			data.MultiplayerGames.Wins++
		} else {
			data.MultiplayerGames.Losses++
		}
	case chess.Draw:
		data.MultiplayerGames.Draws++
	}

	if err := store.Save(data); err != nil {
		fmt.Printf("Error saving stats: %v\n", err)
	} else {
		fmt.Println("Multiplayer game stats have been saved.")
	}
}

func runAIGame(data *store.Data, useUnicode bool) {
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

	// Color Selection
	var userColor chess.Color
	for {
		fmt.Print("Select your color (w/b): ")
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "w" || input == "white" {
			userColor = chess.White
			break
		} else if input == "b" || input == "black" {
			userColor = chess.Black
			break
		}
		fmt.Println("Invalid input. Enter 'w' for White or 'b' for Black.")
	}

	// Map 1-5 to Stockfish skill levels (0-20)
	skillLevel := (difficulty - 1) * 5
	fmt.Printf("Starting AI game as %s with Skill Level %d (Difficulty %d)...\n", userColor.Name(), skillLevel, difficulty)

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

	// Initial save for resume feature
	data.CurrentGame = &store.CurrentGame{
		FEN:        chess.NewGame().FEN(),
		Difficulty: difficulty,
		UserColor:  userColor.Name(),
	}
	if err := store.Save(data); err != nil {
		fmt.Printf("Warning: Could not save game for resume: %v\n", err)
	}

	session := puzzle.NewAISession(eng, os.Stdin, os.Stdout, userColor, difficulty)
	session.Board.Unicode = useUnicode
	
	// Setup OnMove callback
	session.OnMove = func(fen string) {
		data.CurrentGame.FEN = fen
		_ = store.Save(data)
	}

	result, err := session.Run()
	if err != nil {
		fmt.Printf("Error during game: %v\n", err)
		os.Exit(1)
	}

	if result == "Abandoned" {
		return
	}

	// Game finished, cleanup
	data.CurrentGame = nil

	switch session.Board.Outcome() {
	case chess.WhiteWon:
		if userColor == chess.White {
			data.AIGames.Wins++
		} else {
			data.AIGames.Losses++
		}
	case chess.BlackWon:
		if userColor == chess.Black {
			data.AIGames.Wins++
		} else {
			data.AIGames.Losses++
		}
	case chess.Draw:
		data.AIGames.Draws++
	}

	if err := store.Save(data); err != nil {
		fmt.Printf("Error saving stats: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("AI game stats have been saved.")
}

func runPuzzle(data *store.Data, id string, useUnicode bool) {
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
	session.Board.Unicode = useUnicode

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
	playCmd.Flags().BoolP("multiplayer", "m", false, "Start a multiplayer game and wait for an opponent.")
	playCmd.Flags().StringP("join", "j", "", "Join an existing multiplayer game using a code.")
	playCmd.Flags().Bool("unicode", false, "Force Graphical Board (Unicode) mode for board rendering.")
}
