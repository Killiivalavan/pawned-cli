package puzzle

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"pawned-cli/internal/api"
	"pawned-cli/internal/board"
)

// Session manages the state and flow of a single puzzle-solving session.
type Session struct {
	Puzzle      *api.LichessPuzzle
	Board       *board.Board
	in          io.Reader
	out         io.Writer
	solutionIdx int
}

// NewSession creates a new puzzle-solving session.
func NewSession(puzzle *api.LichessPuzzle, in io.Reader, out io.Writer) (*Session, error) {
	b, err := board.NewFromPGN(puzzle.Game.PGN)
	if err != nil {
		return nil, fmt.Errorf("failed to create board for session: %w", err)
	}
	return &Session{
		Puzzle:      puzzle,
		Board:       b,
		in:          in,
		out:         out,
		solutionIdx: 0,
	}, nil
}

// Run starts the interactive puzzle-solving loop.
// It returns whether the puzzle was solved, the number of incorrect attempts, and any error.
func (s *Session) Run() (solved bool, attempts int, err error) {
	s.displayPuzzleInfo()
	s.Board.Render(s.out)

	scanner := bufio.NewScanner(s.in)

	for {
		promptColor := color.New(color.FgCyan)
		promptColor.Fprint(s.out, "\nYour move (UCI) -> ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, attempts, fmt.Errorf("error reading input: %w", err)
			}
			// EOF (Ctrl+D)
			return false, attempts, nil
		}

		input := strings.TrimSpace(scanner.Text())

		// Handle commands
		switch strings.ToLower(input) {
		case "q", "quit":
			fmt.Fprintln(s.out, "\nPuzzle abandoned.")
			return false, attempts, nil
		case "h", "hint":
			s.giveHint()
			continue
		}

		// It's a move, let's process it.
		if err := s.processMove(input); err != nil {
			errorColor := color.New(color.FgRed)
			errorColor.Fprintf(s.out, "✗ %v. Try again.\n", err)
			if err.Error() == "wrong move" {
				attempts++
			}
			continue
		}

		// Check if the puzzle is fully solved
		if s.solutionIdx >= len(s.Puzzle.Puzzle.Solution) {
			successColor := color.New(color.Bold, color.FgGreen)
			successColor.Fprintln(s.out, "\n✓ Puzzle solved! 🎉")
			return true, attempts, nil
		}

		// Apply opponent's move automatically
		if err := s.applyOpponentMove(); err != nil {
			return false, attempts, fmt.Errorf("error applying opponent move: %w", err)
		}

		s.Board.Render(s.out)
	}
}

// processMove checks if the user's move is correct.
func (s *Session) processMove(uci string) error {
	correctMove := s.Puzzle.Puzzle.Solution[s.solutionIdx]
	if uci != correctMove {
		return fmt.Errorf("wrong move")
	}

	if err := s.Board.Move(uci); err != nil {
		return fmt.Errorf("invalid move format '%s'", uci)
	}

	successColor := color.New(color.FgGreen)
	successColor.Fprintln(s.out, "✓ Correct!")

	s.solutionIdx++
	return nil
}

// applyOpponentMove applies the next move from the solution array, which is the opponent's response.
func (s *Session) applyOpponentMove() error {
	if s.solutionIdx >= len(s.Puzzle.Puzzle.Solution) {
		return nil // No more moves left
	}
	opponentMove := s.Puzzle.Puzzle.Solution[s.solutionIdx]
	if err := s.Board.Move(opponentMove); err != nil {
		return fmt.Errorf("internal error applying opponent's move %s: %w", opponentMove, err)
	}

	fmt.Fprintf(s.out, "Opponent plays %s\n", opponentMove)
	s.solutionIdx++
	return nil
}

// giveHint provides a hint to the user as specified in the PRD.
func (s *Session) giveHint() {
	if s.solutionIdx >= len(s.Puzzle.Puzzle.Solution) {
		return // Puzzle already solved
	}
	correctMove := s.Puzzle.Puzzle.Solution[s.solutionIdx]
	fromSquare := correctMove[0:2]

	hintColor := color.New(color.FgMagenta)
	hintColor.Fprintf(s.out, "Hint: the move starts from square %s.\n", fromSquare)
}

// displayPuzzleInfo shows metadata about the current puzzle.
func (s *Session) displayPuzzleInfo() {
	infoColor := color.New(color.FgCyan)
	infoColor.Fprintf(s.out, "Puzzle #%s | Rating: %d | Themes: %s\n",
		s.Puzzle.Puzzle.ID,
		s.Puzzle.Puzzle.Rating,
		strings.Join(s.Puzzle.Puzzle.Themes, ", "),
	)
}
