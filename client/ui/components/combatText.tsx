import styles from "./combatText.module.css";

type Props = {
  text: string;
  kind: string;
};

export default function CombatText(props: Props) {
  const kindClass =
    props.kind === "crit"
      ? styles.crit
      : props.kind === "miss"
      ? styles.miss
      : styles.hit;

  return (
    <div className={`${styles.container} ${kindClass}`}>
      {props.text}
    </div>
  );
}
