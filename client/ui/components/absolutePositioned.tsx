type Props = {
  top: number;
  left: number;
  children: React.ReactNode;
};

export default function AbsolutePositioned(props: Props) {
  return (
    <div style={{ position: "absolute", top: props.top, left: props.left }}>
      {props.children}
    </div>
  );
}