import styles from "./entityHealthBar.module.css";

type Props = {
  currentHealth: number;
  maxHealth: number;
};

export default function EntityHealthBar(props: Props) {
  const healthPercentage = Math.max(0, Math.min(100, (props.currentHealth / props.maxHealth) * 100));

  return (
    <div className={styles.container}>
      <div className={styles.barBackground}>
        <div
          className={styles.barFill}
          style={{ width: `${healthPercentage}%` }}
        />
      </div>
    </div>
  );
}

