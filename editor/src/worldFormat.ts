import { ID_PATTERN, isObject, serializeJson, type ValidationResult } from "./formatUtils";

export type { ValidationResult } from "./formatUtils";

export type WorldFormat = {
  formatVersion: 2;
  id: string;
  displayName?: string;
  coordinate: { x: number; y: number };
  /** Editor-only dimensions supplied by the uniform project chunk size. */
  size: WorldSize;
  terrain: string[];
  heights: number[];
  blockers: boolean[];
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

const DEFAULT_SIZE: WorldSize = { x: 32, y: 32 };
const MIN_HEIGHT = 0;
const MAX_HEIGHT = 10;
const APPEARANCE_VALUES = {
  skinTone: ["porcelain", "fair", "tan", "brown", "deep"],
  hairStyle: ["cropped", "swept", "bob", "curls"],
  hairColor: ["black", "darkBrown", "chestnut", "auburn", "golden", "gray"],
  tunicColor: ["slateBlue", "forest", "rust", "mustard", "plum", "teal", "burgundy"],
  trousersColor: ["charcoal", "navy", "umber", "olive", "taupe"],
  shoeColor: ["darkBrown", "oxblood", "charcoal", "tan"],
} as const;

export function createBlankWorld(size: WorldSize = DEFAULT_SIZE): WorldFormat {
  return {
    formatVersion: 2,
    id: "chunk_0_0",
    displayName: "Chunk (0, 0)",
    coordinate: { x: 0, y: 0 },
    size,
    terrain: new Array(size.x * size.y).fill("grass"),
    heights: new Array(size.x * size.y).fill(0),
    blockers: new Array(size.x * size.y).fill(false),
    walls: [],
    entities: [
      { id: "player_spawn", components: { position: { x: 0, y: 0 }, playerSpawn: {} } },
    ],
  };
}

export function tileIndex(size: WorldSize, x: number, y: number): number {
  return y * size.x + x;
}

export function normalizeWorld(value: unknown, chunkSize: WorldSize = DEFAULT_SIZE): WorldFormat {
  if (!isObject(value)) {
    throw new Error("chunk data must contain a JSON object");
  }
  if (value.formatVersion !== 2) {
    throw new Error("chunk formatVersion must be 2");
  }

  const size = chunkSize;
  const coordinateValue = isObject(value.coordinate) ? value.coordinate : {};

  const world: WorldFormat = {
    formatVersion: 2,
    id: typeof value.id === "string" ? value.id : "untitled",
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    coordinate: { x: Number(coordinateValue.x), y: Number(coordinateValue.y) },
    size,
    terrain: Array.isArray(value.terrain) ? value.terrain.map(String) : [],
    heights: Array.isArray(value.heights) ? ([...value.heights] as number[]) : [],
    blockers: Array.isArray(value.blockers) ? value.blockers.map(Boolean) : [],
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

  if (world.formatVersion !== 2) {
    errors.push("formatVersion must be 2");
  }

  if (!ID_PATTERN.test(world.id)) {
    errors.push("id must use lowercase letters, numbers, underscores, or dashes");
  }

  if (!Number.isInteger(world.coordinate.x) || !Number.isInteger(world.coordinate.y)) {
    errors.push("coordinate must contain integers");
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

  if (world.blockers.length !== tileCount) {
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
    } else {
      const footprint = entitySize(entity);
      if (position.x < 0 || position.y < 0 || position.x + footprint.width > world.size.x || position.y + footprint.height > world.size.y) {
        errors.push(`entity "${entity.id}" footprint is out of chunk bounds`);
      }
    }

    validateWoodcuttable(entity.id, entity.components, errors);
    validateFishable(entity.id, entity.components, errors);
    validateAppearance(entity.id, entity.components, errors);
    const spawn = isObject(entity.components.spawn) ? entity.components.spawn : null;
    const template = spawn && isObject(spawn.entity) ? spawn.entity : null;
    const templateComponents = template && isObject(template.components) ? template.components : null;
    if (templateComponents) {
      validateWoodcuttable(`${entity.id} child template`, templateComponents, errors);
      validateFishable(`${entity.id} child template`, templateComponents, errors);
      validateAppearance(`${entity.id} child template`, templateComponents, errors);
    }
  }

  return { valid: errors.length === 0, errors };
}

function validateFishable(
  entityId: string,
  components: Record<string, unknown>,
  errors: string[]
): void {
  if (!Object.prototype.hasOwnProperty.call(components, "fishable")) {
    return;
  }
  const fishable = isObject(components.fishable) ? components.fishable : null;
  if (!fishable) {
    errors.push(`entity "${entityId}" fishable must be an object`);
    return;
  }
  if (!Number.isInteger(fishable.catchChancePercent) || Number(fishable.catchChancePercent) < 1 || Number(fishable.catchChancePercent) > 100) {
    errors.push(`entity "${entityId}" fishable.catchChancePercent must be an integer from 1 to 100`);
  }
  const fishingYield = isObject(fishable.yield) ? fishable.yield : null;
  if (!fishingYield) {
    errors.push(`entity "${entityId}" fishable.yield must be an object`);
    return;
  }
  if (typeof fishingYield.name !== "string" || fishingYield.name.trim().length === 0) {
    errors.push(`entity "${entityId}" fishable.yield.name must be a non-empty string`);
  }
  if (typeof fishingYield.type !== "string" || fishingYield.type.trim().length === 0) {
    errors.push(`entity "${entityId}" fishable.yield.type must be a non-empty string`);
  }
  if (!Number.isInteger(fishingYield.count) || Number(fishingYield.count) < 1) {
    errors.push(`entity "${entityId}" fishable.yield.count must be a positive integer`);
  }
}

function validateAppearance(
  entityId: string,
  components: Record<string, unknown>,
  errors: string[]
): void {
  if (!Object.prototype.hasOwnProperty.call(components, "appearance")) {
    return;
  }
  const appearance = isObject(components.appearance) ? components.appearance : null;
  if (!appearance) {
    errors.push(`entity "${entityId}" appearance must be an object`);
    return;
  }

  const fields = Object.keys(APPEARANCE_VALUES) as Array<keyof typeof APPEARANCE_VALUES>;
  for (const key of Object.keys(appearance)) {
    if (!fields.includes(key as keyof typeof APPEARANCE_VALUES)) {
      errors.push(`entity "${entityId}" appearance contains unknown field "${key}"`);
    }
  }
  for (const field of fields) {
    const value = appearance[field];
    const allowed: readonly string[] = APPEARANCE_VALUES[field];
    if (typeof value !== "string" || !allowed.includes(value)) {
      errors.push(`entity "${entityId}" appearance.${field} must be one of ${allowed.join(", ")}`);
    }
  }
}

function validateWoodcuttable(
  entityId: string,
  components: Record<string, unknown>,
  errors: string[]
): void {
  if (!Object.prototype.hasOwnProperty.call(components, "woodcuttable")) {
    return;
  }
  const woodcuttable = isObject(components.woodcuttable) ? components.woodcuttable : null;
  if (!woodcuttable) {
    errors.push(`entity "${entityId}" woodcuttable must be an object`);
    return;
  }
  if (!Number.isInteger(woodcuttable.maxDurability) || Number(woodcuttable.maxDurability) < 1) {
    errors.push(`entity "${entityId}" woodcuttable.maxDurability must be a positive integer`);
  }
  if (!Number.isInteger(woodcuttable.respawnTicks) || Number(woodcuttable.respawnTicks) < 1) {
    errors.push(`entity "${entityId}" woodcuttable.respawnTicks must be a positive integer`);
  }
  const materialYield = isObject(woodcuttable.yield) ? woodcuttable.yield : null;
  if (!materialYield) {
    errors.push(`entity "${entityId}" woodcuttable.yield must be an object`);
    return;
  }
  if (typeof materialYield.name !== "string" || materialYield.name.trim().length === 0) {
    errors.push(`entity "${entityId}" woodcuttable.yield.name must be a non-empty string`);
  }
  if (materialYield.type !== "material") {
    errors.push(`entity "${entityId}" woodcuttable.yield.type must be "material"`);
  }
  if (!Number.isInteger(materialYield.count) || Number(materialYield.count) < 1) {
    errors.push(`entity "${entityId}" woodcuttable.yield.count must be a positive integer`);
  }
}

export function serializeWorld(world: WorldFormat): string {
  const { size: _size, ...authored } = world;
  return serializeJson({
    ...authored,
    heights: world.heights,
    blockers: world.blockers,
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
