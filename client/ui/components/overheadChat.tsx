import styles from "./overheadChat.module.css";

type Props = {
  text: string;
};

export default function OverheadChat(props: Props) {
  return (
    <div className={styles.container}>
      <p className={styles.text}>{props.text}</p>
    </div>
  );
}
