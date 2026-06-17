import { useEffect, useState } from "react";
import Game from "../../game/game";
import {
  ConversationCloseEventName,
  ConversationEvent,
  ConversationEventName,
  type ConversationPayload,
} from "../../events/conversation";
import panelStyles from "./uiPanel.module.css";
import styles from "./conversationPanel.module.css";

type Props = {
  game: Game;
};

export default function ConversationPanel(props: Props) {
  const [conversation, setConversation] = useState<ConversationPayload | null>(null);
  const [messageIndex, setMessageIndex] = useState(0);

  useEffect(() => {
    const handler = (event: ConversationEvent) => {
      setConversation(event.payload);
      setMessageIndex(0);
    };
    const closeHandler = () => {
      props.game.setPointerOverUi(false);
      setConversation(null);
      setMessageIndex(0);
    };

    props.game.addEventListener(ConversationEventName, handler as EventListener);
    props.game.addEventListener(ConversationCloseEventName, closeHandler);
    return () => {
      props.game.removeEventListener(ConversationEventName, handler as EventListener);
      props.game.removeEventListener(ConversationCloseEventName, closeHandler);
    };
  }, [props.game]);

  if (!conversation) {
    return null;
  }

  const messages = conversation.messages.length > 0 ? conversation.messages : [{ text: "" }];
  const currentMessage = messages[Math.min(messageIndex, messages.length - 1)];
  const showingOptions = messageIndex >= messages.length - 1;
  const entityName = props.game.getEntityName(conversation.targetEntityId, "Conversation");

  function handleContinue() {
    if (!conversation) {
      return;
    }
    if (messageIndex < messages.length - 1) {
      setMessageIndex((current) => current + 1);
      return;
    }
    if (conversation.endConversation) {
      props.game.handleConversationClose(conversation.conversationId, conversation.nodeId);
      props.game.setPointerOverUi(false);
      setConversation(null);
    }
  }

  function handleOption(optionId: string) {
    if (!conversation) {
      return;
    }

    props.game.handleConversationOptionClick(
      conversation.conversationId,
      conversation.nodeId,
      optionId
    );
  }

  return (
    <div
      className={`${panelStyles.panel} ${styles.container}`}
      onClick={(event) => event.stopPropagation()}
      onMouseMove={(event) => event.stopPropagation()}
      onMouseEnter={() => props.game.setPointerOverUi(true)}
      onMouseLeave={() => props.game.setPointerOverUi(false)}
    >
      <div className={panelStyles.panelHeader}>{entityName}</div>
      <div className={`${panelStyles.panelContent} ${styles.content}`}>
        <div className={styles.message}>
          <div className={styles.text}>{currentMessage.text}</div>
        </div>

        <div className={styles.actions}>
          {!showingOptions ? (
            <button className={styles.continueButton} type="button" onClick={handleContinue}>
              Continue
            </button>
          ) : conversation.endConversation ? (
            <button className={styles.closeButton} type="button" onClick={handleContinue}>
              Close
            </button>
          ) : (
            (conversation.options ?? []).map((option) => (
              <button
                className={styles.option}
                key={option.id}
                type="button"
                onClick={() => handleOption(option.id)}
              >
                {option.text}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
