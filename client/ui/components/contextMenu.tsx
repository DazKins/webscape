import { useEffect, useLayoutEffect, useRef } from "react";
import { createPortal } from "react-dom";
import Game from "../../game/game";
import styles from "./contextMenu.module.css";

type Props = {
  game: Game;
  name: string;
  x: number;
  y: number;
  actions: { label: string; onSelect: () => void }[];
  onClose: () => void;
  trigger?: HTMLElement | null;
};

export default function ContextMenu({ game, name, x, y, actions, onClose, trigger }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    const element = ref.current;
    if (!element) return;
    const bounds = element.getBoundingClientRect();
    element.style.left = `${Math.max(8, Math.min(x, window.innerWidth - bounds.width - 8))}px`;
    element.style.top = `${Math.max(8, Math.min(y, window.innerHeight - bounds.height - 8))}px`;
  }, [x, y, name, actions]);

  useEffect(() => {
    ref.current?.querySelector<HTMLButtonElement>("button")?.focus({ preventScroll: true });
  }, [x, y, name]);

  useEffect(() => {
    const close = () => {
      game.setPointerOverUi(false);
      onClose();
    };
    const onPointerDown = (event: PointerEvent) => {
      if (!ref.current?.contains(event.target as Node)) close();
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape" && event.key !== "Tab") return;
      if (event.key === "Escape") event.preventDefault();
      close();
      trigger?.focus({ preventScroll: true });
    };
    const onScroll = (event: Event) => {
      if (!ref.current?.contains(event.target as Node)) close();
    };
    window.addEventListener("pointerdown", onPointerDown, true);
    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("resize", close);
    window.addEventListener("scroll", onScroll, true);
    return () => {
      window.removeEventListener("pointerdown", onPointerDown, true);
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("resize", close);
      window.removeEventListener("scroll", onScroll, true);
    };
  }, [game, onClose, trigger]);

  useEffect(() => () => game.setPointerOverUi(false), [game]);

  return createPortal(
    <div
      ref={ref}
      className={styles.container}
      role="menu"
      aria-label={`${name} actions`}
      style={{ left: x, top: y }}
      onContextMenu={(event) => { event.preventDefault(); event.stopPropagation(); }}
      onClick={(event) => event.stopPropagation()}
      onPointerMove={(event) => event.stopPropagation()}
      onPointerDown={(event) => {
        event.stopPropagation();
        game.setPointerOverUi(true);
      }}
      onPointerEnter={() => game.setPointerOverUi(true)}
      onPointerLeave={() => game.setPointerOverUi(false)}
      onKeyDown={(event) => {
        if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
        event.preventDefault();
        event.stopPropagation();
        const buttons = Array.from(event.currentTarget.querySelectorAll<HTMLButtonElement>("button"));
        const current = buttons.indexOf(document.activeElement as HTMLButtonElement);
        const next = event.key === "Home" ? 0 : event.key === "End" ? buttons.length - 1 :
          (current + (event.key === "ArrowDown" ? 1 : -1) + buttons.length) % buttons.length;
        buttons[next]?.focus();
      }}
    >
      <div className={styles.name}>{name}</div>
      {actions.map((action) => (
        <button type="button" role="menuitem" key={action.label} onClick={() => {
          action.onSelect();
          game.setPointerOverUi(false);
          onClose();
        }}>
          {action.label}
        </button>
      ))}
    </div>,
    document.getElementById("uiLayerRoot")!,
  );
}
