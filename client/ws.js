export class WebSocketClient {
  constructor(options = {}) {
    this.ws = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = options.maxReconnectAttempts || 5;
    this.reconnectDelay = options.reconnectDelay || 1000; // 1 second

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

    this.ws.onerror = (error) => {
      this.onError(error);
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

  handleMessage(data) {
    // Call the user-provided message handler
    this.onMessage(data);
  }

  sendMessage(msg) {
    if (!this.isConnected) {
      console.error("WebSocket is not connected");
      return;
    }

    try {
      const message = JSON.stringify(msg);
      this.ws.send(message);
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
