package puzzle

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"chesshell-cli/internal/board"
	"chesshell-cli/internal/engine"
	"github.com/corentings/chess"
	"github.com/fatih/color"
)

// AISession manages a full game session against the Stockfish AI.
type AISession struct {
	Board      *board.Board
	Engine     *engine.Process
	in         io.Reader
	out        io.Writer
	UserColor  chess.Color
	Difficulty int
	OnMove     func(fen string)
}

// NewAISession initializes a new game against the AI.
func NewAISession(eng *engine.Process, in io.Reader, out io.Writer, userColor chess.Color, difficulty int) *AISession {
	b := board.NewGame()
	if userColor == chess.Black {
		b.Flipped = true
	}
	return &AISession{
		Board:      b,
		Engine:     eng,
		in:         in,
		out:        out,
		UserColor:  userColor,
		Difficulty: difficulty,
	}
}

// Run starts the interactive AI game loop.
// Returns the game result (e.g., "1/2-1/2 by Repetition") and an error if one occurred.
func (s *AISession) Run() (string, error) {
	fmt.Fprintf(s.out, "Game started against AI. You are %s.\n", s.UserColor.Name())
	s.Board.Render(s.out)

	scanner := bufio.NewScanner(s.in)

	for {
		// --- 1. Current Turn ---
		turn := s.Board.FENColor()

		if turn == s.UserColor {
			// User's Turn
			promptColor := color.New(color.FgCyan)
			promptColor.Fprint(s.out, "\nYour move (UCI) -> ")

			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return "", fmt.Errorf("error reading input: %w", err)
				}
				return "Abandoned", nil // EOF
			}

			input := strings.TrimSpace(scanner.Text())

			switch strings.ToLower(input) {
			case "q", "quit":
				fmt.Fprintln(s.out, "\nGame abandoned.")
				return "Abandoned", nil
			case "h", "hint":
				fmt.Fprintln(s.out, "Hints are not available in AI games.")
				continue
			}

			// Process User Move
			if err := s.Board.Move(input); err != nil {
				errorColor := color.New(color.FgRed)
				errorColor.Fprintf(s.out, "✗ %v. Try again.\n", err)
				continue
			}

			if s.OnMove != nil {
				s.OnMove(s.Board.FEN())
			}
		} else {
			// Engine's Turn
			fmt.Fprintln(s.out, "\nAI is thinking...")
			bestMove, err := s.Engine.GetBestMove(s.Board.FEN())
			if err != nil {
				return "", fmt.Errorf("engine error: %w", err)
			}

			if err := s.Board.Move(bestMove); err != nil {
				return "", fmt.Errorf("engine returned an invalid move '%s': %w", bestMove, err)
			}

			fmt.Fprintf(s.out, "AI plays %s\n", bestMove)

			if s.OnMove != nil {
				s.OnMove(s.Board.FEN())
			}
		}

		s.Board.Render(s.out)

		if s.Board.IsGameOver() {
			result := s.Board.Result()
			s.printGameOver(result)
			return result, nil
		}
	}
}

func (s *AISession) printGameOver(result string) {
	successColor := color.New(color.Bold, color.FgGreen)
	successColor.Fprintf(s.out, "\nGame Over! Result: %s\n", result)
}
