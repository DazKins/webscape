export type WorldFormat = {
  formatVersion: 1;
  id: string;
  displayName?: string;
  size: WorldSize;
  terrain: string[];
  blockers?: boolean[];
  walls: WorldWall[];
  objects: WorldObject[];
  spawns: SpawnPoint[];
};

export type WorldSize = {
  x: number;
  y: number;
};

export type WorldObject = {
  id: string;
  type: string;
  x: number;
  y: number;
  width?: number;
  height?: number;
  blocksMovement?: boolean;
  interactable?: string[];
  state?: Record<string, unknown>;
};

export type WorldWall = {
  id: string;
  type: string;
  x: number;
  y: number;
};

export type SpawnPoint = {
  type: "player" | "npc" | "monster";
  x: number;
  y: number;
  entityType?: string;
  name?: string;
};

export type ValidationResult = {
  valid: boolean;
  errors: string[];
};

const DEFAULT_SIZE: WorldSize = { x: 12, y: 8 };

export function createBlankWorld(): WorldFormat {
  return {
    formatVersion: 1,
    id: "new_world",
    displayName: "New World",
    size: DEFAULT_SIZE,
    terrain: new Array(DEFAULT_SIZE.x * DEFAULT_SIZE.y).fill("grass"),
    blockers: new Array(DEFAULT_SIZE.x * DEFAULT_SIZE.y).fill(false),
    walls: [],
    objects: [],
    spawns: [],
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
    blockers: Array.isArray(value.blockers) ? value.blockers.map(Boolean) : undefined,
    walls: Array.isArray(value.walls) ? value.walls.map(normalizeWall) : [],
    objects: Array.isArray(value.objects) ? value.objects.map(normalizeObject) : [],
    spawns: Array.isArray(value.spawns) ? value.spawns.map(normalizeSpawn) : [],
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

  if (!/^[a-z0-9][a-z0-9_-]*$/.test(world.id)) {
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

  if (world.blockers && world.blockers.length !== tileCount) {
    errors.push(`blockers length must be ${tileCount}`);
  }

  for (const wall of world.walls) {
    if (!wall.id || !/^[a-z0-9][a-z0-9_-]*$/.test(wall.id)) {
      errors.push(`wall id "${wall.id}" is invalid`);
    }
    if (!wall.type) {
      errors.push(`wall "${wall.id}" must have a type`);
    }
    if (!isInBounds(world.size, wall.x, wall.y)) {
      errors.push(`wall "${wall.id}" is out of bounds`);
    }
  }

  for (const object of world.objects) {
    if (!object.id || !/^[a-z0-9][a-z0-9_-]*$/.test(object.id)) {
      errors.push(`object id "${object.id}" is invalid`);
    }
    if (!object.type) {
      errors.push(`object "${object.id}" must have a type`);
    }
    if (!isInBounds(world.size, object.x, object.y)) {
      errors.push(`object "${object.id}" is out of bounds`);
    }
    if (object.width !== undefined && (!Number.isInteger(object.width) || object.width < 1)) {
      errors.push(`object "${object.id}" width must be a positive integer`);
    }
    if (object.height !== undefined && (!Number.isInteger(object.height) || object.height < 1)) {
      errors.push(`object "${object.id}" height must be a positive integer`);
    }
  }

  for (const spawn of world.spawns) {
    if (!spawn.type) {
      errors.push("spawn type is required");
    }
    if (!isInBounds(world.size, spawn.x, spawn.y)) {
      errors.push(`spawn "${spawn.type}" at (${spawn.x}, ${spawn.y}) is out of bounds`);
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
    objects: world.objects.filter((object) => isInBounds(nextSize, object.x, object.y)),
    spawns: world.spawns.filter((spawn) => isInBounds(nextSize, spawn.x, spawn.y)),
  };
}

export function serializeWorld(world: WorldFormat): string {
  return `${JSON.stringify(
    {
      ...world,
      blockers: world.blockers ?? new Array(world.size.x * world.size.y).fill(false),
      walls: world.walls,
      objects: world.objects,
      spawns: world.spawns,
    },
    null,
    2
  )}\n`;
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

function normalizeObject(value: unknown): WorldObject {
  if (!isObject(value)) {
    return { id: "object_invalid", type: "object", x: 0, y: 0 };
  }

  return {
    id: typeof value.id === "string" ? value.id : "object_invalid",
    type: typeof value.type === "string" ? value.type : "object",
    x: Number(value.x),
    y: Number(value.y),
    width: value.width === undefined ? undefined : Number(value.width),
    height: value.height === undefined ? undefined : Number(value.height),
    blocksMovement: value.blocksMovement === undefined ? undefined : Boolean(value.blocksMovement),
    interactable: Array.isArray(value.interactable) ? value.interactable.map(String) : [],
    state: isObject(value.state) ? value.state : undefined,
  };
}

function normalizeSpawn(value: unknown): SpawnPoint {
  if (!isObject(value)) {
    return { type: "player", x: 0, y: 0 };
  }

  const type = value.type === "npc" || value.type === "monster" ? value.type : "player";

  return {
    type,
    x: Number(value.x),
    y: Number(value.y),
    entityType: typeof value.entityType === "string" ? value.entityType : undefined,
    name: typeof value.name === "string" ? value.name : undefined,
  };
}

function isInBounds(size: WorldSize, x: number, y: number): boolean {
  return Number.isInteger(x) && Number.isInteger(y) && x >= 0 && y >= 0 && x < size.x && y < size.y;
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
