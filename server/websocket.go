package server

import (
	"log"
	"net/http"
	"sync"
	"time"
	"webscape/server/auth"
	"webscape/server/game/model"
	"webscape/server/message"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type client struct {
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
	clients      map[string]*client
	broadcast    chan []byte
	register     chan *client
	unregister   chan *client
	mutex        sync.Mutex
	onMessage    MessageHandler
	onConnect    ConnectHandler
	onDisconnect DisconnectHandler
}

func NewWsServer() *wsServer {
	wss := &wsServer{
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
			close(client.send)
			delete(w.clients, clientID)
		}
	}
}

func (w *wsServer) run() {
	for {
		select {
		case client := <-w.register:
			w.mutex.Lock()
			w.clients[client.id] = client
			w.mutex.Unlock()
			client.session.WithActive(func() {
				if w.onConnect != nil {
					w.onConnect(client.id)
				}
			})
			close(client.ready)
		case client := <-w.unregister:
			w.mutex.Lock()
			if _, ok := w.clients[client.id]; ok {
				delete(w.clients, client.id)
				close(client.send)
			}
			w.mutex.Unlock()
			if w.onDisconnect != nil {
				w.onDisconnect(client.id)
			}
		case message := <-w.broadcast:
			w.mutex.Lock()
			for _, client := range w.clients {
				if !client.session.Active() {
					continue
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(w.clients, client.id)
				}
			}
			w.mutex.Unlock()
		}
	}
}

func (c *client) readPump(onMessage MessageHandler, unregister chan *client) {
	defer func() {
		close(c.done)
		unregister <- c
		c.conn.Close()
	}()

	<-c.ready
	c.conn.SetReadLimit(16 * 1024)
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if onMessage != nil {
			if !c.session.WithActive(func() { onMessage(c.id, string(msg)) }) {
				return
			}
		}
	}
}

func (c *client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok || !c.session.Active() {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
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

	go client.writePump()
	go client.readPump(wss.onMessage, wss.unregister)
	go func() {
		select {
		case <-session.Done:
			conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(4001, "Session ended"), time.Now().Add(time.Second))
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
func (w *wsServer) PlayerID(clientID string) (model.EntityId, bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	c := w.clients[clientID]
	if c == nil || !c.session.Active() {
		return model.EntityId{}, false
	}
	return c.session.PlayerID, true
}
