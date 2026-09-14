type WebSocketClientOptions = {
  maxReconnectAttempts?: number;
  reconnectDelay?: number;
  beforeConnect?: () => Promise<boolean>;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onUnavailable?: () => void;
  onError?: (event: Event) => void;
  onMessage?: (data: any) => void;
};

export class WebSocketClient {
  private ws?: WebSocket;
  private timer?: number;
  private generation = 0;
  private stopped = true;
  private connecting = false;
  private attempts = 0;
  isConnected = false;

  constructor(private options: WebSocketClientOptions = {}) {}

  async connect() {
    if (this.connecting || this.isConnected) return;
    this.stopped = false;
    this.connecting = true;
    const generation = this.generation;
    try {
      const allowed = await (this.options.beforeConnect?.() ?? Promise.resolve(true));
      if (this.stopped || generation !== this.generation) return;
      if (!allowed) { this.stopped = true; return; }
      const protocol = location.protocol === "https:" ? "wss" : "ws";
      const ws = new WebSocket(`${protocol}://${location.host}/ws`);
      this.ws = ws;
      const current = () => !this.stopped && generation === this.generation && this.ws === ws;
      ws.onopen = () => {
        if (!current()) { ws.close(); return; }
        this.isConnected = true;
        this.attempts = 0;
        this.options.onConnect?.();
      };
      ws.onmessage = (event) => {
        if (!current()) return;
        try { this.options.onMessage?.(JSON.parse(event.data)); }
        catch (error) { console.error("Invalid game message", error); }
      };
      ws.onclose = () => {
        if (!current()) return;
        this.ws = undefined;
        this.isConnected = false;
        this.options.onDisconnect?.();
        this.reconnect();
      };
      ws.onerror = (event) => { if (current()) this.options.onError?.(event); };
    } catch {
      if (!this.stopped && generation === this.generation) this.reconnect();
    } finally {
      if (generation === this.generation) this.connecting = false;
    }
  }

  private reconnect() {
    if (this.stopped) return;
    if (this.attempts++ >= (this.options.maxReconnectAttempts ?? 5)) {
      this.options.onUnavailable?.();
      return;
    }
    this.timer = window.setTimeout(() => { void this.connect(); }, this.options.reconnectDelay ?? 1000);
  }

  sendMessage(message: any) {
    if (this.ws?.readyState === WebSocket.OPEN) this.ws.send(JSON.stringify(message));
  }

  disconnect() {
    this.stopped = true;
    this.generation++;
    this.connecting = false;
    this.isConnected = false;
    this.attempts = 0;
    window.clearTimeout(this.timer);
    this.ws?.close();
    this.ws = undefined;
  }
}
