import { WebSocketClient } from "./ws.js";
import {
  createMessage,
  getMessageType,
  getMessageData,
} from "./message/message.js";
import Game from "./game/game.js";

const game = new Game();

const wsClient = new WebSocketClient({
  onConnect: () => {
    console.log("Connected to WebSocket server");
    wsClient.sendMessage(createMessage("join", {}));
  },
  onDisconnect: () => {
    console.log("Disconnected from WebSocket server");
  },
  onError: (error) => {
    console.error("WebSocket error:", error);
  },
  onMessage: (msg) => {
    const type = getMessageType(msg);
    const data = getMessageData(msg);

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
        const id = data.entityId;
        game.registerMyId(id);
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
