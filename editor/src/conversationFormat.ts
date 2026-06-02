import { ID_PATTERN, isObject, serializeJson, titleFromId, type ValidationResult } from "./formatUtils";

export type { ValidationResult } from "./formatUtils";

export type ConversationDocument = {
  formatVersion: 1;
  id: string;
  displayName?: string;
  conversations: Conversation[];
};

export type Conversation = {
  id: string;
  startNodeId: string;
  nodes: ConversationNode[];
};

export type ConversationNode = {
  id: string;
  messages: ConversationMessage[];
  options?: ConversationOption[];
  endConversation?: boolean;
  tags?: string[];
};

export type ConversationMessage = {
  text: string;
  portrait?: string;
};

export type ConversationOption = {
  id: string;
  text: string;
  nextNodeId: string;
};

export function createBlankConversationDocument(id: string): ConversationDocument {
  const conversationId = sanitizeConversationId(id);
  return {
    formatVersion: 1,
    id: conversationId,
    displayName: titleFromId(conversationId),
    conversations: [createBlankConversation(conversationId)],
  };
}

export function createBlankConversation(id: string): Conversation {
  const conversationId = sanitizeConversationId(id);
  return {
    id: conversationId,
    startNodeId: "start",
    nodes: [
      {
        id: "start",
        messages: [
          {
            text: "Hello.",
          },
        ],
        endConversation: true,
      },
    ],
  };
}

export function createBlankConversationNode(existingNodes: ConversationNode[]): ConversationNode {
  const id = nextUniqueId("node", existingNodes.map((node) => node.id));
  return {
    id,
    messages: [
      {
        text: "",
      },
    ],
    endConversation: true,
  };
}

export function createBlankConversationOption(existingOptions: ConversationOption[], nextNodeId: string): ConversationOption {
  return {
    id: nextUniqueId("option", existingOptions.map((option) => option.id)),
    text: "",
    nextNodeId,
  };
}

export function normalizeConversationDocument(value: unknown): ConversationDocument {
  if (!isObject(value)) {
    throw new Error("conversation data must contain a JSON object");
  }

  const id = typeof value.id === "string" ? value.id : "untitled_conversations";
  const document: ConversationDocument = {
    formatVersion: 1,
    id,
    displayName: typeof value.displayName === "string" ? value.displayName : undefined,
    conversations: Array.isArray(value.conversations)
      ? value.conversations.map(normalizeConversation)
      : [],
  };

  const validation = validateConversationDocument(document);
  if (!validation.valid) {
    throw new Error(validation.errors.join("\n"));
  }

  return document;
}

export function validateConversationDocument(document: ConversationDocument): ValidationResult {
  const errors: string[] = [];

  if (document.formatVersion !== 1) {
    errors.push(`conversation document "${document.id}" formatVersion must be 1`);
  }
  if (!ID_PATTERN.test(document.id)) {
    errors.push(`conversation document id "${document.id}" is invalid`);
  }
  if (document.conversations.length === 0) {
    errors.push(`conversation document "${document.id}" must contain at least one conversation`);
  }

  const conversationIds = new Set<string>();
  for (const conversation of document.conversations) {
    if (!ID_PATTERN.test(conversation.id)) {
      errors.push(`conversation id "${conversation.id}" is invalid`);
    }
    if (conversationIds.has(conversation.id)) {
      errors.push(`conversation id "${conversation.id}" is duplicated`);
    }
    conversationIds.add(conversation.id);

    const nodeIds = new Set(conversation.nodes.map((node) => node.id));
    if (!nodeIds.has(conversation.startNodeId)) {
      errors.push(`conversation "${conversation.id}" start node "${conversation.startNodeId}" does not exist`);
    }
    if (conversation.nodes.length === 0) {
      errors.push(`conversation "${conversation.id}" must contain at least one node`);
    }

    const seenNodeIds = new Set<string>();
    for (const node of conversation.nodes) {
      if (!ID_PATTERN.test(node.id)) {
        errors.push(`conversation "${conversation.id}" node id "${node.id}" is invalid`);
      }
      if (seenNodeIds.has(node.id)) {
        errors.push(`conversation "${conversation.id}" node id "${node.id}" is duplicated`);
      }
      seenNodeIds.add(node.id);
      if (node.messages.length === 0) {
        errors.push(`conversation "${conversation.id}" node "${node.id}" must contain at least one message`);
      }
      for (const message of node.messages) {
        if (message.text.trim().length === 0) {
          errors.push(`conversation "${conversation.id}" node "${node.id}" has an empty message`);
        }
      }
      if (node.endConversation && node.options && node.options.length > 0) {
        errors.push(`conversation "${conversation.id}" node "${node.id}" cannot both end and have options`);
      }
      if (!node.endConversation && (!node.options || node.options.length === 0)) {
        errors.push(`conversation "${conversation.id}" node "${node.id}" must end or have options`);
      }
      for (const option of node.options ?? []) {
        if (!ID_PATTERN.test(option.id)) {
          errors.push(`conversation "${conversation.id}" node "${node.id}" option id "${option.id}" is invalid`);
        }
        if (option.text.trim().length === 0) {
          errors.push(`conversation "${conversation.id}" node "${node.id}" option "${option.id}" has empty text`);
        }
        if (!nodeIds.has(option.nextNodeId)) {
          errors.push(`conversation "${conversation.id}" option "${option.id}" targets missing node "${option.nextNodeId}"`);
        }
      }
    }
  }

  return { valid: errors.length === 0, errors };
}

export function serializeConversationDocument(document: ConversationDocument): string {
  return serializeJson(document);
}

export function sanitizeConversationId(value: string): string {
  const token = value.trim().toLowerCase().replace(/\.json$/i, "").replace(/[^a-z0-9_-]+/g, "_").replace(/^_+|_+$/g, "");
  return token || "conversation";
}

function normalizeConversation(value: unknown): Conversation {
  if (!isObject(value)) {
    return createBlankConversation("invalid_conversation");
  }

  const id = typeof value.id === "string" ? value.id : "invalid_conversation";
  const nodes = Array.isArray(value.nodes) ? value.nodes.map(normalizeConversationNode) : [];
  return {
    id,
    startNodeId: typeof value.startNodeId === "string" ? value.startNodeId : nodes[0]?.id ?? "start",
    nodes,
  };
}

function normalizeConversationNode(value: unknown): ConversationNode {
  if (!isObject(value)) {
    return createBlankConversationNode([]);
  }

  const options = Array.isArray(value.options) ? value.options.map(normalizeConversationOption) : undefined;
  return {
    id: typeof value.id === "string" ? value.id : "invalid_node",
    messages: Array.isArray(value.messages) ? value.messages.map(normalizeConversationMessage) : [],
    options,
    endConversation: value.endConversation === undefined ? undefined : Boolean(value.endConversation),
    tags: Array.isArray(value.tags) ? value.tags.map(String) : undefined,
  };
}

function normalizeConversationMessage(value: unknown): ConversationMessage {
  if (!isObject(value)) {
    return { text: String(value ?? "") };
  }

  return {
    text: typeof value.text === "string" ? value.text : "",
    portrait: typeof value.portrait === "string" ? value.portrait : undefined,
  };
}

function normalizeConversationOption(value: unknown): ConversationOption {
  if (!isObject(value)) {
    return { id: "invalid_option", text: "", nextNodeId: "start" };
  }

  return {
    id: typeof value.id === "string" ? value.id : "invalid_option",
    text: typeof value.text === "string" ? value.text : "",
    nextNodeId: typeof value.nextNodeId === "string" ? value.nextNodeId : "start",
  };
}

function nextUniqueId(prefix: string, existingIds: string[]): string {
  const existing = new Set(existingIds);
  let index = 1;
  let id = `${prefix}_${String(index).padStart(2, "0")}`;
  while (existing.has(id)) {
    index += 1;
    id = `${prefix}_${String(index).padStart(2, "0")}`;
  }
  return id;
}
