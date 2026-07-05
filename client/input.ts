export interface InputReceiver {
  onKeyDown(key: string): void;
}

type PointerCallbacks = {
  onTap?: (event: PointerEvent) => void;
  onLongPress?: (event: PointerEvent) => void;
};

const LONG_PRESS_MS = 450;
const TAP_MOVE_THRESHOLD_PX = 12;

class Input {
  keys: Record<string, boolean>;
  pointerPosition: { x: number; y: number };
  pointerButtons: Record<number, boolean>;
  pointerBlocked: boolean;

  private pointerCallbacks: PointerCallbacks;
  private activePointer:
    | {
        id: number;
        x: number;
        y: number;
        longPressTimer: number;
        longPressed: boolean;
      }
    | undefined;

  activeReceiver?: InputReceiver;

  constructor() {
    this.keys = {};
    this.pointerPosition = { x: 0, y: 0 };
    this.pointerButtons = {};
    this.pointerBlocked = false;
    this.pointerCallbacks = {};

    this.onKeyDown = this.onKeyDown.bind(this);
    this.onKeyUp = this.onKeyUp.bind(this);
    this.onPointerMove = this.onPointerMove.bind(this);
    this.onPointerDown = this.onPointerDown.bind(this);
    this.onPointerUp = this.onPointerUp.bind(this);
    this.onPointerCancel = this.onPointerCancel.bind(this);

    window.addEventListener("keydown", this.onKeyDown);
    window.addEventListener("keyup", this.onKeyUp);
    window.addEventListener("pointermove", this.onPointerMove);
    window.addEventListener("pointerdown", this.onPointerDown);
    window.addEventListener("pointerup", this.onPointerUp);
    window.addEventListener("pointercancel", this.onPointerCancel);
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

  onPointerMove(event: PointerEvent) {
    this.pointerPosition.x = event.clientX;
    this.pointerPosition.y = event.clientY;
  }

  onPointerDown(event: PointerEvent) {
    this.pointerPosition.x = event.clientX;
    this.pointerPosition.y = event.clientY;
    this.pointerButtons[event.button] = true;

    window.clearTimeout(this.activePointer?.longPressTimer);
    const activePointer = {
      id: event.pointerId,
      x: event.clientX,
      y: event.clientY,
      longPressTimer: 0,
      longPressed: false,
    };

    activePointer.longPressTimer = window.setTimeout(() => {
      if (this.pointerBlocked || this.activePointer?.id !== event.pointerId) {
        return;
      }
      this.activePointer.longPressed = true;
      this.pointerCallbacks.onLongPress?.(event);
    }, LONG_PRESS_MS);

    this.activePointer = activePointer;
  }

  onPointerUp(event: PointerEvent) {
    this.pointerPosition.x = event.clientX;
    this.pointerPosition.y = event.clientY;
    this.pointerButtons[event.button] = false;

    const activePointer = this.activePointer;
    if (!activePointer || activePointer.id !== event.pointerId) {
      return;
    }

    window.clearTimeout(activePointer.longPressTimer);
    this.activePointer = undefined;

    if (this.pointerBlocked || activePointer.longPressed || event.button !== 0) {
      return;
    }

    const movedDistance = Math.hypot(
      event.clientX - activePointer.x,
      event.clientY - activePointer.y
    );
    if (movedDistance <= TAP_MOVE_THRESHOLD_PX) {
      this.pointerCallbacks.onTap?.(event);
    }
  }

  onPointerCancel(event: PointerEvent) {
    this.pointerButtons[event.button] = false;
    window.clearTimeout(this.activePointer?.longPressTimer);
    this.activePointer = undefined;
  }

  getKey(key: string) {
    return this.keys[key.toLowerCase()];
  }

  getPointerPosition() {
    return { ...this.pointerPosition };
  }

  getMousePosition() {
    return this.getPointerPosition();
  }

  getMouseButton(button: number) {
    return this.pointerButtons[button] || false;
  }

  setPointerBlocked(blocked: boolean) {
    this.pointerBlocked = blocked;
  }

  isPointerBlocked() {
    return this.pointerBlocked;
  }

  registerPointerCallbacks(callbacks: PointerCallbacks) {
    this.pointerCallbacks = callbacks;
  }

  registerClickCallback(callback: () => void) {
    this.pointerCallbacks.onTap = callback;
  }

  registerRightClickCallback(callback: (event: MouseEvent) => void) {
    window.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      callback(event);
    });
  }
}

export default Input;
