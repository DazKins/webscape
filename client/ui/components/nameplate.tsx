import styles from "./nameplate.module.css";

export default function Nameplate({ text }: { text: string }) {
  return <div className={styles.nameplate}>{text}</div>;
}
