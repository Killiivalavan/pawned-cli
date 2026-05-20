package relay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const DefaultRelayURL = "https://chesshell-relay.kavalavan04.workers.dev"

// Client handles REST and WebSocket connections to the multiplayer relay server.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new relay client.
func NewClient(relayURL string) *Client {
	if relayURL == "" {
		relayURL = DefaultRelayURL
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    strings.TrimSuffix(relayURL, "/"),
	}
}

// CreateGame creates a new multiplayer game room on the relay server.
func (c *Client) CreateGame() (string, error) {
	url := fmt.Sprintf("%s/api/game", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("invalid response: %w", err)
	}
	return result.Code, nil
}

// ValidateCode checks if a given game code exists.
func (c *Client) ValidateCode(code string) (bool, error) {
	url := fmt.Sprintf("%s/api/game/%s", c.baseURL, code)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("invalid response: %w", err)
	}
	return result.Exists, nil
}

// ConnectWebSocket upgrades the connection to a WebSocket for the game room.
func (c *Client) ConnectWebSocket(code string) (*websocket.Conn, error) {
	wsURL := strings.Replace(c.baseURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = fmt.Sprintf("%s/api/ws/%s", wsURL, code)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket connection failed: %w", err)
	}
	return conn, nil
}
