import { ID_PATTERN, isObject, serializeJson, type ValidationResult } from "./formatUtils";

export type { ValidationResult } from "./formatUtils";

export type WorldFormat = {
  formatVersion: 1;
  id: string;
  displayName?: string;
  size: WorldSize;
  terrain: string[];
  heights: number[];
  blockers?: boolean[];
  walls: WorldWall[];
  entities: WorldEntity[];
};

export type WorldSize = {
  x: number;
  y: number;
};

export type WorldEntity = {
  id: string;
  components: Record<string, unknown>;
};

export type WorldWall = {
  id: string;
  type: string;
  x: number;
  y: number;
};

const DEFAULT_SIZE: WorldSize = { x: 12, y: 8 };
const MIN_HEIGHT = 0;
const MAX_HEIGHT = 10;

export function createBlankWorld(): WorldFormat {
  return {
    formatVersion: 1,
    id: "new_world",
    displayName: "New World",
    size: DEFAULT_SIZE,
    terrain: new Array(DEFAULT_SIZE.x * DEFAULT_SIZE.y).fill("grass"),
    heights: new Array(DEFAULT_SIZE.x * DEFAULT_SIZE.y).fill(0),
    blockers: new Array(DEFAULT_SIZE.x * DEFAULT_SIZE.y).fill(false),
    walls: [],
    entities: [],
  };
}

export function tileIndex(size: WorldSize, x: number, y: number): number {
  return y * size.x + x;
}

export function normalizeWorld(value: unknown): WorldFormat {
  if (!isObject(value)) {
    throw new Error("map data must contain a JSON object");
  }

  const sizeValue = value.size;
  if (!isObject(sizeValue)) {
    throw new Error("world.size is required");
  }

  const size = {
    x: Number(sizeValue.x),
    y: Number(sizeValue.y),
  };

  const world: WorldFormat = {
    formatVersion: 1,
    id: typeof value.id === "string" ? value.id : "untitled",
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    size,
    terrain: Array.isArray(value.terrain) ? value.terrain.map(String) : [],
    heights: Array.isArray(value.heights) ? ([...value.heights] as number[]) : [],
    blockers: Array.isArray(value.blockers) ? value.blockers.map(Boolean) : undefined,
    walls: Array.isArray(value.walls) ? value.walls.map(normalizeWall) : [],
    entities: Array.isArray(value.entities) ? value.entities.map(normalizeEntity) : [],
  };

  const validation = validateWorld(world);
  if (!validation.valid) {
    throw new Error(validation.errors.join("\n"));
  }

  return world;
}

export function validateWorld(world: WorldFormat): ValidationResult {
  const errors: string[] = [];
  const tileCount = world.size.x * world.size.y;

  if (world.formatVersion !== 1) {
    errors.push("formatVersion must be 1");
  }

  if (!ID_PATTERN.test(world.id)) {
    errors.push("id must use lowercase letters, numbers, underscores, or dashes");
  }

  if (!Number.isInteger(world.size.x) || world.size.x < 1) {
    errors.push("size.x must be a positive integer");
  }

  if (!Number.isInteger(world.size.y) || world.size.y < 1) {
    errors.push("size.y must be a positive integer");
  }

  if (world.terrain.length !== tileCount) {
    errors.push(`terrain length must be ${tileCount}`);
  }

  const heights = Array.isArray(world.heights) ? world.heights : [];
  if (heights.length !== tileCount) {
    errors.push(`heights length must be ${tileCount}`);
  }
  heights.forEach((height, index) => {
    if (!Number.isInteger(height) || height < MIN_HEIGHT || height > MAX_HEIGHT) {
      errors.push(`heights[${index}] must be an integer from ${MIN_HEIGHT} to ${MAX_HEIGHT}`);
    }
  });

  if (world.blockers && world.blockers.length !== tileCount) {
    errors.push(`blockers length must be ${tileCount}`);
  }

  for (const wall of world.walls) {
    if (!wall.id || !ID_PATTERN.test(wall.id)) {
      errors.push(`wall id "${wall.id}" is invalid`);
    }
    if (!wall.type) {
      errors.push(`wall "${wall.id}" must have a type`);
    }
    if (!isInBounds(world.size, wall.x, wall.y)) {
      errors.push(`wall "${wall.id}" is out of bounds`);
    }
  }

  for (const entity of world.entities) {
    const position = entityPosition(entity);
    if (!entity.id || !ID_PATTERN.test(entity.id)) {
      errors.push(`entity id "${entity.id}" is invalid`);
    }
    if (!position) {
      errors.push(`entity "${entity.id}" must include a position component`);
    } else if (!isInBounds(world.size, position.x, position.y)) {
      errors.push(`entity "${entity.id}" is out of bounds`);
    }
  }

  return { valid: errors.length === 0, errors };
}

export function resizeWorld(world: WorldFormat, nextSize: WorldSize, fillTerrain: string): WorldFormat {
  const terrain = new Array(nextSize.x * nextSize.y).fill(fillTerrain || "grass");
  const blockers = new Array(nextSize.x * nextSize.y).fill(false);

  for (let y = 0; y < Math.min(world.size.y, nextSize.y); y += 1) {
    for (let x = 0; x < Math.min(world.size.x, nextSize.x); x += 1) {
      terrain[tileIndex(nextSize, x, y)] = world.terrain[tileIndex(world.size, x, y)];
      blockers[tileIndex(nextSize, x, y)] = Boolean(world.blockers?.[tileIndex(world.size, x, y)]);
    }
  }

  return {
    ...world,
    size: nextSize,
    terrain,
    blockers,
    walls: world.walls.filter((wall) => isInBounds(nextSize, wall.x, wall.y)),
    entities: world.entities.filter((entity) => {
      const position = entityPosition(entity);
      return Boolean(position && isInBounds(nextSize, position.x, position.y));
    }),
  };
}

export function serializeWorld(world: WorldFormat): string {
  return serializeJson({
    ...world,
    heights: world.heights,
    blockers: world.blockers ?? new Array(world.size.x * world.size.y).fill(false),
    walls: world.walls,
    entities: world.entities,
  });
}

function normalizeEntity(value: unknown): WorldEntity {
  if (!isObject(value)) {
    return { id: "entity_invalid", components: {} };
  }

  return {
    id: typeof value.id === "string" ? value.id : "entity_invalid",
    components: isObject(value.components) ? value.components : {},
  };
}

function normalizeWall(value: unknown): WorldWall {
  if (!isObject(value)) {
    return { id: "wall_invalid", type: "stone", x: 0, y: 0 };
  }

  return {
    id: typeof value.id === "string" ? value.id : "wall_invalid",
    type: typeof value.type === "string" ? value.type : "stone",
    x: Number(value.x),
    y: Number(value.y),
  };
}

export function entityPosition(entity: WorldEntity): { x: number; y: number } | null {
  const position = isObject(entity.components.position) ? entity.components.position : null;
  if (!position) {
    return null;
  }
  const x = Number(position.x);
  const y = Number(position.y);
  if (!Number.isInteger(x) || !Number.isInteger(y)) {
    return null;
  }
  return { x, y };
}

export function entitySize(entity: WorldEntity): { width: number; height: number } {
  const metadata = isObject(entity.components.metadata) ? entity.components.metadata : {};
  const width = Number(metadata.width);
  const height = Number(metadata.height);
  return {
    width: Number.isInteger(width) && width > 0 ? width : 1,
    height: Number.isInteger(height) && height > 0 ? height : 1,
  };
}

function isInBounds(size: WorldSize, x: number, y: number): boolean {
  return Number.isInteger(x) && Number.isInteger(y) && x >= 0 && y >= 0 && x < size.x && y < size.y;
}
