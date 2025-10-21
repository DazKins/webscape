class Input {
  keys: Record<string, boolean>;
  mousePosition: { x: number; y: number };
  mouseButtons: Record<number, boolean>;
  typedCharsBuffer: string;

  constructor() {
    this.keys = {};
    this.mousePosition = { x: 0, y: 0 };
    this.mouseButtons = {};
    this.typedCharsBuffer = "";

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

  onKeyDown(event: KeyboardEvent) {
    const key = event.key;
    if (key.length === 1) {
      this.typedCharsBuffer += key;
    }
    if (key === "Backspace") {
      this.typedCharsBuffer = this.typedCharsBuffer.slice(0, -1);
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
    return this.keys[key];
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

  getTypedCharsBuffer(): string {
    return this.typedCharsBuffer;
  }
}

export default Input;
