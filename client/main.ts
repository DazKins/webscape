import UiRoot from "./ui/uiRoot.tsx";
import { WebSocketClient } from "./ws.js";
import { createCommand } from "./command/command.ts";
import Game from "./game/game.ts";
import { createRoot } from "react-dom/client";
import React from "react";
import type { RegistrationViewState } from "./ui/components/onboardingOverlay.tsx";

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

const rememberedName = window.localStorage.getItem("myPlayerName") ?? "";
let registration: RegistrationViewState = {
  phase: "connecting",
  name: rememberedName,
  error: "",
};

function renderUi() {
  game.setRegistrationBlocked(registration.phase !== "registered");
  root.render(
    React.createElement(UiRoot, {
      game,
      registration,
      onRegister: register,
    })
  );
}

function setRegistration(update: Partial<RegistrationViewState>) {
  registration = { ...registration, ...update };
  renderUi();
}

function register(name: string) {
  const normalizedName = name.trim();
  setRegistration({ phase: "registering", name: normalizedName, error: "" });
  wsClient.sendMessage(
    createCommand("register", { id: myPlayerId, name: normalizedName })
  );
}

renderUi();

const wsClient = new WebSocketClient({
  onConnect: () => {
    if (registration.name) {
      register(registration.name);
    } else {
      setRegistration({ phase: "nameEntry", error: "" });
    }
  },
  onDisconnect: () => {
    game.prepareForReconnect();
    setRegistration({
      phase: registration.name ? "reconnecting" : "connecting",
      error: "",
    });
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
      case "world":
        game.registerWorld(data);
        break;
      case "chunkUpdate":
        game.handleChunkUpdate(data);
        break;
      case "registered":
        game.registerMyPlayerId(data.entityId);
        window.localStorage.setItem("myPlayerName", data.name);
        setRegistration({ phase: "registered", name: data.name, error: "" });
        break;
      case "registrationFailed":
        setRegistration({
          phase: "nameEntry",
          error: data.reason || "Registration failed. Please try again.",
        });
        break;
      case "conversation":
        game.handleConversation(data);
        break;
      case "questCompleted":
        game.handleQuestCompleted(data);
        break;
      case "questStarted":
        game.handleQuestStarted(data);
        break;
      case "chatMessage":
        game.handleChatMessage(data);
        break;
      case "combatResolved":
        game.handleCombatResolved(data);
        break;
      case "woodcuttingSwing":
        game.handleWoodcuttingSwing(data);
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
