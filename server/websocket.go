package server

import (
	"log"
	"net/http"
	"sync"
	"webscape/server/message"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
	send chan []byte
	id   string
}

type MessageHandler func(clientID string, message message.Message)
type DisconnectHandler func(clientID string)

type wsServer struct {
	clients      map[string]*client
	broadcast    chan []byte
	register     chan *client
	unregister   chan *client
	mutex        sync.Mutex
	onMessage    MessageHandler
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

func (w *wsServer) SetMessageHandler(handler MessageHandler) {
	w.onMessage = handler
}

func (w *wsServer) SetDisconnectHandler(handler DisconnectHandler) {
	w.onDisconnect = handler
}

// Broadcast sends a message to all connected clients
func (w *wsServer) Broadcast(message message.Message) {
	log.Println("Broadcasting message to", len(w.clients), "clients: ", message.Marshal())
	w.broadcast <- []byte(message.Marshal())
	log.Println("Broadcasted message to", len(w.clients), "clients: ", message.Marshal())
}

// SendToClient sends a message to a specific client identified by clientID
func (w *wsServer) SendToClient(clientID string, message message.Message) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if client, ok := w.clients[clientID]; ok {
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
		case client := <-w.unregister:
			w.mutex.Lock()
			if _, ok := w.clients[client.id]; ok {
				delete(w.clients, client.id)
				close(client.send)
			}
			w.mutex.Unlock()
			go w.onDisconnect(client.id)
		case message := <-w.broadcast:
			w.mutex.Lock()
			log.Println("Broadcasting message to", len(w.clients), "clients: ", string(message))

			for _, client := range w.clients {
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
		unregister <- c
		c.conn.Close()
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if onMessage != nil {
			msg, err := message.Unmarshal(string(msg))
			if err != nil {
				log.Printf("error: %v", err)
			} else {
				onMessage(c.id, msg)
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
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func (wss *wsServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Generate a UUID for the client
	clientID := uuid.New().String()

	client := &client{
		conn: conn,
		send: make(chan []byte, 256),
		id:   clientID,
	}

	wss.register <- client

	go client.writePump()
	go client.readPump(wss.onMessage, wss.unregister)
}
