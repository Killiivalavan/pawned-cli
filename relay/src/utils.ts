import { StrictRelayMessage } from "./types";

export const GAME_TIMEOUT_MS = 30 * 60 * 1000;

export function generateCode(): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"; // No O, 0, I, 1
  let code = "CH-";
  for (let i = 0; i < 4; i++) {
    code += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return code;
}

export function isValidCode(code: string): boolean {
  return /^CH-[A-Z2-9]{4}$/.test(code);
}

export function parseMessage(data: string): StrictRelayMessage | null {
  try {
    const msg = JSON.parse(data);
    if (typeof msg === "object" && msg !== null && typeof msg.type === "string") {
      return msg as StrictRelayMessage;
    }
  } catch (e) {
    // Ignore JSON parse errors
  }
  return null;
}
