import { useMemo, useState } from "react";
import {
  createBlankWorld,
  normalizeWorld,
  resizeWorld,
  tileIndex,
  validateWorld,
  type SpawnPoint,
  type WorldFormat,
  type WorldObject,
} from "./worldFormat.ts";
import {
  openWorldFile,
  saveWorldFile,
  saveWorldFileAs,
  supportsFileSystemAccess,
  type WorldFileHandle,
} from "./fileSystem.ts";

type Tool = "terrain" | "collision" | "object" | "spawn" | "select";
type Selection =
  | { type: "object"; id: string }
  | { type: "spawn"; index: number }
  | { type: "tile"; x: number; y: number }
  | null;

const TERRAIN_SWATCHES = ["grass", "dirt", "road", "water", "stone"];
const OBJECT_TYPES = ["tree", "door", "building", "chest", "rock"];
const SPAWN_TYPES: SpawnPoint["type"][] = ["player", "npc", "monster"];

function App() {
  const [world, setWorld] = useState<WorldFormat>(() => createBlankWorld());
  const [fileHandle, setFileHandle] = useState<WorldFileHandle | null>(null);
  const [fileName, setFileName] = useState<string>("unsaved world");
  const [status, setStatus] = useState<string>("ready");
  const [dirty, setDirty] = useState(false);
  const [tool, setTool] = useState<Tool>("terrain");
  const [selection, setSelection] = useState<Selection>(null);
  const [terrainType, setTerrainType] = useState("grass");
  const [collisionValue, setCollisionValue] = useState(true);
  const [objectType, setObjectType] = useState("tree");
  const [objectBlocksMovement, setObjectBlocksMovement] = useState(true);
  const [spawnType, setSpawnType] = useState<SpawnPoint["type"]>("player");
  const [stateText, setStateText] = useState("{}");

  const validation = useMemo(() => validateWorld(world), [world]);
  const selectedObject =
    selection?.type === "object" ? world.objects.find((object) => object.id === selection.id) : undefined;
  const selectedSpawn = selection?.type === "spawn" ? world.spawns[selection.index] : undefined;

  async function handleNew() {
    setWorld(createBlankWorld());
    setFileHandle(null);
    setFileName("unsaved world");
    setSelection(null);
    setDirty(false);
    setStatus("created new world");
  }

  async function handleOpen() {
    try {
      const opened = await openWorldFile();
      const parsed = normalizeWorld(JSON.parse(opened.text));
      setWorld(parsed);
      setFileHandle(opened.handle);
      setFileName((await opened.handle.getFile()).name);
      setSelection(null);
      setDirty(false);
      setStatus("opened world");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  async function handleSave() {
    try {
      if (!validation.valid) {
        setStatus("fix validation errors before saving");
        return;
      }
      if (fileHandle) {
        await saveWorldFile(fileHandle, world);
      } else {
        const nextHandle = await saveWorldFileAs(world);
        setFileHandle(nextHandle);
        setFileName((await nextHandle.getFile()).name);
      }
      setDirty(false);
      setStatus("saved world");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  async function handleSaveAs() {
    try {
      if (!validation.valid) {
        setStatus("fix validation errors before saving");
        return;
      }
      const nextHandle = await saveWorldFileAs(world);
      setFileHandle(nextHandle);
      setFileName((await nextHandle.getFile()).name);
      setDirty(false);
      setStatus("saved world");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  function updateWorld(updater: (current: WorldFormat) => WorldFormat) {
    setWorld((current) => updater(current));
    setDirty(true);
  }

  function handleTileClick(x: number, y: number) {
    const index = tileIndex(world.size, x, y);

    if (tool === "terrain") {
      updateWorld((current) => {
        const terrain = [...current.terrain];
        terrain[index] = terrainType.trim() || "grass";
        return { ...current, terrain };
      });
      setSelection({ type: "tile", x, y });
      return;
    }

    if (tool === "collision") {
      updateWorld((current) => {
        const collision = current.collision
          ? [...current.collision]
          : new Array(current.size.x * current.size.y).fill(false);
        collision[index] = collisionValue;
        return { ...current, collision };
      });
      setSelection({ type: "tile", x, y });
      return;
    }

    if (tool === "object") {
      const object = createObject(world, objectType, x, y, objectBlocksMovement);
      updateWorld((current) => ({ ...current, objects: [...current.objects, object] }));
      setSelection({ type: "object", id: object.id });
      setStateText(JSON.stringify(object.state ?? {}, null, 2));
      return;
    }

    if (tool === "spawn") {
      const spawn: SpawnPoint = { type: spawnType, x, y };
      updateWorld((current) => ({ ...current, spawns: [...current.spawns, spawn] }));
      setSelection({ type: "spawn", index: world.spawns.length });
      return;
    }

    selectTile(x, y);
  }

  function selectTile(x: number, y: number) {
    const object = [...world.objects]
      .reverse()
      .find((candidate) => coversTile(candidate, x, y));
    if (object) {
      setSelection({ type: "object", id: object.id });
      setStateText(JSON.stringify(object.state ?? {}, null, 2));
      return;
    }

    const spawnIndex = world.spawns.findIndex((spawn) => spawn.x === x && spawn.y === y);
    if (spawnIndex >= 0) {
      setSelection({ type: "spawn", index: spawnIndex });
      return;
    }

    setSelection({ type: "tile", x, y });
  }

  function updateMetadata<K extends keyof Pick<WorldFormat, "id" | "displayName">>(
    key: K,
    value: WorldFormat[K]
  ) {
    updateWorld((current) => ({ ...current, [key]: value }));
  }

  function applyResize(width: number, height: number) {
    const x = Math.max(1, Math.floor(width || 1));
    const y = Math.max(1, Math.floor(height || 1));
    updateWorld((current) => resizeWorld(current, { x, y }, terrainType));
    setSelection(null);
  }

  function updateSelectedObject(patch: Partial<WorldObject>) {
    if (!selectedObject) {
      return;
    }
    updateWorld((current) => ({
      ...current,
      objects: current.objects.map((object) =>
        object.id === selectedObject.id ? { ...object, ...patch } : object
      ),
    }));
    if (patch.id) {
      setSelection({ type: "object", id: patch.id });
    }
  }

  function updateSelectedSpawn(patch: Partial<SpawnPoint>) {
    if (selection?.type !== "spawn") {
      return;
    }
    updateWorld((current) => ({
      ...current,
      spawns: current.spawns.map((spawn, index) =>
        index === selection.index ? { ...spawn, ...patch } : spawn
      ),
    }));
  }

  function deleteSelection() {
    if (selection?.type === "object") {
      updateWorld((current) => ({
        ...current,
        objects: current.objects.filter((object) => object.id !== selection.id),
      }));
    } else if (selection?.type === "spawn") {
      updateWorld((current) => ({
        ...current,
        spawns: current.spawns.filter((_, index) => index !== selection.index),
      }));
    }
    setSelection(null);
  }

  function applyObjectState() {
    if (!selectedObject) {
      return;
    }
    try {
      const parsed = JSON.parse(stateText) as unknown;
      if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
        setStatus("object state must be a JSON object");
        return;
      }
      updateSelectedObject({ state: parsed as Record<string, unknown> });
      setStatus("object state updated");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  return (
    <div className="appShell">
      <header className="topBar">
        <div>
          <h1>World Editor</h1>
          <p>
            {fileName}
            {dirty ? " *" : ""}
          </p>
        </div>
        <div className="fileActions">
          <button type="button" onClick={handleNew}>New</button>
          <button type="button" onClick={handleOpen} disabled={!supportsFileSystemAccess()}>Open</button>
          <button type="button" onClick={handleSave} disabled={!supportsFileSystemAccess()}>Save</button>
          <button type="button" onClick={handleSaveAs} disabled={!supportsFileSystemAccess()}>Save As</button>
        </div>
      </header>

      {!supportsFileSystemAccess() ? (
        <div className="unsupported">
          This editor needs a Chromium browser on localhost for direct file read/write.
        </div>
      ) : null}

      <main className="workspace">
        <aside className="panel">
          <section>
            <h2>World</h2>
            <label>
              Id
              <input value={world.id} onChange={(event) => updateMetadata("id", event.target.value)} />
            </label>
            <label>
              Name
              <input
                value={world.displayName ?? ""}
                onChange={(event) => updateMetadata("displayName", event.target.value)}
              />
            </label>
            <div className="fieldRow">
              <label>
                Width
                <input
                  type="number"
                  min="1"
                  value={world.size.x}
                  onChange={(event) => applyResize(Number(event.target.value), world.size.y)}
                />
              </label>
              <label>
                Height
                <input
                  type="number"
                  min="1"
                  value={world.size.y}
                  onChange={(event) => applyResize(world.size.x, Number(event.target.value))}
                />
              </label>
            </div>
          </section>

          <section>
            <h2>Tools</h2>
            <div className="segmented">
              {(["terrain", "collision", "object", "spawn", "select"] as Tool[]).map((item) => (
                <button
                  key={item}
                  type="button"
                  className={tool === item ? "active" : ""}
                  onClick={() => setTool(item)}
                >
                  {item}
                </button>
              ))}
            </div>

            {tool === "terrain" ? (
              <div className="toolSettings">
                <label>
                  Terrain
                  <input value={terrainType} onChange={(event) => setTerrainType(event.target.value)} />
                </label>
                <div className="swatches">
                  {TERRAIN_SWATCHES.map((terrain) => (
                    <button
                      key={terrain}
                      type="button"
                      title={terrain}
                      aria-label={terrain}
                      className={terrainType === terrain ? "selectedSwatch" : ""}
                      style={{ backgroundColor: terrainColor(terrain) }}
                      onClick={() => setTerrainType(terrain)}
                    />
                  ))}
                </div>
              </div>
            ) : null}

            {tool === "collision" ? (
              <div className="toolSettings">
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={collisionValue}
                    onChange={(event) => setCollisionValue(event.target.checked)}
                  />
                  blocked
                </label>
              </div>
            ) : null}

            {tool === "object" ? (
              <div className="toolSettings">
                <label>
                  Object type
                  <input value={objectType} onChange={(event) => setObjectType(event.target.value)} />
                </label>
                <div className="quickPicks">
                  {OBJECT_TYPES.map((type) => (
                    <button key={type} type="button" onClick={() => setObjectType(type)}>
                      {type}
                    </button>
                  ))}
                </div>
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={objectBlocksMovement}
                    onChange={(event) => setObjectBlocksMovement(event.target.checked)}
                  />
                  blocks movement
                </label>
              </div>
            ) : null}

            {tool === "spawn" ? (
              <div className="toolSettings">
                <label>
                  Spawn type
                  <select
                    value={spawnType}
                    onChange={(event) => setSpawnType(event.target.value as SpawnPoint["type"])}
                  >
                    {SPAWN_TYPES.map((type) => (
                      <option key={type} value={type}>{type}</option>
                    ))}
                  </select>
                </label>
              </div>
            ) : null}
          </section>

          <section>
            <h2>Validation</h2>
            {validation.valid ? <p className="ok">valid</p> : null}
            {validation.errors.map((error) => (
              <p key={error} className="error">{error}</p>
            ))}
          </section>
        </aside>

        <section className="canvasPanel">
          <div className="gridWrap">
            <div
              className="grid"
              style={{ gridTemplateColumns: `repeat(${world.size.x}, 32px)` }}
            >
              {world.terrain.map((terrain, index) => {
                const x = index % world.size.x;
                const y = Math.floor(index / world.size.x);
                const object = world.objects.find((candidate) => coversTile(candidate, x, y));
                const spawn = world.spawns.find((candidate) => candidate.x === x && candidate.y === y);
                const blocked = Boolean(world.collision?.[index]);
                const selected = isTileSelected(selection, x, y, object, spawn, world.spawns);

                return (
                  <button
                    key={`${x}:${y}`}
                    type="button"
                    className={`tile ${blocked ? "blocked" : ""} ${selected ? "selectedTile" : ""}`}
                    style={{ backgroundColor: terrainColor(terrain) }}
                    title={`${x}, ${y}: ${terrain}`}
                    onClick={() => handleTileClick(x, y)}
                  >
                    {object ? <span className="objectBadge">{object.type.slice(0, 2)}</span> : null}
                    {spawn ? <span className="spawnBadge">{spawn.type.slice(0, 1)}</span> : null}
                  </button>
                );
              })}
            </div>
          </div>
        </section>

        <aside className="panel">
          <section>
            <h2>Selection</h2>
            {selection?.type === "tile" ? (
              <TileDetails world={world} x={selection.x} y={selection.y} />
            ) : null}

            {selectedObject ? (
              <div className="details">
                <label>
                  Id
                  <input value={selectedObject.id} onChange={(event) => updateSelectedObject({ id: event.target.value })} />
                </label>
                <label>
                  Type
                  <input value={selectedObject.type} onChange={(event) => updateSelectedObject({ type: event.target.value })} />
                </label>
                <div className="fieldRow">
                  <label>
                    X
                    <input type="number" value={selectedObject.x} onChange={(event) => updateSelectedObject({ x: Number(event.target.value) })} />
                  </label>
                  <label>
                    Y
                    <input type="number" value={selectedObject.y} onChange={(event) => updateSelectedObject({ y: Number(event.target.value) })} />
                  </label>
                </div>
                <div className="fieldRow">
                  <label>
                    Width
                    <input type="number" min="1" value={selectedObject.width ?? 1} onChange={(event) => updateSelectedObject({ width: Number(event.target.value) })} />
                  </label>
                  <label>
                    Height
                    <input type="number" min="1" value={selectedObject.height ?? 1} onChange={(event) => updateSelectedObject({ height: Number(event.target.value) })} />
                  </label>
                </div>
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={Boolean(selectedObject.blocksMovement)}
                    onChange={(event) => updateSelectedObject({ blocksMovement: event.target.checked })}
                  />
                  blocks movement
                </label>
                <label>
                  Interactions
                  <input
                    value={(selectedObject.interactable ?? []).join(", ")}
                    onChange={(event) =>
                      updateSelectedObject({
                        interactable: event.target.value
                          .split(",")
                          .map((value) => value.trim())
                          .filter(Boolean),
                      })
                    }
                  />
                </label>
                <label>
                  State JSON
                  <textarea value={stateText} onChange={(event) => setStateText(event.target.value)} />
                </label>
                <div className="buttonRow">
                  <button type="button" onClick={applyObjectState}>Apply State</button>
                  <button type="button" className="danger" onClick={deleteSelection}>Delete</button>
                </div>
              </div>
            ) : null}

            {selectedSpawn ? (
              <div className="details">
                <label>
                  Type
                  <select
                    value={selectedSpawn.type}
                    onChange={(event) => updateSelectedSpawn({ type: event.target.value as SpawnPoint["type"] })}
                  >
                    {SPAWN_TYPES.map((type) => (
                      <option key={type} value={type}>{type}</option>
                    ))}
                  </select>
                </label>
                <div className="fieldRow">
                  <label>
                    X
                    <input type="number" value={selectedSpawn.x} onChange={(event) => updateSelectedSpawn({ x: Number(event.target.value) })} />
                  </label>
                  <label>
                    Y
                    <input type="number" value={selectedSpawn.y} onChange={(event) => updateSelectedSpawn({ y: Number(event.target.value) })} />
                  </label>
                </div>
                <label>
                  Entity type
                  <input value={selectedSpawn.entityType ?? ""} onChange={(event) => updateSelectedSpawn({ entityType: event.target.value || undefined })} />
                </label>
                <label>
                  Name
                  <input value={selectedSpawn.name ?? ""} onChange={(event) => updateSelectedSpawn({ name: event.target.value || undefined })} />
                </label>
                <button type="button" className="danger" onClick={deleteSelection}>Delete</button>
              </div>
            ) : null}

            {!selection ? <p className="muted">nothing selected</p> : null}
          </section>

          <section>
            <h2>Status</h2>
            <p className="status">{status}</p>
          </section>
        </aside>
      </main>
    </div>
  );
}

function TileDetails({ world, x, y }: { world: WorldFormat; x: number; y: number }) {
  const index = tileIndex(world.size, x, y);
  return (
    <div className="details">
      <p className="metric">x {x}</p>
      <p className="metric">y {y}</p>
      <p className="metric">terrain {world.terrain[index]}</p>
      <p className="metric">blocked {world.collision?.[index] ? "yes" : "no"}</p>
    </div>
  );
}

function createObject(
  world: WorldFormat,
  rawType: string,
  x: number,
  y: number,
  blocksMovement: boolean
): WorldObject {
  const type = sanitizeToken(rawType || "object");
  let nextNumber = 1;
  let id = `${type}_${String(nextNumber).padStart(3, "0")}`;

  while (world.objects.some((object) => object.id === id)) {
    nextNumber += 1;
    id = `${type}_${String(nextNumber).padStart(3, "0")}`;
  }

  return {
    id,
    type,
    x,
    y,
    width: 1,
    height: 1,
    blocksMovement,
    interactable: [],
    state: {},
  };
}

function coversTile(object: WorldObject, x: number, y: number): boolean {
  const width = object.width ?? 1;
  const height = object.height ?? 1;
  return x >= object.x && y >= object.y && x < object.x + width && y < object.y + height;
}

function isTileSelected(
  selection: Selection,
  x: number,
  y: number,
  object: WorldObject | undefined,
  spawn: SpawnPoint | undefined,
  spawns: SpawnPoint[]
): boolean {
  if (selection?.type === "tile") {
    return selection.x === x && selection.y === y;
  }
  if (selection?.type === "object") {
    return object?.id === selection.id;
  }
  if (selection?.type === "spawn") {
    return Boolean(spawn && spawns[selection.index] === spawn);
  }
  return false;
}

function terrainColor(terrain: string): string {
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

function sanitizeToken(value: string): string {
  const token = value.trim().toLowerCase().replace(/[^a-z0-9_-]+/g, "_").replace(/^_+|_+$/g, "");
  return token || "object";
}

function errorMessage(error: unknown): string {
  if (error instanceof DOMException && error.name === "AbortError") {
    return "file action canceled";
  }
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

export default App;
