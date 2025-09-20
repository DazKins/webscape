class Input {
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

  onKeyDown(event) {
    this.keys[event.key] = true;
  }

  onKeyUp(event) {
    this.keys[event.key] = false;
  }

  onMouseMove(event) {
    this.mousePosition.x = event.clientX;
    this.mousePosition.y = event.clientY;
  }

  onMouseDown(event) {
    this.mouseButtons[event.button] = true;
  }

  onMouseUp(event) {
    this.mouseButtons[event.button] = false;
  }

  getKey(key) {
    return this.keys[key];
  }

  getMousePosition() {
    return { ...this.mousePosition };
  }

  getMouseButton(button) {
    return this.mouseButtons[button] || false;
  }

  registerClickCallback(callback) {
    window.addEventListener("click", callback);
  }

  registerRightClickCallback(callback) {
    window.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      callback(event);
    });
  }
}

export default Input;
