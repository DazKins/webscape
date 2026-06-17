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
    wsClient.sendMessage(createCommand("join", { id: myPlayerId }));
  },
  onDisconnect: () => {},
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
        game.handleEntityRemove(data.entityId ?? data.id);
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
      case "conversation":
        game.handleConversation(data);
        break;
      case "questCompleted":
        game.handleQuestCompleted(data);
        break;
      default:
        console.warn("Unknown message type:", type);
    }
  },
});

game.registerWsClient(wsClient);

wsClient.connect();

let previousFrameTime: number | null = null;

function animate(frameTime: number) {
  requestAnimationFrame(animate);

  const deltaSeconds =
    previousFrameTime === null ? 0 : Math.min((frameTime - previousFrameTime) / 1000, 0.1);
  previousFrameTime = frameTime;

  game.update(deltaSeconds);
}

requestAnimationFrame(animate);
