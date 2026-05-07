# chesshell: Technical Documentation

A minimal, high-performance chess tool for the terminal. `chesshell` is designed to provide a distraction-free environment for chess puzzles and AI games, prioritizing speed, local-first data persistence, and zero-configuration setup.

## Architecture Overview

The project is structured into several modular packages within the `internal/` directory, following standard Go project layouts.

### `internal/board`
The core state machine. It wraps the `github.com/corentings/chess` library to manage game logic.
- **Rendering**: Implements two distinct rendering pipelines:
  - **Graphical (Unicode)**: Uses Unicode chess glyphs, ANSI background colors for square parity, and high-contrast foreground coloring.
  - **Classic (ASCII)**: A fallback renderer for legacy terminals using alphanumeric characters.
- **Detection**: Automatically detects terminal capabilities (`UTF-8` support) via environment variable analysis (`LANG`, `TERM`, etc.).

### `internal/engine`
Manages communication with the Stockfish chess engine.
- **UCI Protocol**: Implements the Universal Chess Interface (UCI) protocol via stdin/stdout pipes.
- **Lifecycle**: Handles the automatic downloading, caching, and execution of platform-specific Stockfish binaries if they are not found in the system `PATH`.

### `internal/api`
A thin client for the Lichess.org API.
- **Zero Auth**: Uses public, unauthenticated endpoints to fetch daily puzzles or specific puzzle IDs.
- **Fault Tolerance**: Implements custom error handling for rate-limiting (429) and network timeouts.

### `internal/store`
The persistence layer.
- **Local-First**: Data is stored in a single `data.json` file in the user's config directory (e.g., `~/.config/chesshell/`).
- **State Management**: Tracks aggregate statistics, puzzle history, and the **Resume** state for unfinished AI games.

---

## Project Structure

```text
├── cmd/                # Cobra command definitions (entry points)
├── internal/
│   ├── api/            # Lichess API integration
│   ├── board/          # Board logic and ASCII/Unicode rendering
│   ├── engine/         # Stockfish process management & binary downloader
│   ├── puzzle/         # Session managers for Puzzles and AI Games
│   └── store/          # JSON persistence and data models
├── main.go             # Application entry point
└── data.json           # (Created at runtime) Local state & config
```

---

## Development & Contribution

### Prerequisites
- **Go**: 1.26.2 or higher.
- **Git**: For version control.
- **Stockfish**: (Optional) The tool will download it automatically, but you can use a system-installed version if available in your `PATH`.

### Building from Source
```bash
git clone https://github.com/Killiivalavan/chesshell-cli.git
cd chesshell-cli
go build -o chesshell main.go
```

### Core Principles for Contributors
When contributing, please adhere to these architectural constraints:
1. **Minimalism**: Avoid adding dependencies. Prefer the Go standard library or high-quality, lightweight packages.
2. **Zero Config**: Features should work immediately. If a feature requires configuration, provide sensible defaults and an interactive setup.
3. **Local-First**: Never require a network connection for features that don't strictly need it (like AI games or stats).
4. **Non-Destructive Migrations**: When updating `data.json` schemas, ensure old versions are migrated without data loss.

### Code Style
- Run `go fmt ./...` before committing.
- Ensure all exported symbols have concise, descriptive comments.
- Keep the `cmd/` package thin; move business logic into `internal/`.

---

## Data Schema (`data.json`)
The application persists data using the following schema:
- `config`: User preferences (e.g., `unicode` mode).
- `stats`: Aggregate win/loss and solve counts.
- `aiGames`: Specific records for games against Stockfish.
- `currentGame`: Mid-game state (FEN and difficulty) to support the **Resume** feature.
- `history`: A rolling buffer of the last 200 puzzle attempts.

## License
This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
