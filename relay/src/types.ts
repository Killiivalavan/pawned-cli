export interface GameState {
  code: string;
  fen: string;
  playerCount: number;
  createdAt: number;
}

export type RelayMessageType =
  | "create_game"
  | "game_created"
  | "join_game"
  | "game_start"
  | "move"
  | "opponent_disconnected"
  | "opponent_reconnected"
  | "ping"
  | "pong"
  | "error";

export interface MovePayload {
  fen: string;
  uci: string;
  gameOver: boolean;
  result: string;
}

export interface RelayMessage {
  type: RelayMessageType;
  payload?: any; // We will use specific payload types based on the message type in practice, but keeping it flexible here or use union types.
}

// A more strictly typed RelayMessage
export type StrictRelayMessage =
  | { type: "create_game" }
  | { type: "game_created"; payload: { code: string } }
  | { type: "join_game"; payload: { code: string } }
  | { type: "game_start"; payload: { color: "white" | "black" } }
  | { type: "move"; payload: MovePayload }
  | { type: "opponent_disconnected" }
  | { type: "opponent_reconnected" }
  | { type: "ping" }
  | { type: "pong" }
  | { type: "error"; payload: { message: string } };
