import Game from "../game/game";
import ChatBox from "./components/chatBox";
import InteractionMenu from "./components/interactionMenu";
import Inventory from "./components/inventory";
import CombatLog from "./components/combatLog";
import ConversationPanel from "./components/conversationPanel";

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
      <ConversationPanel game={props.game} />
      <Inventory game={props.game} />
      <CombatLog game={props.game} />
    </div>
  );
}
