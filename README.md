# chesshell: Chess Drill CLI

A minimal, elegant command-line tool that fetches chess puzzles from Lichess and lets users solve them directly in the terminal.

Built for developers and terminal power users who want a quick chess puzzle fix without leaving their command line. No account required, no configuration needed.

## Features

- **Zero auth:** No Lichess account required.
- **Zero config:** Works immediately after install.
- **Local-first:** All user data (stats, history) lives on your machine.
- **Single binary:** Fast, self-contained executable.
- **Local AI:** Play full games vs Stockfish (downloads on first use).

## Installation

### macOS and Linux
Install `chesshell` using our quick installer script:

```bash
curl -fsSL https://raw.githubusercontent.com/Killiivalavan/chesshell-cli/main/install.sh | bash
```

### Windows
Open PowerShell as an Administrator and run the following command to automatically download and install `chesshell` and add it to your PATH:

```powershell
iwr -useb https://raw.githubusercontent.com/Killiivalavan/chesshell-cli/main/install.ps1 | iex
```

## Usage

### Play the Daily Puzzle

Fetch today's featured puzzle from Lichess and start an interactive session:

```bash
chesshell play
```

### Play a Specific Puzzle

If you know the Lichess ID of a puzzle, you can play it directly:

```bash
chesshell play --id pId3s
```

### Play against Local AI

Play a full chess game directly in your terminal against the world-class Stockfish engine! You'll be prompted to select your difficulty from Beginner to Expert.

*(Note: The first time you run this command, `chesshell` will seamlessly download the ~50MB Stockfish binary for your OS. It runs entirely locally on your machine.)*

```bash
chesshell play --ai
```
Difficulty is selected interactively (1–5), and `q`/`quit` abandons the game. (`hint` is not available in AI games.)

### View Your Stats

Check your progress, including total puzzles solved, your accuracy, and your record against the AI:

```bash
chesshell stats
```

### View History

See a table of your most recently attempted puzzles:

```bash
chesshell history
```

You can limit the number of entries shown using the `--limit` flag:

```bash
chesshell history --limit 5
```

## How It Works

`chesshell` interacts with the public, unauthenticated endpoints of the [Lichess API](https://lichess.org/api) to fetch puzzle data and PGN move sequences. 

The application reconstructs the board state locally and validates your inputs against the puzzle's solution. All of your personal statistics and history are saved locally on your machine in a simple `data.json` file.

For AI games, `chesshell` runs Stockfish locally via the UCI protocol. If Stockfish is not found on your system `PATH`, `chesshell` caches a downloaded copy under your OS config directory.

## Input Format

When playing a puzzle, enter your moves using **UCI (Universal Chess Interface) notation**.

- Example: `e2e4` (moves a piece from e2 to e4)
- Captures: `f3d5` (moves a piece from f3 to capture on d5)
- Promotion: `e7e8q` (promotes a pawn on e8 to a queen)

If you get stuck, type `hint` (or `h`) to see which square the correct piece moves from. Type `quit` (or `q`) to abandon the puzzle.

## Contributing

Contributions are welcome! Please ensure that your pull requests adhere to the core principles of the project: simple, elegant, efficient, and quick. Avoid scope creep.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
