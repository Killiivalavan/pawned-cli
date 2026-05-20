package store

import "time"

// Data represents the entire structure of the local data file (e.g., data.json).
type Data struct {
	Version          int              `json:"version"`
	Config           Config           `json:"config"`
	Stats            Stats            `json:"stats"`
	AIGames          AIGames          `json:"aiGames"`
	MultiplayerGames MultiplayerGames `json:"multiplayerGames"`
	CurrentGame      *CurrentGame     `json:"currentGame,omitempty"`
	History          []HistoryItem    `json:"history"`
}

// CurrentGame stores the state of an unfinished AI game.
type CurrentGame struct {
	FEN        string `json:"fen"`
	Difficulty int    `json:"difficulty"`
	UserColor  string `json:"userColor"` // "white" or "black"
}

// Config holds user preferences.
type Config struct {
	Unicode *bool `json:"unicode,omitempty"`
}

// AIGames holds the user's statistics against the local AI engine.
type AIGames struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// MultiplayerGames holds the user's statistics for multiplayer matches.
type MultiplayerGames struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// Stats holds the user's aggregate puzzle-solving statistics.
type Stats struct {
	TotalAttempted         int        `json:"totalAttempted"`
	TotalSolved            int        `json:"totalSolved"`
	CurrentStreak          int        `json:"currentStreak"`
	BestStreak             int        `json:"bestStreak"`
	FirstPlayedAt          *time.Time `json:"firstPlayedAt,omitempty"`
	LastPlayedAt           *time.Time `json:"lastPlayedAt,omitempty"`
	LastUpdateCheck        *time.Time `json:"lastUpdateCheck,omitempty"`
	LastUpdateAcknowledged string     `json:"lastUpdateAcknowledged,omitempty"`
}

// HistoryItem represents a single recorded puzzle attempt.
type HistoryItem struct {
	PuzzleID    string    `json:"puzzleId"`
	Rating      int       `json:"rating"`
	Themes      []string  `json:"themes"`
	AttemptedAt time.Time `json:"attemptedAt"`
	Solved      bool      `json:"solved"`
	Attempts    int       `json:"attempts"` // Number of wrong moves before solving
}
