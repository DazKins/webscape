import { WebSocketClient } from "./ws.js";
import {
  createCommand,
  getCommandType,
  getCommandData,
} from "./command/command.ts";
import Game from "./game/game.ts";

let myPlayerId = window.localStorage.getItem("myPlayerId");
if (!myPlayerId) {
  myPlayerId = crypto.randomUUID();
  window.localStorage.setItem("myPlayerId", myPlayerId);
}

const game = new Game();

const wsClient = new WebSocketClient({
  onConnect: () => {
    console.log("Connected to WebSocket server");
    wsClient.sendMessage(createCommand("join", { id: myPlayerId }));
  },
  onDisconnect: () => {
    console.log("Disconnected from WebSocket server");
  },
  onError: (error) => {
    console.error("WebSocket error:", error);
  },
  onMessage: (msg: any) => {
    const type = msg.metadata.type;
    const data = msg.data;

    switch (type) {
      case "entityUpdate":
        game.handleEntityUpdate(data);
        break;
      case "entityRemove":
        game.handleEntityRemove(data.entityId);
        break;
      case "world":
        game.registerWorld(data);
        break;
      case "joined":
        game.registerMyPlayerId(data.entityId);
        break;
      case "joinFailed":
        window.alert(`JOIN FAILED: ${data.reason}`);
        break;
      default:
        console.log("Unknown message type:", data.type);
    }
  },
});

game.registerWsClient(wsClient);

wsClient.connect();

function animate() {
  requestAnimationFrame(animate);

  game.update();
}

animate();
