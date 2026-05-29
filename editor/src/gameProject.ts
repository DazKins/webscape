export type GameProject = {
  formatVersion: 1;
  id: string;
  displayName?: string;
  files: GameProjectFiles;
};

export type GameProjectFiles = {
  maps: string[];
  conversations: string[];
  quests: string[];
};

export type ValidationResult = {
  valid: boolean;
  errors: string[];
};

export const DEFAULT_WORLD_PATH = "maps/new_world.json";
export const MAP_FILE_DIRECTORY = "maps";
export const CONVERSATION_FILE_DIRECTORY = "conversations";

export function createBlankGameProject(): GameProject {
  return {
    formatVersion: 1,
    id: "new_game",
    displayName: "New Game",
    files: {
      maps: [DEFAULT_WORLD_PATH],
      conversations: [],
      quests: [],
    },
  };
}

export function normalizeGameProject(value: unknown): GameProject {
  if (!isObject(value)) {
    throw new Error("project data must contain a JSON object");
  }

  const files = isObject(value.files) ? value.files : {};
  const project: GameProject = {
    formatVersion: 1,
    id: typeof value.id === "string" ? value.id : "untitled_game",
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    files: {
      maps: normalizePathList(files.maps),
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

  if (project.formatVersion !== 1) {
    errors.push("project formatVersion must be 1");
  }

  if (!/^[a-z0-9][a-z0-9_-]*$/.test(project.id)) {
    errors.push("project id must use lowercase letters, numbers, underscores, or dashes");
  }

  validatePathList(errors, "map", project.files.maps);
  validatePathList(errors, "conversation", project.files.conversations);
  validatePathList(errors, "quest", project.files.quests);

  if (project.files.maps.length === 0) {
    errors.push("project must include at least one map");
  }

  return { valid: errors.length === 0, errors };
}

export function serializeGameProject(project: GameProject): string {
  return `${JSON.stringify(project, null, 2)}\n`;
}

export function ensureProjectMapPath(project: GameProject, mapPath: string): GameProject {
  if (!isValidProjectPath(mapPath) || project.files.maps.includes(mapPath)) {
    return project;
  }

  return {
    ...project,
    files: {
      ...project.files,
      maps: [...project.files.maps, mapPath],
    },
  };
}

export function setProjectMapPath(project: GameProject, previousPath: string, nextPath: string): GameProject {
  if (!isValidProjectPath(nextPath)) {
    return project;
  }

  const maps = project.files.maps.length > 0 ? project.files.maps : [nextPath];
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
      maps: [...new Set(nextMaps)],
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

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
