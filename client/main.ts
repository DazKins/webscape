import { prepareModelAssets } from "./game/assets/constructionClient";
import UiRoot from "./ui/uiRoot.tsx";
import { WebSocketClient } from "./ws.js";
import { createCommand } from "./command/command.ts";
import Game from "./game/game.ts";
import { createRoot } from "react-dom/client";
import React from "react";
import type { RegistrationViewState } from "./ui/components/onboardingOverlay.tsx";

let accountId = "";
let csrfToken = "";
let authenticated = false;
let guest = false;
let username = "";
let authGeneration = 0;
const loginFailed = new URLSearchParams(location.search).get("login") === "failed";
if (loginFailed) history.replaceState(null, "", location.pathname);

const sceneLayerRoot = document.getElementById("sceneLayerRoot")!;
const hudLayerRoot = document.getElementById("hudLayerRoot")!;

const game = new Game(sceneLayerRoot, hudLayerRoot);

const uiRoot = document.getElementById("uiLayerRoot")!;
const root = createRoot(uiRoot);

let registration: RegistrationViewState = {
  phase: "connecting",
  name: "",
  error: "",
};

function renderUi() {
  const inGame = registration.phase === "registered";
  sceneLayerRoot.style.visibility = inGame ? "visible" : "hidden";
  hudLayerRoot.style.visibility = inGame ? "visible" : "hidden";
  game.setRegistrationBlocked(!inGame);
  root.render(
    React.createElement(UiRoot, {
      game,
      registration,
      onRegister: register,
      onLogin: () => { location.assign("/auth/login"); },
      onRetry: () => { wsClient.disconnect(); setRegistration({ phase: "connecting", error: "" }); void wsClient.connect(); },
      onLogout: () => { void logout(); },
      authenticated,
      guest,
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
    createCommand("register", username ? {} : { name: normalizedName })
  );
}

renderUi();

async function checkSession(): Promise<boolean> {
  const generation = authGeneration;
  const response = await fetch("/auth/session", { cache: "no-store", signal: AbortSignal.timeout(10000) });
  if (!response.ok) throw new Error("Session check failed");
  const session = await response.json();
  if (generation !== authGeneration) return false;
  guest = session.guest === true;
  username = typeof session.username === "string" ? session.username : "";
  if (!session.authenticated) {
    authenticated = false;
    csrfToken = "";
    game.clearSession();
    setRegistration({ phase: "signedOut", name: "", error: loginFailed ? "Sign-in did not complete. Please try again." : "" });
    return false;
  }
  if (accountId && accountId !== session.accountId) {
    // A full reload also clears React panel history and pending presentation events.
    wsClient.disconnect();
    game.clearSession();
    location.reload();
    return false;
  }
  authenticated = true;
  csrfToken = session.csrfToken;
  accountId = session.accountId;
  if (username) registration.name = username;
  else if (!registration.name) registration.name = localStorage.getItem(`playerName:${accountId}`) ?? "";
  return true;
}

async function logout() {
  authGeneration++;
  wsClient.disconnect();
  game.clearSession();
  setRegistration({ phase: "signingOut", error: "" });
  try {
    const response = await fetch("/auth/logout", {
      method: "POST", headers: { "X-CSRF-Token": csrfToken }, signal: AbortSignal.timeout(10000),
    });
    if (!response.ok) throw new Error("Logout failed");
    location.replace("/");
  } catch {
    setRegistration({ phase: "connectionError", error: "Could not sign out. Retry sign-out or reconnect." });
  }
}

const wsClient = new WebSocketClient({
  beforeConnect: checkSession,
  onUnavailable: () => setRegistration({ phase: "connectionError", error: "Could not connect. Please try again." }),
  onConnect: () => {
    if (registration.name) {
      register(registration.name);
    } else {
      setRegistration({ phase: "nameEntry", error: "" });
    }
  },
  onDisconnect: () => {
    game.clearSession();
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
      case "tradeResult":
        game.handleTradeResult(data);
        break;
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
        if (!username) window.localStorage.setItem(`playerName:${accountId}`, data.name);
        setRegistration({ phase: "registered", name: data.name, error: "" });
        break;
      case "registrationFailed":
        setRegistration({
          phase: username ? "connectionError" : "nameEntry",
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
      case "itemPickedUp":
        game.handleItemPickedUp(data);
        break;
      case "combatProjectileLaunched":
        game.handleCombatProjectileLaunched(data);
        break;
      default:
        console.warn("Unknown message type:", type);
    }
  },
});

game.registerWsClient(wsClient);

void prepareModelAssets().catch(error => console.error("Model preparation failed", error)).finally(() => wsClient.connect());

// Avoid rendering at 120/144+ FPS on high-refresh displays.
const FRAME_INTERVAL_MS = 1000 / 60;
let previousFrameTime: number | null = null;
let nextFrameTime = 0;
let animationFrameId = 0;

function animate(frameTime: number) {
  if (document.hidden) return;
  animationFrameId = requestAnimationFrame(animate);
  if (frameTime < nextFrameTime - 0.5) return;

  const deltaSeconds =
    previousFrameTime === null ? 0 : Math.min((frameTime - previousFrameTime) / 1000, 0.1);
  previousFrameTime = frameTime;
  // Keep the cadence across skipped refreshes, without catching up after a stall.
  nextFrameTime = Math.max(nextFrameTime + FRAME_INTERVAL_MS, frameTime);

  if (registration.phase === "registered") game.update(deltaSeconds);
}

function resetAnimationLoop() {
  cancelAnimationFrame(animationFrameId);
  previousFrameTime = null;
  nextFrameTime = 0;
  if (!document.hidden) animationFrameId = requestAnimationFrame(animate);
}

document.addEventListener("visibilitychange", resetAnimationLoop);
resetAnimationLoop();
