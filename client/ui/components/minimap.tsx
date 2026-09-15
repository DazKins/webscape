import { useEffect, useRef } from "react";
import type Game from "../../game/game";
import { terrainColor, type ChunkLoad } from "../../game/world/world";
import styles from "./minimap.module.css";
import DayCycleIndicator from "./dayCycleIndicator";

const SIZE = 192;
const CENTER = SIZE / 2;
const MAP_RADIUS = 74;
const TILE_SCALE = 4;

export default function Minimap({ game }: { game: Game }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const context = canvas?.getContext("2d");
    if (!canvas || !context) return;

    // Chunk payloads are replaced on updates. Unloaded chunks can be collected.
    const terrainImages = new WeakMap<ChunkLoad, HTMLCanvasElement>();
    const draw = () => {
      const pixelRatio = Math.min(window.devicePixelRatio || 1, 2);
      const resolution = Math.round(SIZE * pixelRatio);
      if (canvas.width !== resolution || canvas.height !== resolution) {
        canvas.width = resolution;
        canvas.height = resolution;
      }
      context.setTransform(resolution / SIZE, 0, 0, resolution / SIZE, 0, 0);
      context.clearRect(0, 0, SIZE, SIZE);
      const rotation = game.camera.getMapRotation();
      const focus = game.getEntityFocusPoint(game.myPlayerId);
      const world = game.world;

      context.save();
      context.translate(CENTER, CENTER);
      context.beginPath();
      context.arc(0, 0, MAP_RADIUS, 0, Math.PI * 2);
      context.fillStyle = "#18232b";
      context.fill();
      context.clip();
      if (world && focus) {
        context.rotate(rotation);
        context.scale(TILE_SCALE, TILE_SCALE);
        context.translate(-focus.x, -focus.z);
        context.imageSmoothingEnabled = false;
        const radius = MAP_RADIUS / TILE_SCALE;
        for (const { data } of world.chunks.values()) {
          const x = data.coordinate.x * world.chunkSize.x;
          const y = data.coordinate.y * world.chunkSize.y;
          if (x > focus.x + radius || y > focus.z + radius ||
              x + world.chunkSize.x < focus.x - radius ||
              y + world.chunkSize.y < focus.z - radius) continue;
          let terrainImage = terrainImages.get(data);
          if (!terrainImage) {
            terrainImage = document.createElement("canvas");
            terrainImage.width = world.chunkSize.x;
            terrainImage.height = world.chunkSize.y;
            const terrainContext = terrainImage.getContext("2d");
            if (!terrainContext) continue;
            data.terrain.forEach((type, index) => {
              terrainContext.fillStyle = `#${terrainColor(type).toString(16).padStart(6, "0")}`;
              terrainContext.fillRect(index % world.chunkSize.x, Math.floor(index / world.chunkSize.x), 1, 1);
            });
            terrainImages.set(data, terrainImage);
          }
          context.drawImage(terrainImage, x, y);
        }
      }
      context.restore();

      context.strokeStyle = "#a0b3bc";
      context.lineWidth = 1;
      context.beginPath();
      context.arc(CENTER, CENTER, MAP_RADIUS, 0, Math.PI * 2);
      context.stroke();

      // Labels orbit with the terrain but remain upright for readability.
      context.font = "bold 13px sans-serif";
      context.textAlign = "center";
      context.textBaseline = "middle";
      ["N", "E", "S", "W"].forEach((label, index) => {
        const angle = rotation - Math.PI / 2 + index * Math.PI / 2;
        context.fillStyle = label === "N" ? "#ff978a" : "#e5edf1";
        context.fillText(label, CENTER + Math.cos(angle) * 85, CENTER + Math.sin(angle) * 85);
      });

      // Match the rendered player's smooth facing, then rotate into map space.
      const facing = game.entityRenderSystem.getRenderers()[game.myPlayerId]?.getFacingRotationY();
      context.fillStyle = "#ffffff";
      if (focus) {
        if (facing != null) {
          context.save();
          context.translate(CENTER, CENTER);
          // Three.js yaw zero faces +Z (south); the chevron starts north-up.
          context.rotate(rotation + Math.PI - facing);
          context.beginPath();
          context.moveTo(0, -11);
          context.lineTo(-4, -5);
          context.lineTo(4, -5);
          context.closePath();
          context.fill();
          context.restore();
        }
        context.beginPath();
        context.arc(CENTER, CENTER, 3.5, 0, Math.PI * 2);
        context.fill();
        context.strokeStyle = "#17232c";
        context.lineWidth = 1.5;
        context.stroke();
      }
    };

    draw();
    game.addEventListener("frameRendered", draw);
    return () => game.removeEventListener("frameRendered", draw);
  }, [game]);

  return (
    <div className={styles.minimap}>
      <canvas
        ref={canvasRef}
        className={styles.map}
        role="img"
        aria-label="Local terrain minimap and compass, rotating with the camera. The white dot is you; the chevron points in the direction your character is facing."
      />
      <DayCycleIndicator game={game} />
    </div>
  );
}
