type WebSocketClientOptions = {
  maxReconnectAttempts?: number;
  reconnectDelay?: number;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (event: Event) => void;
  onMessage?: (data: any) => void;
};

export class WebSocketClient {
  ws: WebSocket | undefined;
  isConnected: boolean;
  reconnectAttempts: number;
  maxReconnectAttempts: number;
  reconnectDelay: number;
  onConnect: () => void;
  onDisconnect: () => void;
  onError: (event: Event) => void;
  onMessage: (data: any) => void;

  constructor(options: WebSocketClientOptions = {}) {
    this.ws = undefined;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = options.maxReconnectAttempts || 5;
    this.reconnectDelay = options.reconnectDelay || 1000; // 1 second
    this.isConnected = false;

    // Event handlers
    this.onConnect = options.onConnect || (() => {});
    this.onDisconnect = options.onDisconnect || (() => {});
    this.onError =
      options.onError || ((error) => console.error("WebSocket error:", error));
    this.onMessage =
      options.onMessage || ((data) => console.log("Received message:", data));
  }

  connect() {
    try {
      const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
      this.ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws`);
      this.setupEventListeners();
    } catch (error) {
      console.error("WebSocket connection error:", error);
      this.handleReconnect();
    }
  }

  setupEventListeners() {
    if (!this.ws) {
      throw new Error("WebSocket is not initialized");
    }

    this.ws.onopen = () => {
      console.log("WebSocket connection established");
      this.isConnected = true;
      this.reconnectAttempts = 0;
      this.onConnect();
    };

    this.ws.onmessage = (event) => {
      try {
        const jsonMsg = JSON.parse(event.data);
        this.handleMessage(jsonMsg);
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    this.ws.onclose = () => {
      console.log("WebSocket connection closed");
      this.isConnected = false;
      this.onDisconnect();
      this.handleReconnect();
    };

    this.ws.onerror = (event: Event) => {
      this.onError(event);
    };
  }

  handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(
        `Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`
      );
      setTimeout(() => this.connect(), this.reconnectDelay);
    } else {
      console.error("Max reconnection attempts reached");
    }
  }

  handleMessage(message: any) {
    this.onMessage(message);
  }

  sendMessage(message: any) {
    if (!this.ws) {
      throw new Error("WebSocket is not initialized");
    }

    console.log("Sending message", message);

    try {
      const messageString = JSON.stringify(message);
      this.ws.send(messageString);
    } catch (error) {
      console.error("Error sending message:", error);
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}
