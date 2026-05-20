package relay

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

// Session manages an active multiplayer WebSocket connection.
type Session struct {
	client *Client
	code   string
	conn   *websocket.Conn
}

// NewSession creates a new relay session.
func NewSession(client *Client) *Session {
	return &Session{
		client: client,
	}
}

// Connect establishes the WebSocket connection to the given game room code.
func (s *Session) Connect(code string) error {
	s.code = code
	conn, err := s.client.ConnectWebSocket(code)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}

// WaitForOpponent blocks until a game_start message is received.
// Returns the assigned color ("white" or "black").
func (s *Session) WaitForOpponent() (string, error) {
	for {
		msg, err := ReadMessage(s.conn)
		if err != nil {
			return "", fmt.Errorf("error waiting for opponent: %w", err)
		}

		switch msg.Type {
		case MsgGameStart:
			var payload GameStartPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return "", fmt.Errorf("invalid game_start payload: %w", err)
			}
			return payload.Color, nil
		case MsgPing:
			SendMessage(s.conn, MsgPong, nil)
		}
	}
}

// SendMove sends a move to the opponent.
func (s *Session) SendMove(payload MovePayload) error {
	return SendMessage(s.conn, MsgMove, payload)
}

// ListenMove blocks until the opponent makes a move or an event occurs.
// It handles ping/pong automatically.
func (s *Session) ListenMove() (MovePayload, error) {
	for {
		msg, err := ReadMessage(s.conn)
		if err != nil {
			return MovePayload{}, err // Return error to trigger reconnect logic or exit
		}

		switch msg.Type {
		case MsgMove:
			var payload MovePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return MovePayload{}, fmt.Errorf("invalid move payload: %w", err)
			}
			return payload, nil
		case MsgOpponentDisconnect:
			return MovePayload{}, fmt.Errorf("opponent disconnected")
		case MsgPing:
			SendMessage(s.conn, MsgPong, nil)
		}
	}
}

// Close gracefully closes the WebSocket connection.
func (s *Session) Close() error {
	if s.conn != nil {
		err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		s.conn.Close()
		return err
	}
	return nil
}
