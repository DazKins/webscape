import { useEffect, useRef } from "react";
import { useState } from "react";
import styles from "./interactionMenu.module.css";
import Game from "../../game/game";
import AbsolutePositioned from "./absolutePositioned";
import { InteractionMenuOpenEvent } from "../../events/interactionMenu";

type Props = {
  game: Game;
};

export default function InteractionMenu(props: Props) {
  const ref = useRef<HTMLDivElement>(null);

  const [interactionMenuOpen, setInteractionMenuOpen] = useState(false);
  const [name, setName] = useState("");
  const [interactionOptions, setInteractionOptions] = useState<string[]>([]);
  const [positionX, setPositionX] = useState(0);
  const [positionY, setPositionY] = useState(0);

  const handleInteractionMenuOpen = (event: Event) => {
    if (!(event instanceof InteractionMenuOpenEvent)) {
      return;
    }
    setInteractionMenuOpen(true);
    setName(event.name);
    setInteractionOptions(event.interactionOptions);
    setPositionX(event.positionX);
    setPositionY(event.positionY);
  };

  const onWindowClick = (event: MouseEvent) => {
    const element = ref.current;
    if (element && !element.contains(event.target as Node)) {
      setInteractionMenuOpen(false);
    } else {
      event.stopPropagation();
    }
  };

  useEffect(() => {
    props.game.addEventListener(
      "interactionMenuOpen",
      handleInteractionMenuOpen
    );
    window.addEventListener("click", onWindowClick);

    return () => {
      props.game.removeEventListener(
        "interactionMenuOpen",
        handleInteractionMenuOpen
      );
      window.removeEventListener("click", onWindowClick);
    };
  }, []);

  return (
    interactionMenuOpen && (
      <AbsolutePositioned top={positionY} left={positionX}>
        <div
          className={styles.container}
          ref={ref}
          onClick={(e) => e.stopPropagation()}
          onMouseMove={(e) => e.stopPropagation()}
        >
          <p className={styles.name}>{name}</p>
          <div className={styles.interactionOptions}>
            {interactionOptions.map((option, index) => (
              <button className={styles.interactionOption} key={index}>
                {option}
              </button>
            ))}
          </div>
        </div>
      </AbsolutePositioned>
    )
  );
}
