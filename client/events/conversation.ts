export const ConversationEventName = "conversation";

export type ConversationMessage = {
  text: string;
  portrait?: string;
};

export type ConversationOption = {
  id: string;
  text: string;
};

export type ConversationPayload = {
  conversationId: string;
  targetEntityId: string;
  nodeId: string;
  messages: ConversationMessage[];
  options?: ConversationOption[];
  endConversation: boolean;
};

export class ConversationEvent extends Event {
  constructor(public payload: ConversationPayload) {
    super(ConversationEventName);
  }
}
