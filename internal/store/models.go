package store

import "time"

// Data represents the entire structure of the local data file (e.g., data.json).
type Data struct {
	Version int           `json:"version"`
	Stats   Stats         `json:"stats"`
	AIGames AIGames       `json:"aiGames"`
	History []HistoryItem `json:"history"`
}

// AIGames holds the user's statistics against the local AI engine.
type AIGames struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// Stats holds the user's aggregate puzzle-solving statistics.
type Stats struct {
	TotalAttempted int       `json:"totalAttempted"`
	TotalSolved    int       `json:"totalSolved"`
	CurrentStreak  int       `json:"currentStreak"`
	BestStreak     int       `json:"bestStreak"`
	FirstPlayedAt  *time.Time `json:"firstPlayedAt,omitempty"`
	LastPlayedAt   *time.Time `json:"lastPlayedAt,omitempty"`
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
