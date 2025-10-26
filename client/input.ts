export interface InputReceiver {
  onKeyDown(key: string): void;
}

class Input {
  keys: Record<string, boolean>;
  mousePosition: { x: number; y: number };
  mouseButtons: Record<number, boolean>;

  activeReceiver?: InputReceiver;

  constructor() {
    this.keys = {};
    this.mousePosition = { x: 0, y: 0 };
    this.mouseButtons = {};

    this.onKeyDown = this.onKeyDown.bind(this);
    this.onKeyUp = this.onKeyUp.bind(this);
    this.onMouseMove = this.onMouseMove.bind(this);
    this.onMouseDown = this.onMouseDown.bind(this);
    this.onMouseUp = this.onMouseUp.bind(this);

    window.addEventListener("keydown", this.onKeyDown);
    window.addEventListener("keyup", this.onKeyUp);
    window.addEventListener("mousemove", this.onMouseMove);
    window.addEventListener("mousedown", this.onMouseDown);
    window.addEventListener("mouseup", this.onMouseUp);
  }

  setActiveReceiver(receiver: InputReceiver) {
    this.activeReceiver = receiver;
  }

  onKeyDown(event: KeyboardEvent) {
    if (
      event.target instanceof HTMLInputElement ||
      event.target instanceof HTMLTextAreaElement
    )
      return;

    const key = event.key;

    if (this.activeReceiver) {
      this.activeReceiver.onKeyDown(key);
    }

    this.keys[key.toLowerCase()] = true;
  }

  onKeyUp(event: KeyboardEvent) {
    this.keys[event.key.toLowerCase()] = false;
  }

  onMouseMove(event: MouseEvent) {
    this.mousePosition.x = event.clientX;
    this.mousePosition.y = event.clientY;
  }

  onMouseDown(event: MouseEvent) {
    this.mouseButtons[event.button] = true;
  }

  onMouseUp(event: MouseEvent) {
    this.mouseButtons[event.button] = false;
  }

  getKey(key: string) {
    return this.keys[key.toLowerCase()];
  }

  getMousePosition() {
    return { ...this.mousePosition };
  }

  getMouseButton(button: number) {
    return this.mouseButtons[button] || false;
  }

  registerClickCallback(callback: () => void) {
    window.addEventListener("click", callback);
  }

  registerRightClickCallback(callback: (event: MouseEvent) => void) {
    window.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      callback(event);
    });
  }
}

export default Input;
