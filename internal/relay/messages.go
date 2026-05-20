package relay

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	MsgCreateGame         MessageType = "create_game"
	MsgGameCreated        MessageType = "game_created"
	MsgJoinGame           MessageType = "join_game"
	MsgGameStart          MessageType = "game_start"
	MsgMove               MessageType = "move"
	MsgOpponentDisconnect MessageType = "opponent_disconnected"
	MsgOpponentReconnect  MessageType = "opponent_reconnected"
	MsgPing               MessageType = "ping"
	MsgPong               MessageType = "pong"
	MsgError              MessageType = "error"
)

// RelayMessage represents the standard message envelope.
type RelayMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// MovePayload represents the payload for a "move" message.
type MovePayload struct {
	FEN      string `json:"fen"`
	UCI      string `json:"uci"`
	GameOver bool   `json:"gameOver"`
	Result   string `json:"result"` // e.g., "1-0", "0-1", "1/2-1/2"
}

// GameStartPayload represents the payload for a "game_start" message.
type GameStartPayload struct {
	Color string `json:"color"` // "white" or "black"
}

// SendMessage sends a structured message over the WebSocket.
func SendMessage(conn *websocket.Conn, msgType MessageType, payload interface{}) error {
	msg := RelayMessage{Type: msgType}
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		msg.Payload = raw
	}
	return conn.WriteJSON(msg)
}

// ReadMessage reads and parses the next message from the WebSocket.
func ReadMessage(conn *websocket.Conn) (RelayMessage, error) {
	var msg RelayMessage
	err := conn.ReadJSON(&msg)
	return msg, err
}
