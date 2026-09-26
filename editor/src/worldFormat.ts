import { isItemDefinitionId, normalizeItemReference, validateItemReference } from "./itemReference";
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

    validateNamedMetadata(entity.id, entity.components, errors);
    validateWoodcuttable(entity.id, entity.components, errors);
    validateFishable(entity.id, entity.components, errors);
    validateAppearance(entity.id, entity.components, errors);
    validateEquipped(entity.id, entity.components, errors);
    validateShop(entity.id, entity.components, errors);
    validateBanker(entity.id, entity.components, errors);
    validateLootable(entity.id, entity.components, errors);
    const spawn = isObject(entity.components.spawn) ? entity.components.spawn : null;
    const template = spawn && isObject(spawn.entity) ? spawn.entity : null;
    const templateComponents = template && isObject(template.components) ? template.components : null;
    if (templateComponents) {
      validateNamedMetadata(`${entity.id} child template`, templateComponents, errors);
      validateWoodcuttable(`${entity.id} child template`, templateComponents, errors);
      validateFishable(`${entity.id} child template`, templateComponents, errors);
      validateAppearance(`${entity.id} child template`, templateComponents, errors);
      validateEquipped(`${entity.id} child template`, templateComponents, errors);
      validateShop(`${entity.id} child template`, templateComponents, errors);
      validateBanker(`${entity.id} child template`, templateComponents, errors);
      validateLootable(`${entity.id} child template`, templateComponents, errors);
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
  validateItemReference(fishingYield, `entity "${entityId}" yield`, errors);
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
  validateItemReference(materialYield, `entity "${entityId}" yield`, errors);
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
    components: isObject(value.components) ? normalizeItemReferences(value.components) : {},
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

function validateEquipped(entityId: string, components: Record<string, unknown>, errors: string[]) {
  if ("equipped" in components) {
    const equipment = components.equipped;
    if (!isObject(equipment) || Object.keys(equipment).some(key => key !== "slots") ||
        ("slots" in equipment && (!isObject(equipment.slots) ||
          Object.entries(equipment.slots).some(([slot, id]) =>
            !["head", "chest", "legs", "feet", "weapon", "offhand"].includes(slot) || !isItemDefinitionId(id))))) {
      errors.push(`entity "${entityId}" equipped.slots must map equipment slots to non-empty item definition IDs`);
    }
  }
}

function validateBanker(entityId: string, components: Record<string, unknown>, errors: string[]) {
  if (!("banker" in components)) return;
  if (!isObject(components.banker) || Object.keys(components.banker).length !== 0) {
    errors.push(`entity "${entityId}" banker must be an empty object`);
  }
}

function validateShop(entityId: string, components: Record<string, unknown>, errors: string[]) {
  if (!("shop" in components)) return;
  const shop = components.shop;
  if (!isObject(shop) || Object.keys(shop).length !== 1 || !Array.isArray(shop.offers) || shop.offers.length < 1 || shop.offers.length > 100) {
    errors.push(`entity "${entityId}" shop must contain only offers, with 1 to 100 entries`);
    return;
  }
  const seen = new Set<string>();
  for (const rawOffer of shop.offers) {
    const offer = isObject(rawOffer) && !("definitionId" in rawOffer) && "itemId" in rawOffer
      ? { ...rawOffer, definitionId: rawOffer.itemId } : rawOffer;
    if (isObject(offer) && "itemId" in offer && !(isObject(rawOffer) && "definitionId" in rawOffer)) delete offer.itemId;
    if (!isObject(offer) || Object.keys(offer).length !== 3 || !isItemDefinitionId(offer.definitionId) || offer.definitionId === "gold" || seen.has(offer.definitionId)) {
      errors.push(`entity "${entityId}" shop offer must have a non-empty, unique definitionId, buyPrice and sellPrice`);
      continue;
    }
    seen.add(offer.definitionId);
    if (!Number.isInteger(offer.buyPrice) || !Number.isInteger(offer.sellPrice) ||
      Number(offer.sellPrice) < 1 || Number(offer.buyPrice) > 1000000 || Number(offer.sellPrice) >= Number(offer.buyPrice)) {
      errors.push(`entity "${entityId}" shop prices must be integers from 1 to 1000000 with sellPrice below buyPrice`);
    }
  }
}

function validateLootable(id: string, components: Record<string, unknown>, errors: string[]) {
  if (!("lootable" in components)) return;
  const lootable = components.lootable;
  if (!isObject(lootable) || !Array.isArray(lootable.items)) {
    errors.push(`entity "${id}" lootable must contain an items array`);
    return;
  }
  lootable.items.forEach((item, index) => validateItemReference(item, `entity "${id}" loot item ${index + 1}`, errors));
}

function normalizeItemReferences(components: Record<string, unknown>): Record<string, unknown> {
  const result = { ...components };
  for (const key of ["fishable", "woodcuttable"]) {
    const value = result[key];
    if (isObject(value) && "yield" in value) result[key] = { ...value, yield: normalizeItemReference(value.yield) };
  }
  const lootable = result.lootable;
  if (isObject(lootable) && Array.isArray(lootable.items)) result.lootable = { ...lootable, items: lootable.items.map(normalizeItemReference) };
  const shop = result.shop;
  if (isObject(shop) && Array.isArray(shop.offers)) {
    result.shop = { ...shop, offers: shop.offers.map(offer => {
      if (!isObject(offer) || "definitionId" in offer || !("itemId" in offer)) return offer;
      const { itemId, ...rest } = offer;
      return { ...rest, definitionId: itemId };
    }) };
  }
  const spawn = result.spawn;
  if (isObject(spawn) && isObject(spawn.entity) && isObject(spawn.entity.components)) {
    result.spawn = { ...spawn, entity: { ...spawn.entity, components: normalizeItemReferences(spawn.entity.components) } };
  }
  return result;
}

function validateNamedMetadata(id: string, components: Record<string, unknown>, errors: string[]) {
  const metadata = isObject(components.metadata) ? components.metadata : {};
  if (!("named" in metadata)) return;
  if (typeof metadata.named !== "boolean") {
    errors.push(`entity "${id}" metadata.named must be a boolean`);
  } else if (metadata.named && (typeof metadata.name !== "string" || !metadata.name.trim())) {
    errors.push(`entity "${id}" named metadata requires a non-empty name`);
  }
}
