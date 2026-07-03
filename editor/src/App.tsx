import { useMemo, useState } from "react";
import { MapViewport3D } from "./MapViewport3D.tsx";
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
  createBlankQuestDocument,
  createBlankQuestRewardItem,
  createBlankQuestStep,
  sanitizeQuestId,
  validateQuestDocument,
  type Quest,
  type QuestDocument,
  type QuestRewardItem,
  type QuestStep,
  type ValidationResult as QuestValidationResult,
} from "./questFormat.ts";
import {
  conversationPathFromFilename,
  createBlankGameProject,
  DEFAULT_WORLD_PATH,
  filenameFromProjectPath,
  isValidProjectPath,
  isValidProjectFilename,
  mapPathFromFilename,
  questPathFromFilename,
  validateGameProject,
  type GameProject,
} from "./gameProject.ts";
import {
  createBlankWorld,
  entityPosition,
  entitySize,
  resizeWorld,
  tileIndex,
  validateWorld,
  type WorldEntity,
  type WorldFormat,
  type WorldWall,
  type ValidationResult as WorldValidationResult,
} from "./worldFormat.ts";
import {
  openGameProjectFolder,
  saveGameProjectFolder,
  saveGameProjectFolderAs,
  supportsProjectFolders,
  type ConversationDocuments,
  type ProjectDirectoryHandle,
  type QuestDocuments,
  type WorldDocuments,
} from "./fileSystem.ts";
import { TERRAIN_SWATCHES, terrainColor } from "./terrainPalette.ts";

type Tool = "terrain" | "blocker" | "height" | "wall" | "entity" | "select";
type HeightBrushMode = "set" | "raise" | "lower";
type EditorTab = "map" | "conversations" | "quests";
type Selection =
  | { type: "wall"; id: string }
  | { type: "entity"; id: string }
  | { type: "tile"; x: number; y: number }
  | null;

const WALL_TYPES = ["stone", "wood"];
const ENTITY_TYPES = ["tree", "door", "building", "chest", "rock", "human"];

function App() {
  const defaultWorld = createBlankWorld();
  const defaultWorlds = { [DEFAULT_WORLD_PATH]: defaultWorld };
  const [project, setProject] = useState<GameProject>(() => createBlankGameProject());
  const [world, setWorld] = useState<WorldFormat>(() => defaultWorld);
  const [worldDocuments, setWorldDocuments] = useState<WorldDocuments>(() => defaultWorlds);
  const [deletedWorldPaths, setDeletedWorldPaths] = useState<string[]>([]);
  const [conversationDocuments, setConversationDocuments] = useState<ConversationDocuments>({});
  const [deletedConversationPaths, setDeletedConversationPaths] = useState<string[]>([]);
  const [questDocuments, setQuestDocuments] = useState<QuestDocuments>({});
  const [deletedQuestPaths, setDeletedQuestPaths] = useState<string[]>([]);
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
  const [heightBrushMode, setHeightBrushMode] = useState<HeightBrushMode>("set");
  const [heightBrushValue, setHeightBrushValue] = useState(1);
  const [wallType, setWallType] = useState("stone");
  const [entityType, setEntityType] = useState("tree");
  const [entityBlocksMovement, setEntityBlocksMovement] = useState(true);
  const [componentText, setComponentText] = useState("{}");
  const [mapNameInput, setMapNameInput] = useState("New Map");
  const [conversationNameInput, setConversationNameInput] = useState("New Conversation");
  const [questNameInput, setQuestNameInput] = useState("New Quest");
  const [selectedConversationPath, setSelectedConversationPath] = useState<string | null>(null);
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [selectedQuestPath, setSelectedQuestPath] = useState<string | null>(null);
  const [selectedQuestId, setSelectedQuestId] = useState<string | null>(null);

  const projectValidation = useMemo(() => validateGameProject(project), [project]);
  const worldValidations = useMemo(() => validateWorldDocuments(worldDocuments), [worldDocuments]);
  const conversationValidations = useMemo(() => validateConversationDocuments(conversationDocuments), [conversationDocuments]);
  const questValidations = useMemo(() => validateQuestDocuments(questDocuments), [questDocuments]);
  const worldPathValid = isValidProjectPath(worldPath);
  const worldsValid = worldValidations.every((validation) => validation.result.valid);
  const conversationsValid = conversationValidations.every((validation) => validation.result.valid);
  const questsValid = questValidations.every((validation) => validation.result.valid);
  const canSave = projectValidation.valid && worldsValid && conversationsValid && questsValid && worldPathValid;
  const mapSummaries = useMemo(() => summarizeWorlds(project, worldDocuments), [project, worldDocuments]);
  const currentMapName = world.displayName || world.id || "Untitled Map";
  const conversationSummaries = useMemo(() => summarizeConversations(project, conversationDocuments), [project, conversationDocuments]);
  const questSummaries = useMemo(() => summarizeQuests(project, questDocuments), [project, questDocuments]);
  const selectedConversationDocument = selectedConversationPath ? conversationDocuments[selectedConversationPath] : undefined;
  const selectedConversation = selectedConversationDocument?.conversations.find(
    (conversation) => conversation.id === selectedConversationId
  );
  const selectedNode = selectedConversation?.nodes.find((node) => node.id === selectedNodeId);
  const selectedQuestDocument = selectedQuestPath ? questDocuments[selectedQuestPath] : undefined;
  const selectedQuest = selectedQuestDocument?.quests.find((quest) => quest.id === selectedQuestId);
  const selectedWall =
    selection?.type === "wall" ? world.walls.find((wall) => wall.id === selection.id) : undefined;
  const selectedEntity =
    selection?.type === "entity" ? world.entities.find((entity) => entity.id === selection.id) : undefined;

  async function handleNew() {
    const nextWorld = createBlankWorld();
    setProject(createBlankGameProject());
    setWorld(nextWorld);
    setWorldDocuments({ [DEFAULT_WORLD_PATH]: nextWorld });
    setDeletedWorldPaths([]);
    setConversationDocuments({});
    setDeletedConversationPaths([]);
    setQuestDocuments({});
    setDeletedQuestPaths([]);
    setProjectHandle(null);
    setProjectFolderName("unsaved project");
    setWorldPath(DEFAULT_WORLD_PATH);
    setSelection(null);
    setSelectedConversationPath(null);
    setSelectedConversationId(null);
    setSelectedNodeId(null);
    setSelectedQuestPath(null);
    setSelectedQuestId(null);
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
      setQuestDocuments(opened.quests);
      setDeletedQuestPaths([]);
      setProjectHandle(opened.handle);
      setProjectFolderName(opened.handle.name);
      setWorldPath(opened.worldPath);
      setSelection(null);
      const firstConversation = summarizeConversations(opened.project, opened.conversations)[0];
      setSelectedConversationPath(firstConversation?.path ?? null);
      setSelectedConversationId(firstConversation?.conversationId ?? null);
      setSelectedNodeId(firstConversation?.startNodeId ?? null);
      const firstQuest = summarizeQuests(opened.project, opened.quests)[0];
      setSelectedQuestPath(firstQuest?.path ?? null);
      setSelectedQuestId(firstQuest?.questId ?? null);
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
          questDocuments,
          deletedWorldPaths,
          deletedConversationPaths,
          deletedQuestPaths
        );
        setProject(savedProject);
        setDeletedWorldPaths([]);
        setDeletedConversationPaths([]);
        setDeletedQuestPaths([]);
      } else {
        const saved = await saveGameProjectFolderAs(project, worldDocuments, worldPath, conversationDocuments, questDocuments);
        setProject(saved.project);
        setProjectHandle(saved.handle);
        setProjectFolderName(saved.handle.name);
        setDeletedWorldPaths([]);
        setDeletedConversationPaths([]);
        setDeletedQuestPaths([]);
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
      const saved = await saveGameProjectFolderAs(project, worldDocuments, worldPath, conversationDocuments, questDocuments);
      setProject(saved.project);
      setProjectHandle(saved.handle);
      setProjectFolderName(saved.handle.name);
      setDeletedWorldPaths([]);
      setDeletedConversationPaths([]);
      setDeletedQuestPaths([]);
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
      messages: [...node.messages, { text: "" }],
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

  function createQuest() {
    const name = questNameInput.trim();
    const id = sanitizeQuestId(name);
    const filename = `${id}.json`;
    if (!isValidProjectFilename(filename)) {
      setStatus("quest name is invalid");
      return;
    }
    const path = questPathFromFilename(filename);
    if (project.files.quests.includes(path)) {
      setStatus("quest already exists");
      return;
    }

    setProject((current) => ({
      ...current,
      files: {
        ...current.files,
        quests: [...current.files.quests, path],
      },
    }));
    setQuestDocuments((current) => ({
      ...current,
      [path]: {
        ...createBlankQuestDocument(id),
        displayName: name,
      },
    }));
    setDeletedQuestPaths((current) => current.filter((candidate) => candidate !== path));
    setSelectedQuestPath(path);
    setSelectedQuestId(id);
    setDirty(true);
    setStatus("quest created");
  }

  function deleteQuest(path: string, questId: string) {
    const document = questDocuments[path];
    const removesWholeDocument = !document || document.quests.length <= 1;

    setProject((current) => ({
      ...current,
      files: {
        ...current.files,
        quests: removesWholeDocument
          ? current.files.quests.filter((candidate) => candidate !== path)
          : current.files.quests,
      },
    }));
    setQuestDocuments((current) => {
      if (removesWholeDocument) {
        const next = { ...current };
        delete next[path];
        return next;
      }

      return {
        ...current,
        [path]: {
          ...document,
          quests: document.quests.filter((quest) => quest.id !== questId),
        },
      };
    });
    if (removesWholeDocument) {
      setDeletedQuestPaths((current) => [...new Set([...current, path])]);
    }
    if (selectedQuestPath === path && selectedQuestId === questId) {
      const nextQuest = questSummaries.find((summary) => summary.path !== path || summary.questId !== questId);
      setSelectedQuestPath(nextQuest?.path ?? null);
      setSelectedQuestId(nextQuest?.questId ?? null);
    }
    setDirty(true);
    setStatus("quest deleted");
  }

  function selectQuest(path: string, questId: string) {
    setSelectedQuestPath(path);
    setSelectedQuestId(questId);
  }

  function updateSelectedQuest(updater: (quest: Quest) => Quest) {
    if (!selectedQuestPath || !selectedQuestId) {
      return;
    }

    setQuestDocuments((current) => {
      const document = current[selectedQuestPath];
      if (!document) {
        return current;
      }

      let nextSelectedId = selectedQuestId;
      const quests = document.quests.map((quest) => {
        if (quest.id !== selectedQuestId) {
          return quest;
        }
        const nextQuest = updater(quest);
        nextSelectedId = nextQuest.id;
        return nextQuest;
      });
      setSelectedQuestId(nextSelectedId);
      return {
        ...current,
        [selectedQuestPath]: {
          ...document,
          quests,
        },
      };
    });
    setDirty(true);
  }

  function updateSelectedQuestDocument(patch: Partial<QuestDocument>) {
    if (!selectedQuestPath) {
      return;
    }

    setQuestDocuments((current) => {
      const document = current[selectedQuestPath];
      if (!document) {
        return current;
      }
      return {
        ...current,
        [selectedQuestPath]: {
          ...document,
          ...patch,
        },
      };
    });
    setDirty(true);
  }

  function addQuestStep() {
    updateSelectedQuest((quest) => ({
      ...quest,
      steps: [...quest.steps, createBlankQuestStep(quest.steps, "step")],
    }));
  }

  function updateQuestStep(index: number, patch: Partial<QuestStep>) {
    updateSelectedQuest((quest) => ({
      ...quest,
      steps: quest.steps.map((step, stepIndex) =>
        stepIndex === index ? { ...step, ...patch } : step
      ),
    }));
  }

  function updateQuestStepRequirement(index: number, patch: Partial<QuestStep["requirement"]>) {
    updateSelectedQuest((quest) => ({
      ...quest,
      steps: quest.steps.map((step, stepIndex) =>
        stepIndex === index ? { ...step, requirement: { ...step.requirement, ...patch } } : step
      ),
    }));
  }

  function addQuestRewardItem() {
    updateSelectedQuest((quest) => ({
      ...quest,
      rewards: {
        items: [...quest.rewards.items, createBlankQuestRewardItem()],
      },
    }));
  }

  function updateQuestRewardItem(index: number, patch: Partial<QuestRewardItem>) {
    updateSelectedQuest((quest) => ({
      ...quest,
      rewards: {
        items: quest.rewards.items.map((item, itemIndex) =>
          itemIndex === index ? { ...item, ...patch } : item
        ),
      },
    }));
  }

  function deleteQuestRewardItem(index: number) {
    updateSelectedQuest((quest) => {
      if (quest.rewards.items.length <= 1) {
        return quest;
      }
      return {
        ...quest,
        rewards: {
          items: quest.rewards.items.filter((_, itemIndex) => itemIndex !== index),
        },
      };
    });
  }

  function deleteQuestStep(index: number) {
    updateSelectedQuest((quest) => {
      if (quest.steps.length <= 1) {
        return quest;
      }
      return {
        ...quest,
        steps: quest.steps.filter((_, stepIndex) => stepIndex !== index),
      };
    });
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

  function clampHeight(value: number) {
    if (!Number.isFinite(value)) {
      return 0;
    }
    return Math.min(10, Math.max(0, Math.floor(value)));
  }

  function setTileHeight(x: number, y: number, height: number) {
    updateWorld((current) => {
      if (x < 0 || y < 0 || x >= current.size.x || y >= current.size.y) {
        return current;
      }

      const tileCount = current.size.x * current.size.y;
      const heights = current.heights.length === tileCount
        ? [...current.heights]
        : new Array(tileCount).fill(0);
      heights[tileIndex(current.size, x, y)] = clampHeight(height);
      return { ...current, heights };
    });
    setSelection({ type: "tile", x, y });
  }

  function applyHeightBrush(x: number, y: number) {
    if (heightBrushMode === "set") {
      setTileHeight(x, y, heightBrushValue);
      return;
    }

    updateWorld((current) => {
      if (x < 0 || y < 0 || x >= current.size.x || y >= current.size.y) {
        return current;
      }

      const tileCount = current.size.x * current.size.y;
      const heights = current.heights.length === tileCount
        ? [...current.heights]
        : new Array(tileCount).fill(0);
      const index = tileIndex(current.size, x, y);
      const currentHeight = clampHeight(heights[index] ?? 0);
      const nextHeight = currentHeight + (heightBrushMode === "raise" ? 1 : -1);
      heights[index] = clampHeight(nextHeight);
      return { ...current, heights };
    });
    setSelection({ type: "tile", x, y });
  }

  function handleTilePointerDown(x: number, y: number) {
    if (tool === "height") {
      applyHeightBrush(x, y);
      return;
    }

    handleTileClick(x, y);
  }

  function handleTilePointerEnter(x: number, y: number, buttons: number) {
    if (tool === "height" && (buttons & 1) === 1) {
      applyHeightBrush(x, y);
    }
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

    if (tool === "entity") {
      const entity = createEntity(world, entityType, x, y, entityBlocksMovement);
      updateWorld((current) => ({ ...current, entities: [...current.entities, entity] }));
      setSelection({ type: "entity", id: entity.id });
      setComponentText(JSON.stringify(entity.components ?? {}, null, 2));
      return;
    }

    if (tool === "wall") {
      const wall = createWall(world, wallType, x, y);
      updateWorld((current) => ({ ...current, walls: [...current.walls, wall] }));
      setSelection({ type: "wall", id: wall.id });
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

    const entity = [...world.entities]
      .reverse()
      .find((candidate) => coversTile(candidate, x, y));
    if (entity) {
      setSelection({ type: "entity", id: entity.id });
      setComponentText(JSON.stringify(entity.components ?? {}, null, 2));
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

  function updateSelectedEntity(patch: Partial<WorldEntity>) {
    if (!selectedEntity) {
      return;
    }
    updateWorld((current) => ({
      ...current,
      entities: current.entities.map((entity) =>
        entity.id === selectedEntity.id ? { ...entity, ...patch } : entity
      ),
    }));
    if (patch.id) {
      setSelection({ type: "entity", id: patch.id });
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

  function deleteSelection() {
    if (selection?.type === "wall") {
      updateWorld((current) => ({
        ...current,
        walls: current.walls.filter((wall) => wall.id !== selection.id),
      }));
    } else if (selection?.type === "entity") {
      updateWorld((current) => ({
        ...current,
        entities: current.entities.filter((entity) => entity.id !== selection.id),
      }));
    }
    setSelection(null);
  }

  function applyEntityComponents() {
    if (!selectedEntity) {
      return;
    }
    try {
      const parsed = JSON.parse(componentText) as unknown;
      if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
        setStatus("entity components must be a JSON object");
        return;
      }
      updateSelectedEntity({ components: parsed as Record<string, unknown> });
      setStatus("entity components updated");
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
              {(["terrain", "blocker", "height", "wall", "entity", "select"] as Tool[]).map((item) => (
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

            {tool === "height" ? (
              <div className="toolSettings">
                <label>
                  Brush mode
                  <select
                    value={heightBrushMode}
                    onChange={(event) => setHeightBrushMode(event.target.value as HeightBrushMode)}
                  >
                    <option value="set">set</option>
                    <option value="raise">raise</option>
                    <option value="lower">lower</option>
                  </select>
                </label>
                {heightBrushMode === "set" ? (
                  <label>
                    Height
                    <input
                      type="number"
                      min="0"
                      max="10"
                      value={heightBrushValue}
                      onChange={(event) => setHeightBrushValue(clampHeight(Number(event.target.value)))}
                    />
                  </label>
                ) : (
                  <p className="metric">step 1</p>
                )}
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

            {tool === "entity" ? (
              <div className="toolSettings">
                <label>
                  Entity type
                  <input value={entityType} onChange={(event) => setEntityType(event.target.value)} />
                </label>
                <div className="quickPicks">
                  {ENTITY_TYPES.map((type) => (
                    <button key={type} type="button" onClick={() => setEntityType(type)}>
                      {type}
                    </button>
                  ))}
                </div>
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={entityBlocksMovement}
                    onChange={(event) => setEntityBlocksMovement(event.target.checked)}
                  />
                  blocks movement
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
          <MapViewport3D
            world={world}
            selection={selection}
            tool={tool}
            onTilePointerDown={handleTilePointerDown}
            onTilePointerEnter={handleTilePointerEnter}
          />
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

            {selectedEntity ? (
              <div className="details">
                <label>
                  Id
                  <input value={selectedEntity.id} onChange={(event) => updateSelectedEntity({ id: event.target.value })} />
                </label>
                <label>
                  Components JSON
                  <textarea value={componentText} onChange={(event) => setComponentText(event.target.value)} />
                </label>
                <div className="buttonRow">
                  <button type="button" onClick={applyEntityComponents}>Apply Components</button>
                  <button type="button" className="danger" onClick={deleteSelection}>Delete</button>
                </div>
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
        <QuestsWorkspace
          project={project}
          quests={questDocuments}
          summaries={questSummaries}
          selectedQuestPath={selectedQuestPath}
          selectedQuest={selectedQuest}
          nameInput={questNameInput}
          validations={questValidations}
          status={status}
          onNameInputChange={setQuestNameInput}
          onCreateQuest={createQuest}
          onDeleteQuest={deleteQuest}
          onSelectQuest={selectQuest}
          onUpdateQuestDocument={updateSelectedQuestDocument}
          onUpdateQuest={updateSelectedQuest}
          onAddStep={addQuestStep}
          onUpdateStep={updateQuestStep}
          onUpdateStepRequirement={updateQuestStepRequirement}
          onAddRewardItem={addQuestRewardItem}
          onUpdateRewardItem={updateQuestRewardItem}
          onDeleteRewardItem={deleteQuestRewardItem}
          onDeleteStep={deleteQuestStep}
        />
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

function QuestsWorkspace({
  project,
  quests,
  summaries,
  selectedQuestPath,
  selectedQuest,
  nameInput,
  validations,
  status,
  onNameInputChange,
  onCreateQuest,
  onDeleteQuest,
  onSelectQuest,
  onUpdateQuestDocument,
  onUpdateQuest,
  onAddStep,
  onUpdateStep,
  onUpdateStepRequirement,
  onAddRewardItem,
  onUpdateRewardItem,
  onDeleteRewardItem,
  onDeleteStep,
}: {
  project: GameProject;
  quests: QuestDocuments;
  summaries: QuestSummary[];
  selectedQuestPath: string | null;
  selectedQuest: Quest | undefined;
  nameInput: string;
  validations: QuestValidationEntry[];
  status: string;
  onNameInputChange: (value: string) => void;
  onCreateQuest: () => void;
  onDeleteQuest: (path: string, questId: string) => void;
  onSelectQuest: (path: string, questId: string) => void;
  onUpdateQuestDocument: (patch: Partial<QuestDocument>) => void;
  onUpdateQuest: (updater: (quest: Quest) => Quest) => void;
  onAddStep: () => void;
  onUpdateStep: (index: number, patch: Partial<QuestStep>) => void;
  onUpdateStepRequirement: (index: number, patch: Partial<QuestStep["requirement"]>) => void;
  onAddRewardItem: () => void;
  onUpdateRewardItem: (index: number, patch: Partial<QuestRewardItem>) => void;
  onDeleteRewardItem: (index: number) => void;
  onDeleteStep: (index: number) => void;
}) {
  const selectedDocument = selectedQuestPath ? quests[selectedQuestPath] : undefined;

  return (
    <main className="workspace singleWorkspace">
      <aside className="panel">
        <section>
          <h2>Quests</h2>
          <label>
            Name
            <input value={nameInput} onChange={(event) => onNameInputChange(event.target.value)} />
          </label>
          <button type="button" onClick={onCreateQuest}>Create Quest</button>
        </section>

        <section>
          <h2>List</h2>
          {summaries.length > 0 ? (
            <div className="stack">
              {summaries.map((summary) => (
                <button
                  key={`${summary.path}:${summary.questId}`}
                  type="button"
                  className={
                    selectedQuestPath === summary.path && selectedQuest?.id === summary.questId
                      ? "listButton active"
                      : "listButton"
                  }
                  onClick={() => onSelectQuest(summary.path, summary.questId)}
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
          <h2>Quests</h2>
          <p>{project.files.quests.length} quests</p>
        </div>
        {selectedQuest && selectedDocument && selectedQuestPath ? (
          <div className="conversationEditor">
            <section className="editorSection">
              <div className="sectionHeader">
                <h3>Details</h3>
                <button
                  type="button"
                  className="danger"
                  onClick={() => onDeleteQuest(selectedQuestPath, selectedQuest.id)}
                >
                  Delete
                </button>
              </div>
              <div className="fieldRow">
                <label>
                  Name
                  <input
                    value={selectedQuest.displayName ?? selectedDocument.displayName ?? ""}
                    onChange={(event) =>
                      onUpdateQuest((quest) => ({ ...quest, displayName: event.target.value || undefined }))
                    }
                  />
                </label>
                <label>
                  Id
                  <input
                    value={selectedQuest.id}
                    onChange={(event) => onUpdateQuest((quest) => ({ ...quest, id: event.target.value }))}
                  />
                </label>
              </div>
              <label>
                Document name
                <input
                  value={selectedDocument.displayName ?? ""}
                  onChange={(event) => onUpdateQuestDocument({ displayName: event.target.value })}
                />
              </label>
              <label>
                Start event id
                <input
                  value={selectedQuest.startEventId ?? ""}
                  onChange={(event) =>
                    onUpdateQuest((quest) => ({ ...quest, startEventId: event.target.value || undefined }))
                  }
                />
              </label>
              <label>
                Description
                <textarea
                  value={selectedQuest.description ?? ""}
                  onChange={(event) =>
                    onUpdateQuest((quest) => ({ ...quest, description: event.target.value || undefined }))
                  }
                />
              </label>
            </section>

            <section className="editorSection">
              <div className="sectionHeader">
                <h3>Rewards</h3>
                <button type="button" onClick={onAddRewardItem}>Add Reward</button>
              </div>
              <div className="stack">
                {selectedQuest.rewards.items.map((reward, index) => (
                  <div key={index} className="subEditor">
                    <div className="sectionHeader compact">
                      <h3>{reward.name || "Reward"}</h3>
                      <button
                        type="button"
                        className="danger"
                        disabled={selectedQuest.rewards.items.length <= 1}
                        onClick={() => onDeleteRewardItem(index)}
                      >
                        Delete Reward
                      </button>
                    </div>
                    <div className="fieldRow">
                      <label>
                        Name
                        <input
                          value={reward.name}
                          onChange={(event) => onUpdateRewardItem(index, { name: event.target.value })}
                        />
                      </label>
                      <label>
                        Type
                        <input
                          value={reward.type}
                          onChange={(event) => onUpdateRewardItem(index, { type: event.target.value })}
                        />
                      </label>
                      <label>
                        Count
                        <input
                          type="number"
                          min="1"
                          value={reward.count}
                          onChange={(event) =>
                            onUpdateRewardItem(index, { count: Math.max(1, Number(event.target.value)) })
                          }
                        />
                      </label>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section className="editorSection">
              <div className="sectionHeader">
                <h3>Steps</h3>
                <button type="button" onClick={onAddStep}>Add Step</button>
              </div>
              <div className="stack">
                {selectedQuest.steps.map((step, index) => (
                  <div key={index} className="subEditor">
                    <div className="sectionHeader compact">
                      <h3>{index + 1}. {step.id}</h3>
                      <button
                        type="button"
                        className="danger"
                        disabled={selectedQuest.steps.length <= 1}
                        onClick={() => onDeleteStep(index)}
                      >
                        Delete Step
                      </button>
                    </div>
                    <div className="fieldRow">
                      <label>
                        Id
                        <input value={step.id} onChange={(event) => onUpdateStep(index, { id: event.target.value })} />
                      </label>
                      <label>
                        Count
                        <input
                          type="number"
                          min="1"
                          value={step.requirement.count}
                          onChange={(event) =>
                            onUpdateStepRequirement(index, { count: Math.max(1, Number(event.target.value)) })
                          }
                        />
                      </label>
                    </div>
                    <label>
                      Event id
                      <input
                        value={step.requirement.eventId}
                        onChange={(event) => onUpdateStepRequirement(index, { eventId: event.target.value })}
                      />
                    </label>
                    <label>
                      Description
                      <textarea
                        value={step.description}
                        onChange={(event) => onUpdateStep(index, { description: event.target.value })}
                      />
                    </label>
                  </div>
                ))}
              </div>
            </section>
          </div>
        ) : (
          <p className="emptyState">No quest is selected.</p>
        )}
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

type QuestSummary = {
  path: string;
  questId: string;
  name: string;
};

type QuestValidationEntry = {
  path: string;
  result: QuestValidationResult;
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

function summarizeQuests(project: GameProject, quests: QuestDocuments): QuestSummary[] {
  return project.files.quests.flatMap((path) => {
    const document = quests[path];
    if (!document) {
      return [];
    }

    return document.quests.map((quest) => ({
      path,
      questId: quest.id,
      name: quest.displayName || document.displayName || quest.id || titleFromStorageName(filenameFromProjectPath(path)),
    }));
  });
}

function validateQuestDocuments(quests: QuestDocuments): QuestValidationEntry[] {
  return Object.entries(quests).map(([path, document]) => ({
    path,
    result: validateQuestDocument(document),
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
      <p className="metric">height {world.heights[index] ?? 0}</p>
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

function createEntity(
  world: WorldFormat,
  rawType: string,
  x: number,
  y: number,
  blocksMovement: boolean
): WorldEntity {
  const type = sanitizeToken(rawType || "entity");
  let nextNumber = 1;
  let id = `${type}_${String(nextNumber).padStart(3, "0")}`;

  while (world.entities.some((entity) => entity.id === id)) {
    nextNumber += 1;
    id = `${type}_${String(nextNumber).padStart(3, "0")}`;
  }

  const components: Record<string, unknown> = {
    position: { x, y },
    metadata: {
      name: id,
      type,
      width: 1,
      height: 1,
      blocksMovement,
    },
    renderable: type === "door" ? { type, orientation: "north" } : { type },
  };

  if (type === "door") {
    components.openable = { isOpen: false };
  }

  return {
    id,
    components,
  };
}

function coversTile(item: WorldEntity | WorldWall, x: number, y: number): boolean {
  if ("components" in item) {
    const position = entityPosition(item);
    if (!position) {
      return false;
    }
    const size = entitySize(item);
    return x >= position.x && y >= position.y && x < position.x + size.width && y < position.y + size.height;
  }
  return x === item.x && y === item.y;
}

function sanitizeToken(value: string): string {
  const token = value.trim().toLowerCase().replace(/[^a-z0-9_-]+/g, "_").replace(/^_+|_+$/g, "");
  return token || "entity";
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
