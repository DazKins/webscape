import { useEffect, useState } from "react";
import styles from "./chatBox.module.css";
import Game from "../game/game";

type Props = {
  game: Game;
};

export default function ChatBox(props: Props) {
  const [typedText, setTypedText] = useState("");

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
    <div className={styles.container}>
      <div className={styles.header}>Chat</div>
      <div className={styles.content}></div>
      <div className={styles.input}>
        {"> "}
        {typedText}*
      </div>
    </div>
  );
}
