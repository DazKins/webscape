import { useEffect, useState } from "react";
import Game from "../game/game";
import AbsolutePositioned from "./components/absolutePositioned";
import ChatBox from "./components/chatBox";
import InteractionMenu from "./components/interactionMenu";

type Props = {
  game: Game;
};

export default function UiRoot(props: Props) {

  return (
    <div
      style={{
        position: "absolute",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
      }}
    >
      <ChatBox game={props.game} />
      <InteractionMenu game={props.game} />
    </div>
  );
}
