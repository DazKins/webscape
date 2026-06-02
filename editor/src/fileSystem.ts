import {
  ensureProjectMapPath,
  normalizeGameProject,
  serializeGameProject,
  type GameProject,
} from "./gameProject.ts";
import {
  normalizeConversationDocument,
  serializeConversationDocument,
  type ConversationDocument,
} from "./conversationFormat.ts";
import {
  normalizeQuestDocument,
  serializeQuestDocument,
  type QuestDocument,
} from "./questFormat.ts";
import { normalizeWorld, serializeWorld, type WorldFormat } from "./worldFormat.ts";

type WritableFile = {
  write(data: string): Promise<void>;
  close(): Promise<void>;
};

type FileHandle = {
  name: string;
  getFile(): Promise<File>;
  createWritable(): Promise<WritableFile>;
};

export type ProjectDirectoryHandle = {
  name: string;
  getFileHandle(name: string, options?: { create?: boolean }): Promise<FileHandle>;
  getDirectoryHandle(name: string, options?: { create?: boolean }): Promise<ProjectDirectoryHandle>;
  removeEntry?(name: string, options?: { recursive?: boolean }): Promise<void>;
};

export type ConversationDocuments = Record<string, ConversationDocument>;
export type QuestDocuments = Record<string, QuestDocument>;
export type WorldDocuments = Record<string, WorldFormat>;

export type OpenedGameProject = {
  handle: ProjectDirectoryHandle;
  project: GameProject;
  world: WorldFormat;
  worldPath: string;
  worlds: WorldDocuments;
  conversations: ConversationDocuments;
  quests: QuestDocuments;
};

declare global {
  interface Window {
    showDirectoryPicker?: (options?: { mode?: "read" | "readwrite" }) => Promise<ProjectDirectoryHandle>;
  }
}

const MANIFEST_PATH = "game.json";

export function supportsProjectFolders(): boolean {
  return Boolean(window.showDirectoryPicker);
}

export async function openGameProjectFolder(): Promise<OpenedGameProject> {
  if (!window.showDirectoryPicker) {
    throw new Error("File System Access API is not available in this browser");
  }

  const handle = await window.showDirectoryPicker({ mode: "readwrite" });
  const project = normalizeGameProject(JSON.parse(await readTextFile(handle, MANIFEST_PATH)));
  const worldPath = project.files.maps[0];
  if (!worldPath) {
    throw new Error("project must include at least one map");
  }

  const worlds = await readWorldDocuments(handle, project.files.maps);
  const world = worlds[worldPath];
  const conversations = await readConversationDocuments(handle, project.files.conversations);
  const quests = await readQuestDocuments(handle, project.files.quests);
  return { handle, project, world, worldPath, worlds, conversations, quests };
}

export async function saveGameProjectFolder(
  handle: ProjectDirectoryHandle,
  project: GameProject,
  worlds: WorldDocuments,
  worldPath: string,
  conversations: ConversationDocuments,
  quests: QuestDocuments,
  deletedWorldPaths: string[],
  deletedConversationPaths: string[],
  deletedQuestPaths: string[]
): Promise<GameProject> {
  const projectToSave = ensureProjectMapPath(project, worldPath);
  await writeTextFile(handle, MANIFEST_PATH, serializeGameProject(projectToSave));
  await Promise.all(
    projectToSave.files.maps.map(async (path) => {
      const world = worlds[path];
      if (world) {
        await writeTextFile(handle, path, serializeWorld(world));
      }
    })
  );
  await Promise.all(
    projectToSave.files.conversations.map(async (path) => {
      const conversation = conversations[path];
      if (conversation) {
        await writeTextFile(handle, path, serializeConversationDocument(conversation));
      }
    })
  );
  await Promise.all(
    projectToSave.files.quests.map(async (path) => {
      const quest = quests[path];
      if (quest) {
        await writeTextFile(handle, path, serializeQuestDocument(quest));
      }
    })
  );
  await Promise.all(deletedWorldPaths.map((path) => removeTextFile(handle, path)));
  await Promise.all(deletedConversationPaths.map((path) => removeTextFile(handle, path)));
  await Promise.all(deletedQuestPaths.map((path) => removeTextFile(handle, path)));
  return projectToSave;
}

export async function saveGameProjectFolderAs(
  project: GameProject,
  worlds: WorldDocuments,
  worldPath: string,
  conversations: ConversationDocuments,
  quests: QuestDocuments
): Promise<{ handle: ProjectDirectoryHandle; project: GameProject }> {
  if (!window.showDirectoryPicker) {
    throw new Error("File System Access API is not available in this browser");
  }

  const handle = await window.showDirectoryPicker({ mode: "readwrite" });
  const savedProject = await saveGameProjectFolder(handle, project, worlds, worldPath, conversations, quests, [], [], []);
  return { handle, project: savedProject };
}

async function readWorldDocuments(root: ProjectDirectoryHandle, paths: string[]): Promise<WorldDocuments> {
  const entries = await Promise.all(
    paths.map(async (path) => {
      const document = normalizeWorld(JSON.parse(await readTextFile(root, path)));
      return [path, document] as const;
    })
  );

  return Object.fromEntries(entries);
}

async function readConversationDocuments(
  root: ProjectDirectoryHandle,
  paths: string[]
): Promise<ConversationDocuments> {
  const entries = await Promise.all(
    paths.map(async (path) => {
      const document = normalizeConversationDocument(JSON.parse(await readTextFile(root, path)));
      return [path, document] as const;
    })
  );

  return Object.fromEntries(entries);
}

async function readQuestDocuments(root: ProjectDirectoryHandle, paths: string[]): Promise<QuestDocuments> {
  const entries = await Promise.all(
    paths.map(async (path) => {
      const document = normalizeQuestDocument(JSON.parse(await readTextFile(root, path)));
      return [path, document] as const;
    })
  );

  return Object.fromEntries(entries);
}

async function readTextFile(root: ProjectDirectoryHandle, path: string): Promise<string> {
  const { directory, fileName } = await resolveParentDirectory(root, path, false);
  const fileHandle = await directory.getFileHandle(fileName);
  const file = await fileHandle.getFile();
  return file.text();
}

async function writeTextFile(root: ProjectDirectoryHandle, path: string, text: string): Promise<void> {
  const { directory, fileName } = await resolveParentDirectory(root, path, true);
  const fileHandle = await directory.getFileHandle(fileName, { create: true });
  const writable = await fileHandle.createWritable();
  await writable.write(text);
  await writable.close();
}

async function removeTextFile(root: ProjectDirectoryHandle, path: string): Promise<void> {
  const { directory, fileName } = await resolveParentDirectory(root, path, false);
  if (!directory.removeEntry) {
    return;
  }

  try {
    await directory.removeEntry(fileName);
  } catch (error) {
    if (error instanceof DOMException && error.name === "NotFoundError") {
      return;
    }
    throw error;
  }
}

async function resolveParentDirectory(
  root: ProjectDirectoryHandle,
  path: string,
  create: boolean
): Promise<{ directory: ProjectDirectoryHandle; fileName: string }> {
  const parts = path.split("/").filter(Boolean);
  const fileName = parts.pop();
  if (!fileName) {
    throw new Error("invalid project storage");
  }

  let directory = root;
  for (const part of parts) {
    directory = await directory.getDirectoryHandle(part, { create });
  }

  return { directory, fileName };
}
