package board

import (
	"fmt"
	"io"
	"strings"

	"github.com/corentings/chess"
	"github.com/fatih/color"
)

// Board represents the state of a chess puzzle.
// It wraps the underlying chess engine and provides a high-level API.
type Board struct {
	game *chess.Game
}

// NewGame creates a brand new board with the standard starting position for a full game.
func NewGame() *Board {
	return &Board{game: chess.NewGame()}
}

// NewFromPGN creates a new board state from a space-separated string of moves.
// The Lichess puzzle API provides the game's moves as a simple space-separated
// string in SAN (Standard Algebraic Notation), not a full PGN with headers.
func NewFromPGN(pgn string) (*Board, error) {
	game := chess.NewGame()

	// The Lichess API returns moves separated by spaces
	moves := strings.Split(strings.TrimSpace(pgn), " ")

	for _, m := range moves {
		if m == "" {
			continue
		}
		if err := game.MoveStr(m); err != nil {
			return nil, fmt.Errorf("failed to apply move '%s' from PGN: %w", m, err)
		}
	}

	return &Board{game: game}, nil
}

// Move validates and applies a move in UCI notation.
func (b *Board) Move(uci string) error {
	move, err := chess.UCINotation{}.Decode(b.game.Position(), uci)
	if err != nil {
		return fmt.Errorf("invalid move format: %w", err)
	}
	if err := b.game.Move(move); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}
	return nil
}

// FEN returns the current FEN string of the board position.
func (b *Board) FEN() string {
	return b.game.FEN()
}

// IsGameOver returns true if the game has ended (checkmate, stalemate, draw, etc).
func (b *Board) IsGameOver() bool {
	return b.game.Outcome() != chess.NoOutcome
}

// Outcome returns the raw outcome of the game (WhiteWon, BlackWon, Draw, NoOutcome).
func (b *Board) Outcome() chess.Outcome {
	return b.game.Outcome()
}

// Result returns a human-readable result of the game.
func (b *Board) Result() string {
	if b.game.Outcome() == chess.NoOutcome {
		return "In progress"
	}
	return fmt.Sprintf("%s by %s", b.game.Outcome().String(), b.game.Method().String())
}

// Render draws the board to the provided writer (e.g., os.Stdout).
// It follows the formatting and coloring rules specified in the PRD.
func (b *Board) Render(w io.Writer) {
	// Custom color settings for the board.
	coordColor := color.New(color.FgHiBlack) // Dim color for coords
	whitePieceColor := color.New(color.FgWhite, color.Bold)
	blackPieceColor := color.New(color.FgYellow, color.Bold)

	// Top coordinates
	fmt.Fprint(w, "  ")
	coordColor.Fprint(w, "a b c d e f g h\n")

	// Get board state (square -> piece)
	boardMap := b.game.Position().Board().SquareMap()

	// Iterate over ranks (8 to 1)
	for i := 7; i >= 0; i-- {
		rank := chess.Rank(i)
		coordColor.Fprintf(w, "%d ", rank+1) // Print rank number

		// Iterate over files (a to h)
		for j := 0; j < 8; j++ {
			file := chess.File(j)
			sq := chess.NewSquare(file, rank)
			piece, exists := boardMap[sq]

			if !exists {
				fmt.Fprint(w, ". ")
				continue
			}

			pieceStr := pieceToASCII(piece)
			if piece.Color() == chess.White {
				whitePieceColor.Fprint(w, pieceStr)
			} else {
				blackPieceColor.Fprint(w, pieceStr)
			}
			fmt.Fprint(w, " ")
		}

		coordColor.Fprintf(w, " %d\n", rank+1) // Print rank number again
	}

	// Bottom coordinates
	fmt.Fprint(w, "  ")
	coordColor.Fprint(w, "a b c d e f g h\n")

	// Add whose turn it is
	turn := b.game.Position().Turn()
	fmt.Fprintf(w, "\n  %s to move\n", turn.Name())
}

// pieceToASCII converts a chess.Piece to its PRD-specified ASCII representation.
func pieceToASCII(p chess.Piece) string {
	char := ""
	switch p.Type() {
	case chess.King:
		char = "k"
	case chess.Queen:
		char = "q"
	case chess.Rook:
		char = "r"
	case chess.Bishop:
		char = "b"
	case chess.Knight:
		char = "n"
	case chess.Pawn:
		char = "p"
	}
	if p.Color() == chess.White {
		return strings.ToUpper(char)
	}
	return char
}
