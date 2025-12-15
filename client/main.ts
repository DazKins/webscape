import UiRoot from "./ui/uiRoot.tsx";
import { WebSocketClient } from "./ws.js";
import { createCommand } from "./command/command.ts";
import Game from "./game/game.ts";
import { createRoot } from "react-dom/client";
import React from "react";

let myPlayerId = window.localStorage.getItem("myPlayerId");
if (!myPlayerId) {
  myPlayerId = crypto.randomUUID();
  window.localStorage.setItem("myPlayerId", myPlayerId);
}

const sceneLayerRoot = document.getElementById("sceneLayerRoot")!;
const hudLayerRoot = document.getElementById("hudLayerRoot")!;

const game = new Game(sceneLayerRoot, hudLayerRoot);

const uiRoot = document.getElementById("uiLayerRoot")!;
const root = createRoot(uiRoot);
root.render(React.createElement(UiRoot, { game: game }));

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
      case "gameUpdate":
        game.handleGameUpdate(data);
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
        console.log("Unknown message type:", type);
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
