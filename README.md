# chesshell: Technical Documentation

A minimal, high-performance chess tool for the terminal. `chesshell` is designed to provide a distraction-free environment for chess puzzles and AI games, prioritizing speed, local-first data persistence, and zero-configuration setup.

## Architecture Overview

The project is structured into several modular packages within the `internal/` directory, following standard Go project layouts. The multiplayer backend is a serverless Cloudflare Worker.

### Go Application (`internal/`)
- **`internal/board`**: The core state machine. It wraps the `github.com/corentings/chess` library to manage game logic and implements both graphical (Unicode) and classic (ASCII) terminal rendering pipelines.
- **`internal/engine`**: Manages communication with the Stockfish chess engine via the UCI protocol. It handles the automatic downloading, caching, and execution of platform-specific Stockfish binaries.
- **`internal/api`**: A thin client for the Lichess.org API's public, unauthenticated endpoints to fetch chess puzzles.
- **`internal/store`**: The local-first persistence layer. All user data (stats, history, config, and game state) is stored in a single `data.json` file.
- **`internal/update`**: The self-update system. It checks for new releases on GitHub, finds the correct binary for the user's OS/architecture, and replaces the current executable.
- **`internal/relay`**: The client-side logic for the multiplayer feature. It communicates with the Cloudflare Worker via REST (to create/join games) and WebSockets (to exchange moves).

### Serverless Backend (`relay/`)
The `relay/` directory contains a Cloudflare Worker that acts as a simple, "dumb pipe" message broker for multiplayer games. It uses Durable Objects to manage game rooms and forward WebSocket messages between two clients without any server-side chess logic.

---

## Project Structure

```text
├── cmd/                # Cobra command definitions (entry points)
├── internal/
│   ├── api/            # Lichess API integration
│   ├── board/          # Board logic and ASCII/Unicode rendering
│   ├── engine/         # Stockfish process management & binary downloader
│   ├── puzzle/         # Session managers for Puzzles and AI Games
│   ├── relay/          # (v3) Multiplayer client for the relay server
│   ├── store/          # JSON persistence and data models
│   └── update/         # (v3) Self-update logic
├── relay/              # (v3) Cloudflare Worker for the multiplayer relay
├── main.go             # Application entry point
├── install.sh          # Installer script
└── data.json           # (Created at runtime) Local state & config
```

---

## Development & Contribution

### Prerequisites
- **Go**: 1.22 or higher.
- **Git**: For version control.
- **Node.js & Wrangler**: Required for developing the multiplayer relay server.
- **Stockfish**: (Optional) The tool will download it automatically, but you can use a system-installed version if available in your `PATH`.

### Building from Source
```bash
git clone https://github.com/Killiivalavan/chesshell-cli.git
cd chesshell-cli
go build -o chesshell main.go
```

### Testing the Full Stack (including Multiplayer)
To test all features, you need to run the Go application and the multiplayer relay server simultaneously.

1.  **Run the relay server:**
    ```bash
    cd relay/
    npm install
    npm run dev
    ```
    This will start a local server on `http://localhost:8787`.

2.  **Run the Go application:**
    In a separate terminal, you can now run `go run main.go play --friend` or use the compiled binary to test multiplayer functionality against the local relay.

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
- `multiplayerGames`: (v3) Records for multiplayer matches.
- `currentGame`: Mid-game state (FEN and difficulty) to support the **Resume** feature.
- `history`: A rolling buffer of the last 200 puzzle attempts.

## License
This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
