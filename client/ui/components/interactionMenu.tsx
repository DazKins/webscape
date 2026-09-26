import { useEffect, useState } from "react";
import Game from "../../game/game";
import { InteractionMenuOpenEvent, InteractionMenuOpenEventName } from "../../events/interactionMenu";
import ContextMenu from "./contextMenu";

type Props = {
  game: Game;
};

export default function InteractionMenu({ game }: Props) {
  const [menu, setMenu] = useState<InteractionMenuOpenEvent | null>(null);

  useEffect(() => {
    const onOpen = (event: Event) => {
      if (event instanceof InteractionMenuOpenEvent) setMenu(event);
    };
    game.addEventListener(InteractionMenuOpenEventName, onOpen);
    return () => game.removeEventListener(InteractionMenuOpenEventName, onOpen);
  }, [game]);

  return menu && (
    <ContextMenu
      game={game}
      name={menu.name}
      x={menu.positionX}
      y={menu.positionY}
      onClose={() => setMenu(null)}
      actions={menu.interactionOptions.map((option) => ({
        label: option.charAt(0).toUpperCase() + option.slice(1),
        onSelect: () => game.handleInteractionOptionClick(menu.entityId, option),
      }))}
    />
  );
}
