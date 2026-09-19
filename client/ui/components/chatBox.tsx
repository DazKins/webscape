import { useEffect, useRef, useState } from "react";
import styles from "./chatBox.module.css";
import panelStyles from "./uiPanel.module.css";
import Game from "../../game/game";
import { ChatMessageEvent, ChatMessageEventName } from "../../events/chat";

type Props = {
  game: Game;
};

type Chat = {
  message: string;
  from: string;
};

export function ChatBoxContent(props: Props) {
  const [typedText, setTypedText] = useState(props.game.typedChatText);
  const desktopInput = useRef<HTMLInputElement>(null);
  const mobileInput = useRef<HTMLInputElement>(null);
  const [chats, setChats] = useState<Chat[]>([]);

  useEffect(() => {
    const handler = (event: ChatMessageEvent) => {
      setChats((prevChats) => [
        ...prevChats, {
          message: event.message,
          from: event.from
        }
      ]);
    };
    props.game.addEventListener(ChatMessageEventName, handler as EventListener);
    return () => {
      props.game.removeEventListener(ChatMessageEventName, handler as EventListener);
    };
  }, [props.game]);

  useEffect(() => {
    const handler = (event: CustomEvent<string>) => {
      setTypedText(event.detail);
    };
    props.game.addEventListener(
      "typedChatTextChanged",
      handler as EventListener
    );
    return () => {
      props.game.removeEventListener(
        "typedChatTextChanged",
        handler as EventListener
      );
    };
  }, [props.game]);

  useEffect(() => {
    const handleChatKey = (event: KeyboardEvent) => {
      if (event.isComposing || event.keyCode === 229) return;
      if (event.key === "Enter") {
        event.preventDefault();
        props.game.sendTypedChatText();
      } else if (event.key === "Escape") {
        props.game.setTypedChatText("");
      }
    };
    const input = desktopInput.current;
    input?.addEventListener("keydown", handleChatKey);
    return () => {
      input?.removeEventListener("keydown", handleChatKey);
    };
  }, [props.game]);

  return (
    <div
      className={styles.content}
      onClick={(event) => {
        if (event.target instanceof Element && event.target.closest("input, button, a, select, textarea")) return;
        const input = [desktopInput.current, mobileInput.current].find(
          (input) => input && input.getClientRects().length > 0
        );
        input?.focus({ preventScroll: true });
      }}
    >
      <div className={`${panelStyles.panelContent} ${styles.messages}`}>
        {chats.map((chat, index) => (
          <div key={index} className={styles.chat}>
            <span className={styles.chatFrom}>{chat.from}</span>: {chat.message}
          </div>
        ))}
      </div>
      <div className={styles.input}>
        <span className={styles.inputPrefix}>{"> "}</span>
        <span className={styles.editor}>
          <span className={styles.inputSizer} aria-hidden="true">{typedText}</span>
          <input
            ref={desktopInput}
            className={styles.desktopInput}
            type="text"
            aria-label="Chat message"
            autoComplete="off"
            spellCheck={false}
            value={typedText}
            onChange={(event) => props.game.setTypedChatText(event.target.value)}
          />
        </span>
      </div>
      <form
        className={styles.mobileComposer}
        onSubmit={(event) => {
          event.preventDefault();
          props.game.sendTypedChatText();
        }}
      >
        <input
          ref={mobileInput}
          className={styles.mobileInput}
          value={typedText}
          placeholder="Message"
          aria-label="Chat message"
          onChange={(event) => props.game.setTypedChatText(event.target.value)}
          onFocus={() => props.game.setPointerOverUi(true)}
          onBlur={() => props.game.setPointerOverUi(false)}
        />
        <button className={styles.sendButton} type="submit">
          Send
        </button>
      </form>
    </div>
  );
}

export default function ChatBox(props: Props) {
  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Chat</div>
      <ChatBoxContent game={props.game} />
    </div>
  );
}
