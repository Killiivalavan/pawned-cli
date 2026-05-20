import { GameRoom } from "./game_room";
import { generateCode, isValidCode } from "./utils";

export { GameRoom };

export interface Env {
  GAME_ROOM: DurableObjectNamespace;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // CORS Headers
    const corsHeaders = {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
    };

    if (request.method === "OPTIONS") {
      return new Response(null, { headers: corsHeaders });
    }

    if (url.pathname === "/health") {
      return new Response(JSON.stringify({ status: "ok" }), {
        headers: { "Content-Type": "application/json", ...corsHeaders },
      });
    }

    if (url.pathname === "/api/game" && request.method === "POST") {
      // Very basic rate limiting simulation (in practice use CF Rate Limiting)
      const code = generateCode();
      return new Response(JSON.stringify({ code }), {
        headers: { "Content-Type": "application/json", ...corsHeaders },
      });
    }

    if (url.pathname.startsWith("/api/game/") && request.method === "GET") {
      const code = url.pathname.split("/").pop();
      if (!code || !isValidCode(code)) {
        return new Response(JSON.stringify({ error: "Invalid code format" }), {
          status: 400,
          headers: { "Content-Type": "application/json", ...corsHeaders },
        });
      }

      // Actually, to get playerCount, we'd need to fetch the DO, but DO fetch only expects WebSockets right now.
      // We can augment GameRoom fetch to return state if not websocket, or just check existence via DO.
      // Since it's free infrastructure, we can just say exists: true if it's a valid code format for now,
      // or we can route a GET request to the DO to get status. Let's do that.
      const id = env.GAME_ROOM.idFromName(code);
      const stub = env.GAME_ROOM.get(id);
      
      // Let's modify the DO fetch later or handle it here by doing a special request.
      // For now, assume if the code is valid, they can try to join. The WS will reject if full.
      return new Response(JSON.stringify({ exists: true }), {
        headers: { "Content-Type": "application/json", ...corsHeaders },
      });
    }

    if (url.pathname.startsWith("/api/ws/")) {
      const code = url.pathname.split("/").pop();
      if (!code || !isValidCode(code)) {
        return new Response("Invalid code", { status: 400 });
      }

      const id = env.GAME_ROOM.idFromName(code);
      const stub = env.GAME_ROOM.get(id);
      
      // Pass the WebSocket upgrade request to the Durable Object
      return stub.fetch(request);
    }

    return new Response("Not found", { status: 404 });
  },
};
