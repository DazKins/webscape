import { useEffect, useLayoutEffect, useRef } from "react";
import { useState } from "react";
import styles from "./interactionMenu.module.css";
import Game from "../../game/game";
import { InteractionMenuOpenEvent } from "../../events/interactionMenu";

type Props = {
  game: Game;
};

export default function InteractionMenu(props: Props) {
  const ref = useRef<HTMLDivElement>(null);

  const [interactionMenuOpen, setInteractionMenuOpen] = useState(false);
  const [entityId, setEntityId] = useState("");
  const [name, setName] = useState("");
  const [interactionOptions, setInteractionOptions] = useState<string[]>([]);
  const [positionX, setPositionX] = useState(0);
  const [positionY, setPositionY] = useState(0);

  const handleInteractionMenuOpen = (event: Event) => {
    if (!(event instanceof InteractionMenuOpenEvent)) {
      return;
    }
    setInteractionMenuOpen(true);
    setEntityId(event.entityId);
    setName(event.name);
    setInteractionOptions(event.interactionOptions);
    setPositionX(event.positionX);
    setPositionY(event.positionY);
  };

  const onWindowPointerDown = (event: PointerEvent) => {
    const element = ref.current;
    if (element && !element.contains(event.target as Node)) {
      props.game.setPointerOverUi(false);
      setInteractionMenuOpen(false);
    } else {
      props.game.setPointerOverUi(true);
    }
  };

  useEffect(() => {
    props.game.addEventListener(
      "interactionMenuOpen",
      handleInteractionMenuOpen
    );

    return () => {
      props.game.removeEventListener(
        "interactionMenuOpen",
        handleInteractionMenuOpen
      );
    };
  }, []);

  useEffect(() => {
    if (!interactionMenuOpen) {
      return;
    }

    window.addEventListener("pointerdown", onWindowPointerDown);
    return () => {
      window.removeEventListener("pointerdown", onWindowPointerDown);
    };
  }, [interactionMenuOpen]);

  useLayoutEffect(() => {
    if (!interactionMenuOpen) return;
    const positionMenu = () => {
      const element = ref.current;
      if (!element) return;
      const bounds = element.getBoundingClientRect();
      element.style.left = `${Math.max(8, Math.min(positionX, window.innerWidth - bounds.width - 8))}px`;
      element.style.top = `${Math.max(8, Math.min(positionY, window.innerHeight - bounds.height - 8))}px`;
    };
    positionMenu();
    window.addEventListener("resize", positionMenu);
    return () => window.removeEventListener("resize", positionMenu);
  }, [interactionMenuOpen, positionX, positionY, name, interactionOptions]);

  const handleInteractionOptionClick = (option: string): void => {
    props.game.handleInteractionOptionClick(entityId, option);
    props.game.setPointerOverUi(false);
    setInteractionMenuOpen(false);
  };

  return (
    interactionMenuOpen && (
      <div
        className={styles.container}
        ref={ref}
        onClick={(e) => e.stopPropagation()}
        onPointerMove={(e) => e.stopPropagation()}
        onPointerDown={(e) => {
          e.stopPropagation();
          props.game.setPointerOverUi(true);
        }}
        onPointerEnter={() => props.game.setPointerOverUi(true)}
        onPointerLeave={() => props.game.setPointerOverUi(false)}
      >
        <p className={styles.name}>{name}</p>
        <div className={styles.interactionOptions}>
          {interactionOptions.map((option, index) => (
            <button
              className={styles.interactionOption}
              key={index}
              onClick={() => handleInteractionOptionClick(option)}
            >
              {option === "chop" ? "Chop" : option === "bank" ? "Bank" : option}
            </button>
          ))}
        </div>
      </div>
    )
  );
}
