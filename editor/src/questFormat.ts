import { ID_PATTERN, isObject, serializeJson, titleFromId, type ValidationResult } from "./formatUtils";

export type { ValidationResult } from "./formatUtils";

export type QuestDocument = {
  formatVersion: 1;
  id: string;
  displayName?: string;
  quests: Quest[];
};

export type Quest = {
  id: string;
  displayName?: string;
  description?: string;
  startEventId?: string;
  steps: QuestStep[];
};

export type QuestStep = {
  id: string;
  description: string;
  requirement: QuestRequirement;
};

export type QuestRequirement = {
  eventId: string;
  count: number;
};

export function createBlankQuestDocument(id: string): QuestDocument {
  const questId = sanitizeQuestId(id);
  return {
    formatVersion: 1,
    id: questId,
    displayName: titleFromId(questId),
    quests: [createBlankQuest(questId)],
  };
}

export function createBlankQuest(id: string): Quest {
  const questId = sanitizeQuestId(id);
  return {
    id: questId,
    displayName: titleFromId(questId),
    description: "",
    startEventId: "",
    steps: [createBlankQuestStep([], "start")],
  };
}

export function createBlankQuestStep(existingSteps: QuestStep[], preferredId = "step"): QuestStep {
  return {
    id: nextUniqueId(preferredId, existingSteps.map((step) => step.id)),
    description: "Do the next thing.",
    requirement: {
      eventId: "conversation:node:conversation:start",
      count: 1,
    },
  };
}

export function normalizeQuestDocument(value: unknown): QuestDocument {
  if (!isObject(value)) {
    throw new Error("quest data must contain a JSON object");
  }

  const id = typeof value.id === "string" ? value.id : "untitled_quests";
  const document: QuestDocument = {
    formatVersion: 1,
    id,
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    quests: Array.isArray(value.quests) ? value.quests.map(normalizeQuest) : [],
  };

  const validation = validateQuestDocument(document);
  if (!validation.valid) {
    throw new Error(validation.errors.join("\n"));
  }
  return document;
}

export function validateQuestDocument(document: QuestDocument): ValidationResult {
  const errors: string[] = [];

  if (document.formatVersion !== 1) {
    errors.push(`quest document "${document.id}" formatVersion must be 1`);
  }
  if (!ID_PATTERN.test(document.id)) {
    errors.push(`quest document id "${document.id}" is invalid`);
  }
  if (document.quests.length === 0) {
    errors.push(`quest document "${document.id}" must contain at least one quest`);
  }

  const questIds = new Set<string>();
  for (const quest of document.quests) {
    if (!ID_PATTERN.test(quest.id)) {
      errors.push(`quest id "${quest.id}" is invalid`);
    }
    if (questIds.has(quest.id)) {
      errors.push(`quest id "${quest.id}" is duplicated`);
    }
    questIds.add(quest.id);
    if (quest.steps.length === 0) {
      errors.push(`quest "${quest.id}" must contain at least one step`);
    }

    const stepIds = new Set<string>();
    for (const step of quest.steps) {
      if (!ID_PATTERN.test(step.id)) {
        errors.push(`quest "${quest.id}" step id "${step.id}" is invalid`);
      }
      if (stepIds.has(step.id)) {
        errors.push(`quest "${quest.id}" step id "${step.id}" is duplicated`);
      }
      stepIds.add(step.id);
      if (step.description.trim().length === 0) {
        errors.push(`quest "${quest.id}" step "${step.id}" must include a description`);
      }
      if (step.requirement.eventId.trim().length === 0) {
        errors.push(`quest "${quest.id}" step "${step.id}" must include an event id`);
      }
      if (!Number.isFinite(step.requirement.count) || step.requirement.count < 1) {
        errors.push(`quest "${quest.id}" step "${step.id}" count must be at least 1`);
      }
    }
  }

  return { valid: errors.length === 0, errors };
}

export function serializeQuestDocument(document: QuestDocument): string {
  return serializeJson(document);
}

export function sanitizeQuestId(value: string): string {
  const token = value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "_")
    .replace(/^_+|_+$/g, "");
  return token || "quest";
}

function normalizeQuest(value: unknown): Quest {
  if (!isObject(value)) {
    return createBlankQuest("invalid_quest");
  }
  return {
    id: typeof value.id === "string" ? value.id : "invalid_quest",
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    description: typeof value.description === "string" ? value.description : undefined,
    startEventId: typeof value.startEventId === "string" ? value.startEventId : undefined,
    steps: Array.isArray(value.steps) ? value.steps.map(normalizeQuestStep) : [],
  };
}

function normalizeQuestStep(value: unknown): QuestStep {
  if (!isObject(value)) {
    return createBlankQuestStep([], "invalid_step");
  }
  const requirement = isObject(value.requirement) ? value.requirement : {};
  return {
    id: typeof value.id === "string" ? value.id : "invalid_step",
    description: typeof value.description === "string" ? value.description : "",
    requirement: {
      eventId: typeof requirement.eventId === "string" ? requirement.eventId : "",
      count: typeof requirement.count === "number" ? requirement.count : 1,
    },
  };
}

function nextUniqueId(base: string, existing: string[]): string {
  const safeBase = sanitizeQuestId(base);
  const existingIds = new Set(existing);
  if (!existingIds.has(safeBase)) {
    return safeBase;
  }
  let index = 2;
  while (existingIds.has(`${safeBase}_${index}`)) {
    index += 1;
  }
  return `${safeBase}_${index}`;
}
