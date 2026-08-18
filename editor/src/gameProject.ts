import { ID_PATTERN, isObject, serializeJson, type ValidationResult } from "./formatUtils";

export type { ValidationResult } from "./formatUtils";

export type GameProject = {
  formatVersion: 2;
  id: string;
  displayName?: string;
  world: { chunkSize: { x: number; y: number } };
  files: GameProjectFiles;
};

export type GameProjectFiles = {
  chunks: string[];
  conversations: string[];
  quests: string[];
};

export const DEFAULT_WORLD_PATH = "chunks/chunk_0_0.json";
export const MAP_FILE_DIRECTORY = "chunks";
export const CONVERSATION_FILE_DIRECTORY = "conversations";
export const QUEST_FILE_DIRECTORY = "quests";

export function createBlankGameProject(): GameProject {
  return {
    formatVersion: 2,
    id: "new_game",
    displayName: "New Game",
    world: { chunkSize: { x: 32, y: 32 } },
    files: {
      chunks: [DEFAULT_WORLD_PATH],
      conversations: [],
      quests: [],
    },
  };
}

export function normalizeGameProject(value: unknown): GameProject {
  if (!isObject(value)) {
    throw new Error("project data must contain a JSON object");
  }
  if (value.formatVersion !== 2) {
    throw new Error("project formatVersion must be 2");
  }

  const files = isObject(value.files) ? value.files : {};
  const worldValue = isObject(value.world) ? value.world : {};
  const chunkSizeValue = isObject(worldValue.chunkSize) ? worldValue.chunkSize : {};
  const project: GameProject = {
    formatVersion: 2,
    id: typeof value.id === "string" ? value.id : "untitled_game",
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    world: { chunkSize: { x: Number(chunkSizeValue.x), y: Number(chunkSizeValue.y) } },
    files: {
      chunks: normalizePathList(files.chunks),
      conversations: normalizePathList(files.conversations),
      quests: normalizePathList(files.quests),
    },
  };

  const validation = validateGameProject(project);
  if (!validation.valid) {
    throw new Error(validation.errors.join("\n"));
  }

  return project;
}

export function validateGameProject(project: GameProject): ValidationResult {
  const errors: string[] = [];

  if (project.formatVersion !== 2) {
    errors.push("project formatVersion must be 2");
  }

  if (!ID_PATTERN.test(project.id)) {
    errors.push("project id must use lowercase letters, numbers, underscores, or dashes");
  }

  if (!Number.isInteger(project.world.chunkSize.x) || project.world.chunkSize.x < 1 || !Number.isInteger(project.world.chunkSize.y) || project.world.chunkSize.y < 1) {
    errors.push("world chunkSize must contain positive integers");
  }
  validatePathList(errors, "chunk", project.files.chunks);
  validatePathList(errors, "conversation", project.files.conversations);
  validatePathList(errors, "quest", project.files.quests);

  if (project.files.chunks.length === 0) {
    errors.push("project must include at least one chunk");
  }

  return { valid: errors.length === 0, errors };
}

export function serializeGameProject(project: GameProject): string {
  return serializeJson(project);
}

export function ensureProjectMapPath(project: GameProject, mapPath: string): GameProject {
  if (!isValidProjectPath(mapPath) || project.files.chunks.includes(mapPath)) {
    return project;
  }

  return {
    ...project,
    files: {
      ...project.files,
      chunks: [...project.files.chunks, mapPath],
    },
  };
}

export function setProjectMapPath(project: GameProject, previousPath: string, nextPath: string): GameProject {
  if (!isValidProjectPath(nextPath)) {
    return project;
  }

  const maps = project.files.chunks.length > 0 ? project.files.chunks : [nextPath];
  const nextMaps = maps.map((path, index) => {
    if (path === previousPath || (index === 0 && !maps.includes(previousPath))) {
      return nextPath;
    }
    return path;
  });
  if (!nextMaps.includes(nextPath)) {
    nextMaps.push(nextPath);
  }

  return {
    ...project,
    files: {
      ...project.files,
      chunks: [...new Set(nextMaps)],
    },
  };
}

export function filenameFromProjectPath(path: string): string {
  const parts = path.split("/");
  return parts[parts.length - 1] ?? path;
}

export function mapPathFromFilename(filename: string): string {
  return projectPathFromFilename(MAP_FILE_DIRECTORY, filename);
}

export function conversationPathFromFilename(filename: string): string {
  return projectPathFromFilename(CONVERSATION_FILE_DIRECTORY, filename);
}

export function questPathFromFilename(filename: string): string {
  return projectPathFromFilename(QUEST_FILE_DIRECTORY, filename);
}

export function isValidProjectFilename(filename: string): boolean {
  if (!filename || filename.includes("/") || filename.includes("\\") || filename === "." || filename === "..") {
    return false;
  }

  return isValidProjectPath(filename);
}

export function isValidProjectPath(path: string): boolean {
  if (!path || path.startsWith("/") || path.endsWith("/") || path.includes("\\")) {
    return false;
  }

  return path.split("/").every((part) => part.length > 0 && part !== "." && part !== "..");
}

function projectPathFromFilename(directory: string, filename: string): string {
  const trimmed = filename.trim();
  if (!isValidProjectFilename(trimmed)) {
    return trimmed;
  }

  return `${directory}/${trimmed}`;
}

function normalizePathList(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.map(String).filter((path) => path.length > 0);
}

function validatePathList(errors: string[], label: string, paths: string[]) {
  const seen = new Set<string>();
  for (const path of paths) {
    if (!isValidProjectPath(path)) {
      errors.push(`${label} storage is invalid`);
    }
    if (seen.has(path)) {
      errors.push(`${label} storage is duplicated`);
    }
    seen.add(path);
  }
}
