package cmd

import (
	"chesshell-cli/internal/board"
	"chesshell-cli/internal/engine"
	"chesshell-cli/internal/puzzle"
	"chesshell-cli/internal/store"
	"fmt"
	"os"
	"strings"

	"github.com/corentings/chess"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the last unfinished AI game.",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := store.Load()
		if err != nil && err != store.ErrCorruptedFile {
			fmt.Printf("Error loading data: %v\n", err)
			os.Exit(1)
		}

		if data.CurrentGame == nil {
			fmt.Println("No unfinished game found to resume.")
			return
		}

		// Resolve Unicode
		useUnicode := board.IsUnicodeSupported()
		if data.Config.Unicode != nil {
			useUnicode = *data.Config.Unicode
		}

		// Setup color
		userColor := chess.White
		if strings.ToLower(data.CurrentGame.UserColor) == "black" {
			userColor = chess.Black
		}

		fmt.Printf("Resuming game (Difficulty: %d, Color: %s)...\n", data.CurrentGame.Difficulty, userColor.Name())

		// Start Engine
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

		skillLevel := (data.CurrentGame.Difficulty - 1) * 5
		if err := eng.Configure(skillLevel); err != nil {
			fmt.Printf("Error configuring engine: %v\n", err)
			os.Exit(1)
		}

		// Create session
		session := puzzle.NewAISession(eng, os.Stdin, os.Stdout, userColor, data.CurrentGame.Difficulty)
		session.Board.Unicode = useUnicode
		
		// Load FEN
		b, err := board.NewFromFEN(data.CurrentGame.FEN)
		if err != nil {
			fmt.Printf("Error loading game state: %v\n", err)
			os.Exit(1)
		}
		session.Board = b
		session.Board.Unicode = useUnicode
		if userColor == chess.Black {
			session.Board.Flipped = true
		}

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
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
