package puzzle

import (
	"bufio"
	"chesshell-cli/internal/board"
	"chesshell-cli/internal/relay"
	"fmt"
	"io"
	"strings"

	"github.com/corentings/chess"
	"github.com/fatih/color"
)

// MultiplayerSession represents an interactive multiplayer game over WebSocket.
type MultiplayerSession struct {
	Board     *board.Board
	Relay     *relay.Session
	UserColor chess.Color
	in        io.Reader
	out       io.Writer
}

// NewMultiplayerSession creates a new multiplayer session.
func NewMultiplayerSession(relaySession *relay.Session, in io.Reader, out io.Writer, userColor string) *MultiplayerSession {
	c := chess.White
	if userColor == "black" {
		c = chess.Black
	}

	b := board.NewGame()
	b.Flipped = (c == chess.Black) // Flip board for Black player

	return &MultiplayerSession{
		Board:     b,
		Relay:     relaySession,
		UserColor: c,
		in:        in,
		out:       out,
	}
}

// Run starts the interactive session loop.
func (s *MultiplayerSession) Run() (string, error) {
	reader := bufio.NewReader(s.in)

	for {
		fmt.Fprintln(s.out)
		s.Board.Render(s.out)
		fmt.Fprintln(s.out)

		if s.Board.IsGameOver() {
			break
		}

		if s.Board.FENColor() == s.UserColor {
			// User's turn
			fmt.Fprint(s.out, "Your move (UCI) -> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return "Error", err
			}

			input = strings.TrimSpace(input)

			if input == "q" || input == "quit" || input == "exit" {
				fmt.Fprintln(s.out, "Game abandoned.")
				// Could notify relay, but closing WS does that
				return "Abandoned", nil
			}

			if err := s.Board.Move(input); err != nil {
				color.Red(err.Error())
				continue
			}

			// Send move
			payload := relay.MovePayload{
				FEN:      s.Board.FEN(),
				UCI:      input,
				GameOver: s.Board.IsGameOver(),
				Result:   s.Board.Result(),
			}
			if err := s.Relay.SendMove(payload); err != nil {
				return "Error", fmt.Errorf("failed to send move: %w", err)
			}
		} else {
			// Opponent's turn
			fmt.Fprintln(s.out, "Waiting for opponent...")
			
			payload, err := s.Relay.ListenMove()
			if err != nil {
				if err.Error() == "opponent disconnected" {
					color.Yellow("\nOpponent disconnected. Game abandoned.")
					return "Abandoned", nil
				}
				return "Error", fmt.Errorf("failed to receive move: %w", err)
			}

			if err := s.Board.Move(payload.UCI); err != nil {
				// The move was invalid according to our board, but relay says they did it.
				// This implies a desync.
				return "Error", fmt.Errorf("desync: opponent sent invalid move %s", payload.UCI)
			}
		}
	}

	outcome := s.Board.Outcome()
	resStr := s.Board.Result()

	if outcome == chess.Draw {
		color.Yellow("Game Over: Draw (1/2-1/2)")
	} else if outcome == chess.WhiteWon {
		if s.UserColor == chess.White {
			color.Green("Game Over: White wins! (1-0)")
		} else {
			color.Red("Game Over: White wins. (1-0)")
		}
	} else if outcome == chess.BlackWon {
		if s.UserColor == chess.Black {
			color.Green("Game Over: Black wins! (0-1)")
		} else {
			color.Red("Game Over: Black wins. (0-1)")
		}
	}

	return resStr, nil
}
