import Game from "../game/game";
import ChatBox from "./components/chatBox";
import InteractionMenu from "./components/interactionMenu";
import Inventory from "./components/inventory";

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
      <Inventory game={props.game} />
    </div>
  );
}
