export const TERRAIN_SWATCHES = ["grass", "dirt", "road", "water", "stone"];

export function terrainColor(terrain: string): string {
  const builtIn: Record<string, string> = {
    grass: "#73964f",
    dirt: "#9a6b42",
    road: "#b8ab88",
    water: "#4f8fb8",
    stone: "#8b9296",
  };

  if (builtIn[terrain]) {
    return builtIn[terrain];
  }

  let hash = 0;
  for (let i = 0; i < terrain.length; i += 1) {
    hash = terrain.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue} 36% 54%)`;
}
