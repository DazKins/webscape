import { useMemo, useState } from "react";
import {
  createBlankConversationDocument,
  createBlankConversationNode,
  createBlankConversationOption,
  sanitizeConversationId,
  validateConversationDocument,
  type Conversation,
  type ConversationDocument,
  type ConversationMessage,
  type ConversationNode,
  type ConversationOption,
  type ValidationResult as ConversationValidationResult,
} from "./conversationFormat.ts";
import {
  conversationPathFromFilename,
  createBlankGameProject,
  DEFAULT_WORLD_PATH,
  filenameFromProjectPath,
  isValidProjectPath,
  isValidProjectFilename,
  mapPathFromFilename,
  validateGameProject,
  type GameProject,
} from "./gameProject.ts";
import {
  createBlankWorld,
  resizeWorld,
  tileIndex,
  validateWorld,
  type SpawnPoint,
  type WorldFormat,
  type WorldWall,
  type WorldObject,
  type ValidationResult as WorldValidationResult,
} from "./worldFormat.ts";
import {
  openGameProjectFolder,
  saveGameProjectFolder,
  saveGameProjectFolderAs,
  supportsProjectFolders,
  type ConversationDocuments,
  type ProjectDirectoryHandle,
  type WorldDocuments,
} from "./fileSystem.ts";

type Tool = "terrain" | "blocker" | "wall" | "object" | "spawn" | "select";
type EditorTab = "map" | "conversations" | "quests";
type Selection =
  | { type: "wall"; id: string }
  | { type: "object"; id: string }
  | { type: "spawn"; index: number }
  | { type: "tile"; x: number; y: number }
  | null;

const TERRAIN_SWATCHES = ["grass", "dirt", "road", "water", "stone"];
const WALL_TYPES = ["stone", "wood"];
const OBJECT_TYPES = ["tree", "door", "building", "chest", "rock"];
const SPAWN_TYPES: SpawnPoint["type"][] = ["player", "npc", "monster"];

function App() {
  const defaultWorld = createBlankWorld();
  const defaultWorlds = { [DEFAULT_WORLD_PATH]: defaultWorld };
  const [project, setProject] = useState<GameProject>(() => createBlankGameProject());
  const [world, setWorld] = useState<WorldFormat>(() => defaultWorld);
  const [worldDocuments, setWorldDocuments] = useState<WorldDocuments>(() => defaultWorlds);
  const [deletedWorldPaths, setDeletedWorldPaths] = useState<string[]>([]);
  const [conversationDocuments, setConversationDocuments] = useState<ConversationDocuments>({});
  const [deletedConversationPaths, setDeletedConversationPaths] = useState<string[]>([]);
  const [projectHandle, setProjectHandle] = useState<ProjectDirectoryHandle | null>(null);
  const [projectFolderName, setProjectFolderName] = useState<string>("unsaved project");
  const [worldPath, setWorldPath] = useState<string>(DEFAULT_WORLD_PATH);
  const [status, setStatus] = useState<string>("ready");
  const [dirty, setDirty] = useState(false);
  const [activeTab, setActiveTab] = useState<EditorTab>("map");
  const [tool, setTool] = useState<Tool>("terrain");
  const [selection, setSelection] = useState<Selection>(null);
  const [terrainType, setTerrainType] = useState("grass");
  const [blockerValue, setBlockerValue] = useState(true);
  const [wallType, setWallType] = useState("stone");
  const [objectType, setObjectType] = useState("tree");
  const [objectBlocksMovement, setObjectBlocksMovement] = useState(true);
  const [spawnType, setSpawnType] = useState<SpawnPoint["type"]>("player");
  const [stateText, setStateText] = useState("{}");
  const [mapNameInput, setMapNameInput] = useState("New Map");
  const [conversationNameInput, setConversationNameInput] = useState("New Conversation");
  const [selectedConversationPath, setSelectedConversationPath] = useState<string | null>(null);
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  const projectValidation = useMemo(() => validateGameProject(project), [project]);
  const worldValidations = useMemo(() => validateWorldDocuments(worldDocuments), [worldDocuments]);
  const conversationValidations = useMemo(() => validateConversationDocuments(conversationDocuments), [conversationDocuments]);
  const worldPathValid = isValidProjectPath(worldPath);
  const worldsValid = worldValidations.every((validation) => validation.result.valid);
  const conversationsValid = conversationValidations.every((validation) => validation.result.valid);
  const canSave = projectValidation.valid && worldsValid && conversationsValid && worldPathValid;
  const mapSummaries = useMemo(() => summarizeWorlds(project, worldDocuments), [project, worldDocuments]);
  const currentMapName = world.displayName || world.id || "Untitled Map";
  const conversationSummaries = useMemo(() => summarizeConversations(project, conversationDocuments), [project, conversationDocuments]);
  const selectedConversationDocument = selectedConversationPath ? conversationDocuments[selectedConversationPath] : undefined;
  const selectedConversation = selectedConversationDocument?.conversations.find(
    (conversation) => conversation.id === selectedConversationId
  );
  const selectedNode = selectedConversation?.nodes.find((node) => node.id === selectedNodeId);
  const selectedWall =
    selection?.type === "wall" ? world.walls.find((wall) => wall.id === selection.id) : undefined;
  const selectedObject =
    selection?.type === "object" ? world.objects.find((object) => object.id === selection.id) : undefined;
  const selectedSpawn = selection?.type === "spawn" ? world.spawns[selection.index] : undefined;

  async function handleNew() {
    const nextWorld = createBlankWorld();
    setProject(createBlankGameProject());
    setWorld(nextWorld);
    setWorldDocuments({ [DEFAULT_WORLD_PATH]: nextWorld });
    setDeletedWorldPaths([]);
    setConversationDocuments({});
    setDeletedConversationPaths([]);
    setProjectHandle(null);
    setProjectFolderName("unsaved project");
    setWorldPath(DEFAULT_WORLD_PATH);
    setSelection(null);
    setSelectedConversationPath(null);
    setSelectedConversationId(null);
    setSelectedNodeId(null);
    setDirty(false);
    setStatus("created new project");
  }

  async function handleOpen() {
    try {
      const opened = await openGameProjectFolder();
      setProject(opened.project);
      setWorld(opened.world);
      setWorldDocuments(opened.worlds);
      setDeletedWorldPaths([]);
      setConversationDocuments(opened.conversations);
      setDeletedConversationPaths([]);
      setProjectHandle(opened.handle);
      setProjectFolderName(opened.handle.name);
      setWorldPath(opened.worldPath);
      setSelection(null);
      const firstConversation = summarizeConversations(opened.project, opened.conversations)[0];
      setSelectedConversationPath(firstConversation?.path ?? null);
      setSelectedConversationId(firstConversation?.conversationId ?? null);
      setSelectedNodeId(firstConversation?.startNodeId ?? null);
      setDirty(false);
      setStatus("opened project");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  async function handleSave() {
    try {
      if (!canSave) {
        setStatus("fix validation errors before saving");
        return;
      }
      if (projectHandle) {
        const savedProject = await saveGameProjectFolder(
          projectHandle,
          project,
          worldDocuments,
          worldPath,
          conversationDocuments,
          deletedWorldPaths,
          deletedConversationPaths
        );
        setProject(savedProject);
        setDeletedWorldPaths([]);
        setDeletedConversationPaths([]);
      } else {
        const saved = await saveGameProjectFolderAs(project, worldDocuments, worldPath, conversationDocuments);
        setProject(saved.project);
        setProjectHandle(saved.handle);
        setProjectFolderName(saved.handle.name);
        setDeletedWorldPaths([]);
        setDeletedConversationPaths([]);
      }
      setDirty(false);
      setStatus("saved project");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  async function handleSaveAs() {
    try {
      if (!canSave) {
        setStatus("fix validation errors before saving");
        return;
      }
      const saved = await saveGameProjectFolderAs(project, worldDocuments, worldPath, conversationDocuments);
      setProject(saved.project);
      setProjectHandle(saved.handle);
      setProjectFolderName(saved.handle.name);
      setDeletedWorldPaths([]);
      setDeletedConversationPaths([]);
      setDirty(false);
      setStatus("saved project");
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }

  function updateProject<K extends keyof Pick<GameProject, "id" | "displayName">>(
    key: K,
    value: GameProject[K]
  ) {
    setProject((current) => ({ ...current, [key]: value }));
    setDirty(true);
  }

  function createMap() {
    const name = mapNameInput.trim();
    const id = sanitizeConversationId(name);
    const filename = `${id}.json`;
    if (!isValidProjectFilename(filename)) {
      setStatus("map name is invalid");
      return;
    }
    const path = mapPathFromFilename(filename);
    if (project.files.maps.includes(path)) {
      setStatus("map already exists");
      return;
    }

    const nextWorld = {
      ...createBlankWorld(),
      id,
      displayName: name,
    };
    setProject((current) => ({
      ...current,
      files: {
        ...current.files,
        maps: [...current.files.maps, path],
      },
    }));
    setWorldDocuments((current) => ({
      ...current,
      [path]: nextWorld,
    }));
    setDeletedWorldPaths((current) => current.filter((candidate) => candidate !== path));
    setWorld(nextWorld);
    setWorldPath(path);
    setSelection(null);
    setDirty(true);
    setStatus("map created");
  }

  function selectMap(path: string) {
    const nextWorld = worldDocuments[path];
    if (!nextWorld) {
      setStatus("map could not be opened");
      return;
    }

    setWorld(nextWorld);
    setWorldPath(path);
    setSelection(null);
  }

  function deleteMap(path: string) {
    if (project.files.maps.length <= 1) {
      setStatus("project must have at least one map");
      return;
    }

    const nextPath = project.files.maps.find((candidate) => candidate !== path);
    setProject((current) => ({
      ...current,
      files: {
        ...current.files,
        maps: current.files.maps.filter((candidate) => candidate !== path),
      },
    }));
    setWorldDocuments((current) => {
      const next = { ...current };
      delete next[path];
      return next;
    });
    setDeletedWorldPaths((current) => [...new Set([...current, path])]);
    if (worldPath === path && nextPath && worldDocuments[nextPath]) {
      setWorld(worldDocuments[nextPath]);
      setWorldPath(nextPath);
      setSelection(null);
    }
    setDirty(true);
    setStatus("map deleted");
  }

  function createConversation() {
    const name = conversationNameInput.trim();
    const id = sanitizeConversationId(name);
    const filename = `${id}.json`;
    if (!isValidProjectFilename(filename)) {
      setStatus("conversation name is invalid");
      return;
    }
    const path = conversationPathFromFilename(filename);
    if (project.files.conversations.includes(path)) {
      setStatus("conversation already exists");
      return;
    }

    setProject((current) => {
      return {
        ...current,
        files: {
          ...current.files,
          conversations: [...current.files.conversations, path],
        },
      };
    });
    setConversationDocuments((current) => ({
      ...current,
      [path]: {
        ...createBlankConversationDocument(id),
        displayName: name,
      },
    }));
    setDeletedConversationPaths((current) => current.filter((candidate) => candidate !== path));
    setSelectedConversationPath(path);
    setSelectedConversationId(id);
    setSelectedNodeId("start");
    setDirty(true);
    setStatus("conversation created");
  }

  function deleteConversation(path: string, conversationId: string) {
    const document = conversationDocuments[path];
    const removesWholeDocument = !document || document.conversations.length <= 1;

    setProject((current) => ({
      ...current,
      files: {
        ...current.files,
        conversations: removesWholeDocument
          ? current.files.conversations.filter((candidate) => candidate !== path)
          : current.files.conversations,
      },
    }));
    setConversationDocuments((current) => {
      if (removesWholeDocument) {
        const next = { ...current };
        delete next[path];
        return next;
      }

      return {
        ...current,
        [path]: {
          ...document,
          conversations: document.conversations.filter((conversation) => conversation.id !== conversationId),
        },
      };
    });
    if (removesWholeDocument) {
      setDeletedConversationPaths((current) => [...new Set([...current, path])]);
    }
    if (selectedConversationPath === path && selectedConversationId === conversationId) {
      const nextConversation = conversationSummaries.find(
        (summary) => summary.path !== path || summary.conversationId !== conversationId
      );
      setSelectedConversationPath(nextConversation?.path ?? null);
      setSelectedConversationId(nextConversation?.conversationId ?? null);
      setSelectedNodeId(nextConversation?.startNodeId ?? null);
    }
    setDirty(true);
    setStatus("conversation deleted");
  }

  function selectConversation(path: string, conversationId: string) {
    const conversation = conversationDocuments[path]?.conversations.find((candidate) => candidate.id === conversationId);
    setSelectedConversationPath(path);
    setSelectedConversationId(conversationId);
    setSelectedNodeId(conversation?.startNodeId ?? conversation?.nodes[0]?.id ?? null);
  }

  function updateSelectedConversation(updater: (conversation: Conversation) => Conversation) {
    if (!selectedConversationPath || !selectedConversationId) {
      return;
    }

    setConversationDocuments((current) => {
      const document = current[selectedConversationPath];
      if (!document) {
        return current;
      }

      let nextSelectedId = selectedConversationId;
      const conversations = document.conversations.map((conversation) => {
        if (conversation.id !== selectedConversationId) {
          return conversation;
        }
        const nextConversation = updater(conversation);
        nextSelectedId = nextConversation.id;
        return nextConversation;
      });
      setSelectedConversationId(nextSelectedId);
      return {
        ...current,
        [selectedConversationPath]: {
          ...document,
          conversations,
        },
      };
    });
    setDirty(true);
  }

  function updateSelectedConversationDocument(patch: Partial<ConversationDocument>) {
    if (!selectedConversationPath) {
      return;
    }

    setConversationDocuments((current) => {
      const document = current[selectedConversationPath];
      if (!document) {
        return current;
      }

      return {
        ...current,
        [selectedConversationPath]: {
          ...document,
          ...patch,
        },
      };
    });
    setDirty(true);
  }

  function addConversationNode() {
    updateSelectedConversation((conversation) => {
      const node = createBlankConversationNode(conversation.nodes);
      setSelectedNodeId(node.id);
      return {
        ...conversation,
        nodes: [...conversation.nodes, node],
      };
    });
  }

  function deleteConversationNode(nodeId: string) {
    updateSelectedConversation((conversation) => {
      if (conversation.nodes.length <= 1) {
        return conversation;
      }

      const nodes = conversation.nodes
        .filter((node) => node.id !== nodeId)
        .map((node) => ({
          ...node,
          options: node.options?.filter((option) => option.nextNodeId !== nodeId),
        }));
      const nextStartNodeId = conversation.startNodeId === nodeId ? nodes[0].id : conversation.startNodeId;
      if (selectedNodeId === nodeId) {
        setSelectedNodeId(nextStartNodeId);
      }
      return {
        ...conversation,
        startNodeId: nextStartNodeId,
        nodes,
      };
    });
  }

  function updateSelectedNode(updater: (node: ConversationNode, conversation: Conversation) => ConversationNode) {
    if (!selectedNodeId) {
      return;
    }

    updateSelectedConversation((conversation) => ({
      ...conversation,
      nodes: conversation.nodes.map((node) =>
        node.id === selectedNodeId ? updater(node, conversation) : node
      ),
    }));
  }

  function renameSelectedNode(nextId: string) {
    if (!selectedNodeId) {
      return;
    }
    const previousId = selectedNodeId;
    updateSelectedConversation((conversation) => ({
      ...conversation,
      startNodeId: conversation.startNodeId === previousId ? nextId : conversation.startNodeId,
      nodes: conversation.nodes.map((node) => ({
        ...node,
        id: node.id === previousId ? nextId : node.id,
        options: node.options?.map((option) => ({
          ...option,
          nextNodeId: option.nextNodeId === previousId ? nextId : option.nextNodeId,
        })),
      })),
    }));
    setSelectedNodeId(nextId);
  }

  function updateMessage(index: number, patch: Partial<ConversationMessage>) {
    updateSelectedNode((node) => ({
      ...node,
      messages: node.messages.map((message, messageIndex) =>
        messageIndex === index ? { ...message, ...patch } : message
      ),
    }));
  }

  function addMessage() {
    updateSelectedNode((node) => ({
      ...node,
      messages: [...node.messages, { speaker: "npc", text: "" }],
    }));
  }

  function deleteMessage(index: number) {
    updateSelectedNode((node) => ({
      ...node,
      messages: node.messages.filter((_, messageIndex) => messageIndex !== index),
    }));
  }

  function setNodeEndsConversation(endsConversation: boolean) {
    updateSelectedNode((node) => ({
      ...node,
      endConversation: endsConversation,
      options: endsConversation ? undefined : node.options ?? [],
    }));
  }

  function addOption() {
    updateSelectedNode((node, conversation) => ({
      ...node,
      endConversation: false,
      options: [
        ...(node.options ?? []),
        createBlankConversationOption(node.options ?? [], conversation.nodes[0]?.id ?? node.id),
      ],
    }));
  }

  function updateOption(index: number, patch: Partial<ConversationOption>) {
    updateSelectedNode((node) => ({
      ...node,
      options: (node.options ?? []).map((option, optionIndex) =>
        optionIndex === index ? { ...option, ...patch } : option
      ),
    }));
  }

  function deleteOption(index: number) {
    updateSelectedNode((node) => ({
      ...node,
      options: (node.options ?? []).filter((_, optionIndex) => optionIndex !== index),
    }));
  }

  function updateWorld(updater: (current: WorldFormat) => WorldFormat) {
    setWorld((current) => {
      const nextWorld = updater(current);
      setWorldDocuments((documents) => ({
        ...documents,
        [worldPath]: nextWorld,
      }));
      return nextWorld;
    });
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

    if (tool === "blocker") {
      updateWorld((current) => {
        const blockers = current.blockers
          ? [...current.blockers]
          : new Array(current.size.x * current.size.y).fill(false);
        blockers[index] = blockerValue;
        return { ...current, blockers };
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

    if (tool === "wall") {
      const wall = createWall(world, wallType, x, y);
      updateWorld((current) => ({ ...current, walls: [...current.walls, wall] }));
      setSelection({ type: "wall", id: wall.id });
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
    const wall = [...world.walls]
      .reverse()
      .find((candidate) => coversTile(candidate, x, y));
    if (wall) {
      setSelection({ type: "wall", id: wall.id });
      return;
    }

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

  function updateSelectedWall(patch: Partial<WorldWall>) {
    if (!selectedWall) {
      return;
    }
    updateWorld((current) => ({
      ...current,
      walls: current.walls.map((wall) => (wall.id === selectedWall.id ? { ...wall, ...patch } : wall)),
    }));
    if (patch.id) {
      setSelection({ type: "wall", id: patch.id });
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
    if (selection?.type === "wall") {
      updateWorld((current) => ({
        ...current,
        walls: current.walls.filter((wall) => wall.id !== selection.id),
      }));
    } else if (selection?.type === "object") {
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
        <div className="projectHeader">
          <h1>Game Editor</h1>
          <div className="projectHeaderFields">
            <label>
              Project id
              <input value={project.id} onChange={(event) => updateProject("id", event.target.value)} />
            </label>
            <label>
              Project name
              <input
                value={project.displayName ?? ""}
                onChange={(event) => updateProject("displayName", event.target.value)}
              />
            </label>
          </div>
          <p>
            {projectFolderName} / {currentMapName}
            {dirty ? " *" : ""}
            {" | "}
            maps {project.files.maps.length}
            {" | "}
            conversations {project.files.conversations.length}
            {" | "}
            quests {project.files.quests.length}
          </p>
        </div>
        <div className="fileActions">
          <button type="button" onClick={handleNew}>New</button>
          <button type="button" onClick={handleOpen} disabled={!supportsProjectFolders()}>Open Project</button>
          <button type="button" onClick={handleSave} disabled={!supportsProjectFolders()}>Save</button>
          <button type="button" onClick={handleSaveAs} disabled={!supportsProjectFolders()}>Save As</button>
        </div>
      </header>

      {!supportsProjectFolders() ? (
        <div className="unsupported">
          This editor needs a Chromium browser on localhost for project folder read/write.
        </div>
      ) : null}

      <nav className="editorTabs" aria-label="Editor sections">
        {(["map", "conversations", "quests"] as EditorTab[]).map((tab) => (
          <button
            key={tab}
            type="button"
            className={activeTab === tab ? "active" : ""}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </nav>

      {activeTab === "map" ? (
      <main className="workspace">
        <aside className="panel">
          <section>
            <h2>Maps</h2>
            <label>
              Name
              <input
                value={mapNameInput}
                onChange={(event) => setMapNameInput(event.target.value)}
              />
            </label>
            <button type="button" onClick={createMap}>Create Map</button>
          </section>

          <section>
            <h2>List</h2>
            {mapSummaries.length > 0 ? (
              <div className="stack">
                {mapSummaries.map((summary) => (
                  <div key={summary.path} className="listRow">
                    <button
                      type="button"
                      className={worldPath === summary.path ? "listButton active" : "listButton"}
                      onClick={() => selectMap(summary.path)}
                    >
                      {summary.name}
                    </button>
                    <button
                      type="button"
                      className="danger compactButton"
                      disabled={mapSummaries.length <= 1}
                      onClick={() => deleteMap(summary.path)}
                    >
                      Delete
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="muted">none</p>
            )}
          </section>

          <section>
            <h2>Map</h2>
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
              {(["terrain", "blocker", "wall", "object", "spawn", "select"] as Tool[]).map((item) => (
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

            {tool === "blocker" ? (
              <div className="toolSettings">
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={blockerValue}
                    onChange={(event) => setBlockerValue(event.target.checked)}
                  />
                  blocked
                </label>
              </div>
            ) : null}

            {tool === "wall" ? (
              <div className="toolSettings">
                <label>
                  Wall type
                  <input value={wallType} onChange={(event) => setWallType(event.target.value)} />
                </label>
                <div className="quickPicks">
                  {WALL_TYPES.map((type) => (
                    <button key={type} type="button" onClick={() => setWallType(type)}>
                      {type}
                    </button>
                  ))}
                </div>
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
            {canSave ? <p className="ok">valid</p> : null}
            {!worldPathValid ? <p className="error">map storage is invalid</p> : null}
            {projectValidation.errors.map((error) => (
              <p key={error} className="error">{error}</p>
            ))}
            {worldValidations.flatMap((validation) =>
              validation.result.errors.map((error) => (
                <p key={`${validation.path}:${error}`} className="error">{error}</p>
              ))
            )}
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
                const wall = world.walls.find((candidate) => coversTile(candidate, x, y));
                const object = world.objects.find((candidate) => coversTile(candidate, x, y));
                const spawn = world.spawns.find((candidate) => candidate.x === x && candidate.y === y);
                const blocked = Boolean(world.blockers?.[index]);
                const selected = isTileSelected(selection, x, y, wall, object, spawn, world.spawns);

                return (
                  <button
                    key={`${x}:${y}`}
                    type="button"
                    className={`tile ${blocked ? "blocked" : ""} ${selected ? "selectedTile" : ""}`}
                    style={{ backgroundColor: terrainColor(terrain) }}
                    title={`${x}, ${y}: ${terrain}`}
                    onClick={() => handleTileClick(x, y)}
                  >
                    {wall ? <span className="wallBadge">{wall.type.slice(0, 2)}</span> : null}
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

            {selectedWall ? (
              <div className="details">
                <label>
                  Id
                  <input value={selectedWall.id} onChange={(event) => updateSelectedWall({ id: event.target.value })} />
                </label>
                <label>
                  Type
                  <input value={selectedWall.type} onChange={(event) => updateSelectedWall({ type: event.target.value })} />
                </label>
                <div className="fieldRow">
                  <label>
                    X
                    <input type="number" value={selectedWall.x} onChange={(event) => updateSelectedWall({ x: Number(event.target.value) })} />
                  </label>
                  <label>
                    Y
                    <input type="number" value={selectedWall.y} onChange={(event) => updateSelectedWall({ y: Number(event.target.value) })} />
                  </label>
                </div>
                <button type="button" className="danger" onClick={deleteSelection}>Delete</button>
              </div>
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
      ) : activeTab === "conversations" ? (
        <ConversationsWorkspace
          project={project}
          conversations={conversationDocuments}
          summaries={conversationSummaries}
          selectedConversationPath={selectedConversationPath}
          selectedConversation={selectedConversation}
          selectedNode={selectedNode}
          nameInput={conversationNameInput}
          validations={conversationValidations}
          status={status}
          onNameInputChange={setConversationNameInput}
          onCreateConversation={createConversation}
          onDeleteConversation={deleteConversation}
          onSelectConversation={selectConversation}
          onUpdateConversationDocument={updateSelectedConversationDocument}
          onUpdateConversation={updateSelectedConversation}
          onAddNode={addConversationNode}
          onDeleteNode={deleteConversationNode}
          onSelectNode={setSelectedNodeId}
          onRenameNode={renameSelectedNode}
          onAddMessage={addMessage}
          onUpdateMessage={updateMessage}
          onDeleteMessage={deleteMessage}
          onSetNodeEndsConversation={setNodeEndsConversation}
          onAddOption={addOption}
          onUpdateOption={updateOption}
          onDeleteOption={deleteOption}
        />
      ) : (
        <QuestsWorkspace project={project} status={status} />
      )}
    </div>
  );
}

function ConversationsWorkspace({
  project,
  conversations,
  summaries,
  selectedConversationPath,
  selectedConversation,
  selectedNode,
  nameInput,
  validations,
  status,
  onNameInputChange,
  onCreateConversation,
  onDeleteConversation,
  onSelectConversation,
  onUpdateConversationDocument,
  onUpdateConversation,
  onAddNode,
  onDeleteNode,
  onSelectNode,
  onRenameNode,
  onAddMessage,
  onUpdateMessage,
  onDeleteMessage,
  onSetNodeEndsConversation,
  onAddOption,
  onUpdateOption,
  onDeleteOption,
}: {
  project: GameProject;
  conversations: ConversationDocuments;
  summaries: ConversationSummary[];
  selectedConversationPath: string | null;
  selectedConversation: Conversation | undefined;
  selectedNode: ConversationNode | undefined;
  nameInput: string;
  validations: ConversationValidationEntry[];
  status: string;
  onNameInputChange: (value: string) => void;
  onCreateConversation: () => void;
  onDeleteConversation: (path: string, conversationId: string) => void;
  onSelectConversation: (path: string, conversationId: string) => void;
  onUpdateConversationDocument: (patch: Partial<ConversationDocument>) => void;
  onUpdateConversation: (updater: (conversation: Conversation) => Conversation) => void;
  onAddNode: () => void;
  onDeleteNode: (nodeId: string) => void;
  onSelectNode: (nodeId: string) => void;
  onRenameNode: (nodeId: string) => void;
  onAddMessage: () => void;
  onUpdateMessage: (index: number, patch: Partial<ConversationMessage>) => void;
  onDeleteMessage: (index: number) => void;
  onSetNodeEndsConversation: (endsConversation: boolean) => void;
  onAddOption: () => void;
  onUpdateOption: (index: number, patch: Partial<ConversationOption>) => void;
  onDeleteOption: (index: number) => void;
}) {
  const selectedDocument = selectedConversationPath ? conversations[selectedConversationPath] : undefined;

  return (
    <main className="workspace singleWorkspace">
      <aside className="panel">
        <section>
          <h2>Conversations</h2>
          <label>
            Name
            <input value={nameInput} onChange={(event) => onNameInputChange(event.target.value)} />
          </label>
          <button type="button" onClick={onCreateConversation}>Create Conversation</button>
        </section>

        <section>
          <h2>List</h2>
          {summaries.length > 0 ? (
            <div className="stack">
              {summaries.map((summary) => (
                <button
                  key={`${summary.path}:${summary.conversationId}`}
                  type="button"
                  className={
                    selectedConversationPath === summary.path && selectedConversation?.id === summary.conversationId
                      ? "listButton active"
                      : "listButton"
                  }
                  onClick={() => onSelectConversation(summary.path, summary.conversationId)}
                >
                  {summary.name}
                </button>
              ))}
            </div>
          ) : (
            <p className="muted">none</p>
          )}
        </section>

        <section>
          <h2>Validation</h2>
          {validations.length === 0 || validations.every((validation) => validation.result.valid) ? (
            <p className="ok">valid</p>
          ) : null}
          {validations.flatMap((validation) =>
            validation.result.errors.map((error) => (
              <p key={`${validation.path}:${error}`} className="error">{error}</p>
            ))
          )}
        </section>

        <section>
          <h2>Status</h2>
          <p className="status">{status}</p>
        </section>
      </aside>

      <section className="contentPanel">
        <div className="contentHeader">
          <h2>Conversations</h2>
          <p>{project.files.conversations.length} conversations</p>
        </div>

        {selectedConversation && selectedDocument && selectedConversationPath ? (
          <div className="conversationEditor">
            <section className="editorSection">
              <div className="sectionHeader">
                <h3>Details</h3>
                <button
                  type="button"
                  className="danger"
                  onClick={() => onDeleteConversation(selectedConversationPath, selectedConversation.id)}
                >
                  Delete
                </button>
              </div>
              <div className="fieldRow">
                <label>
                  Name
                  <input
                    value={selectedDocument.displayName ?? ""}
                    onChange={(event) => onUpdateConversationDocument({ displayName: event.target.value })}
                  />
                </label>
                <label>
                  Id
                  <input
                    value={selectedConversation.id}
                    onChange={(event) =>
                      onUpdateConversation((conversation) => ({ ...conversation, id: event.target.value }))
                    }
                  />
                </label>
              </div>
              <label>
                Start
                <select
                  value={selectedConversation.startNodeId}
                  onChange={(event) =>
                    onUpdateConversation((conversation) => ({ ...conversation, startNodeId: event.target.value }))
                  }
                >
                  {selectedConversation.nodes.map((node) => (
                    <option key={node.id} value={node.id}>{node.id}</option>
                  ))}
                </select>
              </label>
            </section>

            <section className="editorSection">
              <div className="sectionHeader">
                <h3>Nodes</h3>
                <button type="button" onClick={onAddNode}>Add Node</button>
              </div>
              <div className="nodeTabs">
                {selectedConversation.nodes.map((node) => (
                  <button
                    key={node.id}
                    type="button"
                    className={selectedNode?.id === node.id ? "active" : ""}
                    onClick={() => onSelectNode(node.id)}
                  >
                    {node.id}
                  </button>
                ))}
              </div>
            </section>

            {selectedNode ? (
              <section className="editorSection">
                <div className="sectionHeader">
                  <h3>Node</h3>
                  <button
                    type="button"
                    className="danger"
                    disabled={selectedConversation.nodes.length <= 1}
                    onClick={() => onDeleteNode(selectedNode.id)}
                  >
                    Delete Node
                  </button>
                </div>

                <label>
                  Id
                  <input value={selectedNode.id} onChange={(event) => onRenameNode(event.target.value)} />
                </label>

                <div className="sectionHeader compact">
                  <h3>Messages</h3>
                  <button type="button" onClick={onAddMessage}>Add Message</button>
                </div>
                <div className="stack">
                  {selectedNode.messages.map((message, index) => (
                    <div key={index} className="subEditor">
                      <div className="fieldRow">
                        <label>
                          Speaker
                          <input
                            value={message.speaker ?? ""}
                            onChange={(event) => onUpdateMessage(index, { speaker: event.target.value || undefined })}
                          />
                        </label>
                        <label>
                          Portrait
                          <input
                            value={message.portrait ?? ""}
                            onChange={(event) => onUpdateMessage(index, { portrait: event.target.value || undefined })}
                          />
                        </label>
                      </div>
                      <label>
                        Text
                        <textarea value={message.text} onChange={(event) => onUpdateMessage(index, { text: event.target.value })} />
                      </label>
                      <button type="button" className="danger" onClick={() => onDeleteMessage(index)}>
                        Delete Message
                      </button>
                    </div>
                  ))}
                </div>

                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={Boolean(selectedNode.endConversation)}
                    onChange={(event) => onSetNodeEndsConversation(event.target.checked)}
                  />
                  ends conversation
                </label>

                {!selectedNode.endConversation ? (
                  <>
                    <div className="sectionHeader compact">
                      <h3>Options</h3>
                      <button type="button" onClick={onAddOption}>Add Option</button>
                    </div>
                    <div className="stack">
                      {(selectedNode.options ?? []).map((option, index) => (
                        <div key={index} className="subEditor">
                          <div className="fieldRow">
                            <label>
                              Id
                              <input
                                value={option.id}
                                onChange={(event) => onUpdateOption(index, { id: event.target.value })}
                              />
                            </label>
                            <label>
                              Next
                              <select
                                value={option.nextNodeId}
                                onChange={(event) => onUpdateOption(index, { nextNodeId: event.target.value })}
                              >
                                {selectedConversation.nodes.map((node) => (
                                  <option key={node.id} value={node.id}>{node.id}</option>
                                ))}
                              </select>
                            </label>
                          </div>
                          <label>
                            Text
                            <input
                              value={option.text}
                              onChange={(event) => onUpdateOption(index, { text: event.target.value })}
                            />
                          </label>
                          <button type="button" className="danger" onClick={() => onDeleteOption(index)}>
                            Delete Option
                          </button>
                        </div>
                      ))}
                    </div>
                  </>
                ) : null}
              </section>
            ) : null}
          </div>
        ) : (
          <p className="emptyState">No conversation is selected.</p>
        )}
      </section>
    </main>
  );
}

function QuestsWorkspace({ project, status }: { project: GameProject; status: string }) {
  return (
    <main className="workspace singleWorkspace">
      <aside className="panel">
        <section>
          <h2>Quests</h2>
          <p className="muted">TBD</p>
        </section>

        <section>
          <h2>Status</h2>
          <p className="status">{status}</p>
        </section>
      </aside>

      <section className="contentPanel">
        <div className="contentHeader">
          <h2>Quests</h2>
          <p>{project.files.quests.length} quests</p>
        </div>
        <p className="emptyState">Quest editing is not implemented yet.</p>
      </section>
    </main>
  );
}

type WorldSummary = {
  path: string;
  name: string;
};

type WorldValidationEntry = {
  path: string;
  result: WorldValidationResult;
};

type ConversationSummary = {
  path: string;
  conversationId: string;
  startNodeId: string;
  name: string;
};

type ConversationValidationEntry = {
  path: string;
  result: ConversationValidationResult;
};

function summarizeWorlds(project: GameProject, worlds: WorldDocuments): WorldSummary[] {
  return project.files.maps.flatMap((path) => {
    const world = worlds[path];
    if (!world) {
      return [];
    }

    return {
      path,
      name: world.displayName || world.id || titleFromStorageName(filenameFromProjectPath(path)),
    };
  });
}

function validateWorldDocuments(worlds: WorldDocuments): WorldValidationEntry[] {
  return Object.entries(worlds).map(([path, world]) => ({
    path,
    result: validateWorld(world),
  }));
}

function summarizeConversations(
  project: GameProject,
  conversations: ConversationDocuments
): ConversationSummary[] {
  return project.files.conversations.flatMap((path) => {
    const document = conversations[path];
    if (!document) {
      return [];
    }

    return document.conversations.map((conversation) => ({
      path,
      conversationId: conversation.id,
      startNodeId: conversation.startNodeId,
      name: document.displayName || conversation.id || titleFromStorageName(filenameFromProjectPath(path)),
    }));
  });
}

function validateConversationDocuments(conversations: ConversationDocuments): ConversationValidationEntry[] {
  return Object.entries(conversations).map(([path, document]) => ({
    path,
    result: validateConversationDocument(document),
  }));
}

function titleFromStorageName(filename: string): string {
  return filename
    .replace(/\.json$/i, "")
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}

function TileDetails({ world, x, y }: { world: WorldFormat; x: number; y: number }) {
  const index = tileIndex(world.size, x, y);
  return (
    <div className="details">
      <p className="metric">x {x}</p>
      <p className="metric">y {y}</p>
      <p className="metric">terrain {world.terrain[index]}</p>
      <p className="metric">blocked {world.blockers?.[index] ? "yes" : "no"}</p>
    </div>
  );
}

function createWall(world: WorldFormat, rawType: string, x: number, y: number): WorldWall {
  const type = sanitizeToken(rawType || "wall");
  let nextNumber = 1;
  let id = `wall_${String(nextNumber).padStart(3, "0")}`;

  while (world.walls.some((wall) => wall.id === id)) {
    nextNumber += 1;
    id = `wall_${String(nextNumber).padStart(3, "0")}`;
  }

  return {
    id,
    type,
    x,
    y,
  };
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

function coversTile(item: WorldObject | WorldWall, x: number, y: number): boolean {
  const width = "width" in item ? item.width ?? 1 : 1;
  const height = "height" in item ? item.height ?? 1 : 1;
  return x >= item.x && y >= item.y && x < item.x + width && y < item.y + height;
}

function isTileSelected(
  selection: Selection,
  x: number,
  y: number,
  wall: WorldWall | undefined,
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
  if (selection?.type === "wall") {
    return wall?.id === selection.id;
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
    return "action canceled";
  }
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

export default App;
