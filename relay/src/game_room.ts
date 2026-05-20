import { GAME_TIMEOUT_MS, parseMessage } from "./utils";
import { StrictRelayMessage } from "./types";

export class GameRoom {
  state: DurableObjectState;
  code: string | null = null;
  player1: WebSocket | null = null;
  player2: WebSocket | null = null;
  currentFEN: string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
  createdAt: number = Date.now();

  constructor(state: DurableObjectState, env: any) {
    this.state = state;
  }

  async fetch(request: Request): Promise<Response> {
    if (request.headers.get("Upgrade") !== "websocket") {
      return new Response("Expected Upgrade: websocket", { status: 426 });
    }

    const { 0: client, 1: server } = new WebSocketPair();
    server.accept();

    const url = new URL(request.url);
    const pathCode = url.pathname.split("/").pop() || "";
    
    if (!this.code) {
        this.code = pathCode;
    }

    if (!this.player1) {
      this.player1 = server;
      this.setupWebSocket(server, 1);
      // Wait to send game_created until we know they created it?
      // For now, if they are the first connection, they are the creator.
    } else if (!this.player2) {
      this.player2 = server;
      this.setupWebSocket(server, 2);
      
      // Both players connected, start the game
      this.sendTo(this.player1, { type: "game_start", payload: { color: "white" } });
      this.sendTo(this.player2, { type: "game_start", payload: { color: "black" } });
    } else {
      server.close(1008, "Game is full");
      return new Response(null, { status: 101, webSocket: client });
    }

    return new Response(null, { status: 101, webSocket: client });
  }

  setupWebSocket(ws: WebSocket, playerNum: 1 | 2) {
    ws.addEventListener("message", (event) => {
      const msg = parseMessage(event.data as string);
      if (!msg) return;

      if (msg.type === "move") {
        this.currentFEN = msg.payload.fen;
        const opponent = playerNum === 1 ? this.player2 : this.player1;
        if (opponent) {
          this.sendTo(opponent, msg);
        }
      } else if (msg.type === "ping") {
        this.sendTo(ws, { type: "pong" });
      }
    });

    ws.addEventListener("close", () => this.handleDisconnect(playerNum));
    ws.addEventListener("error", () => this.handleDisconnect(playerNum));
  }

  handleDisconnect(playerNum: 1 | 2) {
    if (playerNum === 1) {
      this.player1 = null;
      if (this.player2) {
        this.sendTo(this.player2, { type: "opponent_disconnected" });
      }
    } else {
      this.player2 = null;
      if (this.player1) {
        this.sendTo(this.player1, { type: "opponent_disconnected" });
      }
    }

    if (!this.player1 && !this.player2) {
      this.state.storage.setAlarm(Date.now() + GAME_TIMEOUT_MS);
    }
  }

  async alarm() {
    // Both players disconnected and timeout reached, clean up.
    // In a Durable Object, it stays alive as long as there are connections.
    // When connections drop and alarm fires, it will eventually be evicted.
    // We could clear storage here if we were persisting anything.
    this.state.storage.deleteAll();
  }

  sendTo(ws: WebSocket, msg: StrictRelayMessage) {
    try {
      ws.send(JSON.stringify(msg));
    } catch (e) {
      // Handle send error, ws might be closed
    }
  }
}
