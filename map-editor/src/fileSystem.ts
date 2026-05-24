import { serializeWorld, type WorldFormat } from "./worldFormat.ts";

type WritableFile = {
  write(data: string): Promise<void>;
  close(): Promise<void>;
};

export type WorldFileHandle = {
  getFile(): Promise<File>;
  createWritable(): Promise<WritableFile>;
};

type FilePickerAcceptType = {
  description: string;
  accept: Record<string, string[]>;
};

declare global {
  interface Window {
    showOpenFilePicker?: (options?: {
      multiple?: false;
      types?: FilePickerAcceptType[];
    }) => Promise<WorldFileHandle[]>;
    showSaveFilePicker?: (options?: {
      suggestedName?: string;
      types?: FilePickerAcceptType[];
    }) => Promise<WorldFileHandle>;
  }
}

const JSON_FILE_TYPES = [
  {
    description: "World JSON",
    accept: {
      "application/json": [".json"],
    },
  },
];

export function supportsFileSystemAccess(): boolean {
  return Boolean(window.showOpenFilePicker && window.showSaveFilePicker);
}

export async function openWorldFile(): Promise<{ handle: WorldFileHandle; text: string }> {
  if (!window.showOpenFilePicker) {
    throw new Error("File System Access API is not available in this browser");
  }

  const [handle] = await window.showOpenFilePicker({
    multiple: false,
    types: JSON_FILE_TYPES,
  });
  const file = await handle.getFile();
  const text = await file.text();
  return { handle, text };
}

export async function saveWorldFile(handle: WorldFileHandle, world: WorldFormat): Promise<void> {
  const writable = await handle.createWritable();
  await writable.write(serializeWorld(world));
  await writable.close();
}

export async function saveWorldFileAs(world: WorldFormat): Promise<WorldFileHandle> {
  if (!window.showSaveFilePicker) {
    throw new Error("File System Access API is not available in this browser");
  }

  const handle = await window.showSaveFilePicker({
    suggestedName: `${world.id || "world"}.json`,
    types: JSON_FILE_TYPES,
  });
  await saveWorldFile(handle, world);
  return handle;
}
