package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const lichessBaseURL = "https://lichess.org"

// Custom error types for more specific error handling.
var (
	ErrPuzzleNotFound = errors.New("puzzle not found")
	ErrLichessServer  = errors.New("lichess API is unavailable")
	ErrRateLimited    = errors.New("rate limited by Lichess API")
	ErrNetwork        = errors.New("could not connect to Lichess API, check your internet connection")
)

// Client manages communication with the Lichess API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Lichess API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Generous timeout for network requests.
		},
	}
}

// FetchDaily fetches the daily puzzle from Lichess.
func (c *Client) FetchDaily() (*LichessPuzzle, error) {
	url := fmt.Sprintf("%s/api/puzzle/daily", lichessBaseURL)
	return c.fetchAndDecode(url)
}

// FetchByID fetches a puzzle by its ID from Lichess.
func (c *Client) FetchByID(id string) (*LichessPuzzle, error) {
	url := fmt.Sprintf("%s/api/puzzle/%s", lichessBaseURL, id)
	return c.fetchAndDecode(url)
}

// fetchAndDecode is a helper function that handles making the HTTP request,
// checking for errors, and decoding the JSON response into a LichessPuzzle struct.
func (c *Client) fetchAndDecode(url string) (*LichessPuzzle, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// This is an internal error with creating the request, not a network one.
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// The Lichess API prefers a User-Agent.
	req.Header.Set("User-Agent", "chesshell-cli/v1")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrNetwork
	}
	defer res.Body.Close()

	// Handle specific status codes as per the PRD.
	if res.StatusCode >= 500 {
		return nil, ErrLichessServer
	}

	switch res.StatusCode {
	case http.StatusOK:
		// Success case.
		var puzzle LichessPuzzle
		if err := json.NewDecoder(res.Body).Decode(&puzzle); err != nil {
			return nil, fmt.Errorf("failed to decode lichess response: %w", err)
		}
		return &puzzle, nil

	case http.StatusNotFound:
		return nil, ErrPuzzleNotFound

	case http.StatusTooManyRequests:
		return nil, ErrRateLimited

	default:
		return nil, fmt.Errorf("lichess API returned an unexpected status: %s", res.Status)
	}
}
