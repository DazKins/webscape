import Game from "../game/game";
import ChatBox from "./components/chatBox";

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
    </div>
  );
}
