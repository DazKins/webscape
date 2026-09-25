package server

import (
	"log"
	"net/http"
	"sync"
	"time"
	"webscape/server/auth"
	"webscape/server/config"
	"webscape/server/game/model"
	"webscape/server/message"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type client struct {
	// Protected by wsServer.actions, independently of transport pongs.
	registered   bool
	lastActivity time.Time
	warned       bool

	conn    *websocket.Conn
	send    chan []byte
	id      string
	session *auth.Session
	ready   chan struct{}
	done    chan struct{}
}

type MessageHandler func(clientID string, message string)
type ConnectHandler func(clientID string)
type DisconnectHandler func(clientID string)

type wsServer struct {
	// Serialize commands, admission and cleanup so a retired socket cannot mutate
	// a character after another connection takes ownership.
	actions  sync.Mutex
	settings config.ConnectionConfig

	clients      map[string]*client
	broadcast    chan []byte
	register     chan *client
	unregister   chan *client
	mutex        sync.Mutex
	onMessage    MessageHandler
	onConnect    ConnectHandler
	onDisconnect DisconnectHandler
}

func NewWsServer(settings ...config.ConnectionConfig) *wsServer {
	cfg := config.DefaultConnections()
	if len(settings) > 0 {
		cfg = settings[0]
	}

	wss := &wsServer{
		settings:   cfg,
		clients:    make(map[string]*client),
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
	}

	go wss.run()

	return wss
}

func (w *wsServer) SetIncomingMessageHandler(handler MessageHandler) {
	w.onMessage = handler
}

func (w *wsServer) SetConnectHandler(handler ConnectHandler) {
	w.onConnect = handler
}

func (w *wsServer) SetDisconnectHandler(handler DisconnectHandler) {
	w.onDisconnect = handler
}

// Broadcast sends a message to all connected clients
func (w *wsServer) Broadcast(message message.Message) {
	w.broadcast <- []byte(message.Marshal())
}

// SendToClient sends a message to a specific client identified by clientID
func (w *wsServer) SendToClient(clientID string, message message.Message) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if client, ok := w.clients[clientID]; ok && client.session.Active() {
		select {
		case client.send <- []byte(message.Marshal()):
		default:
			client.conn.Close()
		}
	}
}

func (w *wsServer) run() {
	for {
		select {
		case c := <-w.register:
			w.actions.Lock()
			w.mutex.Lock()
			w.clients[c.id] = c
			w.mutex.Unlock()
			c.session.WithActive(func() {
				if w.onConnect != nil {
					w.onConnect(c.id)
				}
			})
			close(c.ready)
			w.actions.Unlock()
		case c := <-w.unregister:
			w.actions.Lock()
			w.remove(c)
			w.actions.Unlock()
		case msg := <-w.broadcast:
			w.mutex.Lock()
			for _, c := range w.clients {
				if !c.session.Active() {
					continue
				}
				select {
				case c.send <- msg:
				default:
					c.conn.Close()
				}
			}
			w.mutex.Unlock()
		}
	}
}

// remove requires actions. Cleanup runs exactly once, including when a replaced
// socket's read goroutine subsequently reports its disconnect.
func (w *wsServer) remove(c *client) {
	w.mutex.Lock()
	present := w.clients[c.id] == c
	if present {
		delete(w.clients, c.id)
		close(c.send)
	}
	w.mutex.Unlock()
	if present && w.onDisconnect != nil {
		w.onDisconnect(c.id)
	}
}

func (w *wsServer) getClient(id string) *client {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.clients[id]
}

func (w *wsServer) handleMessage(id, raw string) {
	w.actions.Lock()
	defer w.actions.Unlock()
	c := w.getClient(id)
	if c == nil {
		return
	}
	if c.registered && w.expireIdle(c, time.Now()) {
		return
	}
	c.session.WithActive(func() {
		if w.onMessage != nil {
			w.onMessage(id, raw)
		}
	})
}

// RecordActivity is called only from the serialized command handler after
// registration or player input. Pongs and simulation ticks never call it.
func (w *wsServer) RecordActivity(id string) {
	c := w.getClient(id)
	if c == nil {
		return
	}
	c.registered = true
	c.lastActivity = time.Now()
	c.warned = false
	// Acknowledge the authoritative deadline so resumed tabs can identify an
	// expired game session even if its close frame was lost during suspension.
	duration := int64(w.settings.IdleTimeoutSeconds) * 1000
	w.SendToClient(id, message.NewInactivityMessage(duration, duration, false))
}

// ReplacePlayer is called only from validated, explicit registration intent,
// under actions. Identity comes from the authenticated connection, never JSON.
func (w *wsServer) ReplacePlayer(id string) {
	current := w.getClient(id)
	if current == nil || !current.session.Active() {
		return
	}
	w.mutex.Lock()
	var previous []*client
	for _, c := range w.clients {
		if c != current && c.registered && c.session.PlayerID == current.session.PlayerID {
			previous = append(previous, c)
		}
	}
	w.mutex.Unlock()
	for _, c := range previous {
		w.kick(c, message.CloseReplaced, "Your game was opened elsewhere")
	}
}

func (w *wsServer) kick(c *client, code int, reason string) {
	c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
	c.conn.Close()
	w.remove(c)
}

func (w *wsServer) expireIdle(c *client, now time.Time) bool {
	timeout := time.Duration(w.settings.IdleTimeoutSeconds) * time.Second
	remaining := timeout - now.Sub(c.lastActivity)
	if remaining <= 0 {
		w.kick(c, message.CloseInactive, "Disconnected for inactivity")
		return true
	}
	if !c.warned && remaining <= time.Duration(w.settings.IdleWarningSeconds)*time.Second {
		c.warned = true
		w.SendToClient(c.id, message.NewInactivityMessage(remaining.Milliseconds(), timeout.Milliseconds(), true))
	}
	return false
}

func (w *wsServer) checkIdle(c *client) {
	w.actions.Lock()
	defer w.actions.Unlock()
	if w.getClient(c.id) == c && c.registered {
		w.expireIdle(c, time.Now())
	}
}

func (c *client) readPump(onMessage MessageHandler, unregister chan *client, pongTimeout time.Duration) {
	defer func() {
		close(c.done)
		unregister <- c
		c.conn.Close()
	}()

	<-c.ready
	c.conn.SetReadLimit(16 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(pongTimeout)) })
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if onMessage != nil {
			onMessage(c.id, string(msg))
		}
	}
}

func (c *client) writePump(wss *wsServer) {
	ping := time.NewTicker(time.Duration(wss.settings.PingIntervalSeconds) * time.Second)
	idle := time.NewTicker(time.Second)
	defer ping.Stop()
	defer idle.Stop()
	defer c.conn.Close()
	for {
		select {
		case <-c.done:
			return
		case <-idle.C:
			wss.checkIdle(c)
		case <-ping.C:
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				return
			}
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			// Revocation has a dedicated close frame in the session watcher.
			if !c.session.Active() {
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

func (wss *wsServer) HandleWebSocket(w http.ResponseWriter, r *http.Request, session *auth.Session) {
	if session == nil || !session.Active() {
		http.Error(w, "Sign in required", http.StatusUnauthorized)
		return
	}
	// Origin was checked against the configured public URL by RequireSocket.
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, EnableCompression: true, CheckOrigin: func(*http.Request) bool { return true }}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn.EnableWriteCompression(true)

	// Generate a UUID for the client
	clientID := uuid.New().String()

	client := &client{
		conn:    conn,
		send:    make(chan []byte, 256),
		id:      clientID,
		session: session, ready: make(chan struct{}), done: make(chan struct{}),
	}

	wss.register <- client

	go client.writePump(wss)
	go client.readPump(wss.handleMessage, wss.unregister, time.Duration(wss.settings.PongTimeoutSeconds)*time.Second)
	go func() {
		select {
		case <-session.Done:
			conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(message.CloseSessionEnded, "Session ended"), time.Now().Add(time.Second))
			conn.Close()
		case <-client.done:
		}
	}()
}

// Close terminates upgraded connections when the runtime stops serving the game.
func (w *wsServer) Close() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, client := range w.clients {
		client.conn.Close()
	}
}

// PlayerID resolves an account only from the authenticated connection.
func (w *wsServer) PlayerID(clientID string) (model.EntityId, string, bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	c := w.clients[clientID]
	if c == nil || !c.session.Active() {
		return model.EntityId{}, "", false
	}
	return c.session.PlayerID, c.session.Username, true
}

// IsGuest uses only the server-issued session, never client claims.
func (w *wsServer) IsGuest(clientID string) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	c := w.clients[clientID]
	return c != nil && c.session.Guest
}
