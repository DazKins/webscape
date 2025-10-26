import styles from "./entityNamecard.module.css";

type Props = {
  name: string;
};

export default function EntityNamecard(props: Props) {
  return (
    <div className={styles.container}>
      <p className={styles.text}>{props.name}</p>
    </div>
  );
}
