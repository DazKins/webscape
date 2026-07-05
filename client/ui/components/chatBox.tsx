import { useEffect, useState } from "react";
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
  const [typedText, setTypedText] = useState("");
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

  return (
    <>
      <div className={`${panelStyles.panelContent} ${styles.messages}`}>
        {chats.map((chat, index) => (
          <div key={index} className={styles.chat}>
            <span className={styles.chatFrom}>{chat.from}</span>: {chat.message}
          </div>
        ))}
      </div>
      <div className={styles.input}>
        <span className={styles.inputPrefix}>{"> "}</span>
        {typedText}*
      </div>
      <form
        className={styles.mobileComposer}
        onSubmit={(event) => {
          event.preventDefault();
          props.game.sendTypedChatText();
        }}
      >
        <input
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
    </>
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
