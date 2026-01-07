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

export default function ChatBox(props: Props) {
  const [typedText, setTypedText] = useState("");
  const [chats, setChats] = useState<Chat[]>([]);

  useEffect(() => {
    const handler = (event: ChatMessageEvent) => {
      console.log("Chat message event", event);
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
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Chat</div>
      <div className={panelStyles.panelContent}>
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
    </div>
  );
}
