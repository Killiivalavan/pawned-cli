package api

// LichessPuzzle represents the full puzzle response from the Lichess API.
type LichessPuzzle struct {
	Game   Game   `json:"game"`
	Puzzle Puzzle `json:"puzzle"`
}

// Game contains information about the game from which the puzzle was derived.
type Game struct {
	ID      string   `json:"id"`
	PGN     string   `json:"pgn"`
	Rated   bool     `json:"rated"`
	Players []Player `json:"players"`
}

// Player represents a player in the game.
type Player struct {
	Name   string `json:"name"`
	Color  string `json:"color"`
	Rating int    `json:"rating"`
}

// Puzzle contains the puzzle-specific information.
type Puzzle struct {
	ID       string   `json:"id"`
	Rating   int      `json:"rating"`
	Plays    int      `json:"plays"`
	Solution []string `json:"solution"`
	Themes   []string `json:"themes"`
}
