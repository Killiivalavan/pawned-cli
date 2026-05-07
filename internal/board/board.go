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
	game    *chess.Game
	Unicode bool
	Flipped bool // If true, render from Black's perspective
}

// NewGame creates a brand new board with the standard starting position for a full game.
func NewGame() *Board {
	return &Board{game: chess.NewGame(), Unicode: IsUnicodeSupported(), Flipped: false}
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

	return &Board{game: game, Unicode: IsUnicodeSupported(), Flipped: false}, nil
}

// NewFromFEN creates a new board state from a FEN string.
func NewFromFEN(fen string) (*Board, error) {
	f, err := chess.FEN(fen)
	if err != nil {
		return nil, fmt.Errorf("invalid FEN: %w", err)
	}
	game := chess.NewGame(f)
	return &Board{game: game, Unicode: IsUnicodeSupported(), Flipped: false}, nil
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

// FENColor returns the color of the player whose turn it is to move.
func (b *Board) FENColor() chess.Color {
	return b.game.Position().Turn()
}

// Render draws the board to the provided writer (e.g., os.Stdout).
// It follows the formatting and coloring rules specified in the PRD.
func (b *Board) Render(w io.Writer) {
	if b.Unicode {
		b.renderUnicode(w)
		return
	}
	b.renderASCII(w)
}

func (b *Board) renderUnicode(w io.Writer) {
	coordColor := color.New(color.FgHiBlack)

	// Top coordinates
	fmt.Fprint(w, "  ")
	if b.Flipped {
		coordColor.Fprint(w, "h g f e d c b a\n")
	} else {
		coordColor.Fprint(w, "a b c d e f g h\n")
	}

	boardMap := b.game.Position().Board().SquareMap()

	ranks := []int{7, 6, 5, 4, 3, 2, 1, 0}
	files := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if b.Flipped {
		ranks = []int{0, 1, 2, 3, 4, 5, 6, 7}
		files = []int{7, 6, 5, 4, 3, 2, 1, 0}
	}

	for _, i := range ranks {
		rank := chess.Rank(i)
		coordColor.Fprintf(w, "%d ", rank+1)

		for _, j := range files {
			file := chess.File(j)
			sq := chess.NewSquare(file, rank)
			piece, exists := boardMap[sq]

			sqBg := color.BgWhite
			if (int(rank)+int(file))%2 != 0 {
				sqBg = color.BgYellow
			}

			if !exists {
				color.New(sqBg).Fprint(w, "  ")
			} else {
				pieceStr := pieceToUnicode(piece)
				// Using FgBlack for all Unicode pieces (outline for white, filled for black)
				// ensures maximum consistency across different background colors.
				color.New(sqBg, color.FgBlack, color.Bold).Fprintf(w, "%s ", pieceStr)
			}
		}

		coordColor.Fprintf(w, " %d\n", rank+1)
	}

	// Bottom coordinates
	fmt.Fprint(w, "  ")
	if b.Flipped {
		coordColor.Fprint(w, "h g f e d c b a\n")
	} else {
		coordColor.Fprint(w, "a b c d e f g h\n")
	}

	turn := b.game.Position().Turn()
	fmt.Fprintf(w, "\n  %s to move\n", turn.Name())
}

func (b *Board) renderASCII(w io.Writer) {
	// Custom color settings for the board.
	coordColor := color.New(color.FgHiBlack) // Dim color for coords
	whitePieceColor := color.New(color.FgWhite, color.Bold)
	blackPieceColor := color.New(color.FgYellow, color.Bold)

	// Top coordinates
	fmt.Fprint(w, "  ")
	if b.Flipped {
		coordColor.Fprint(w, "h g f e d c b a\n")
	} else {
		coordColor.Fprint(w, "a b c d e f g h\n")
	}

	// Get board state (square -> piece)
	boardMap := b.game.Position().Board().SquareMap()

	ranks := []int{7, 6, 5, 4, 3, 2, 1, 0}
	files := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if b.Flipped {
		ranks = []int{0, 1, 2, 3, 4, 5, 6, 7}
		files = []int{7, 6, 5, 4, 3, 2, 1, 0}
	}

	// Iterate over ranks
	for _, i := range ranks {
		rank := chess.Rank(i)
		coordColor.Fprintf(w, "%d ", rank+1) // Print rank number

		// Iterate over files
		for _, j := range files {
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
	if b.Flipped {
		coordColor.Fprint(w, "h g f e d c b a\n")
	} else {
		coordColor.Fprint(w, "a b c d e f g h\n")
	}

	// Add whose turn it is
	turn := b.game.Position().Turn()
	fmt.Fprintf(w, "\n  %s to move\n", turn.Name())
}

func pieceToUnicode(p chess.Piece) string {
	if p.Color() == chess.White {
		switch p.Type() {
		case chess.King:
			return "♔"
		case chess.Queen:
			return "♕"
		case chess.Rook:
			return "♖"
		case chess.Bishop:
			return "♗"
		case chess.Knight:
			return "♘"
		case chess.Pawn:
			return "♙"
		}
	} else {
		switch p.Type() {
		case chess.King:
			return "♚"
		case chess.Queen:
			return "♛"
		case chess.Rook:
			return "♜"
		case chess.Bishop:
			return "♝"
		case chess.Knight:
			return "♞"
		case chess.Pawn:
			return "♟"
		}
	}
	return ""
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
