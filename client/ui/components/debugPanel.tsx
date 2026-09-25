import { useEffect, useState } from "react";
import type Game from "../../game/game";
import styles from "./debugPanel.module.css";

function readStats(game: Game, fps: number | null) {
  const { renderer } = game;
  const position = game.getMyEntity()?.getComponent("position");
  return {
    fps,
    playerX: position?.x ?? null,
    playerY: position?.y ?? null,
    timing: game.renderTiming?.takeSample(),
    calls: renderer.info.render.calls,
    triangles: renderer.info.render.triangles,
    entities: game.entities.length,
    chunks: game.world?.chunks.size ?? 0,
    width: renderer.domElement.width,
    height: renderer.domElement.height,
    pixelRatio: renderer.getPixelRatio(),
  };
}

export default function DebugPanel({ game }: { game: Game }) {
  const [stats, setStats] = useState(() => readStats(game, null));

  useEffect(() => {
    game.setRenderTimingEnabled(true);
    let previousFrame = game.renderer.info.render.frame;
    let previousTime = performance.now();
    const resetSample = () => {
      game.renderTiming?.reset();
      previousFrame = game.renderer.info.render.frame;
      previousTime = performance.now();
      setStats(readStats(game, null));
    };
    const interval = window.setInterval(() => {
      if (document.hidden) return;
      const now = performance.now();
      const frame = game.renderer.info.render.frame;
      // Count actual WebGL renders, including the client's frame-rate cap.
      const fps = Math.max(0, frame - previousFrame) * 1000 / (now - previousTime);
      setStats(readStats(game, fps));
      previousFrame = frame;
      previousTime = now;
    }, 1000);
    document.addEventListener("visibilitychange", resetSample);
    return () => {
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", resetSample);
      game.setRenderTimingEnabled(false);
    };
  }, [game]);

  return (
    <aside className={styles.panel} aria-label="Debug information">
      <h2>Debug mode</h2>
      <div className={styles.fps}>
        <strong>{stats.fps === null ? "—" : stats.fps.toFixed(1)}</strong> FPS
      </div>
      <dl>
        <dt>Player X</dt><dd>{stats.playerX ?? "—"}</dd>
        <dt>Player Y</dt><dd>{stats.playerY ?? "—"}</dd>
        <dt>Frame interval</dt><dd>{stats.fps ? `${(1000 / stats.fps).toFixed(1)} ms` : "—"}</dd>
        <dt>Draw CPU</dt><dd>{stats.timing?.cpuMs != null ? `${stats.timing.cpuMs.toFixed(2)} ms` : "—"}</dd>
        <dt>Draw GPU</dt><dd>{stats.timing?.gpuMs != null ? `${stats.timing.gpuMs.toFixed(2)} ms` : stats.timing?.gpuSupported === false ? "N/A" : "—"}</dd>
        <dt>Draw calls</dt><dd>{stats.calls.toLocaleString()}</dd>
        <dt>Triangles</dt><dd>{stats.triangles.toLocaleString()}</dd>
        <dt>Entities</dt><dd>{stats.entities.toLocaleString()}</dd>
        <dt>Chunks</dt><dd>{stats.chunks}</dd>
        <dt>Resolution</dt><dd>{stats.width}×{stats.height}</dd>
        <dt>Pixel ratio</dt><dd>{stats.pixelRatio.toFixed(2)}</dd>
      </dl>
    </aside>
  );
}
