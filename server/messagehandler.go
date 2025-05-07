package server

import (
	"log"
	"webscape/server/game"
	"webscape/server/message"
)

type MessageSender func(clientId string, message message.Message)

type ClientMessageHandler struct {
	game *game.Game
}

func NewClientMessageHandler(game *game.Game) *ClientMessageHandler {
	return &ClientMessageHandler{
		game: game,
	}
}

func (h *ClientMessageHandler) HandleMessage(clientID string, msg message.Message) {
	log.Printf("Received %s message: %v\n", msg.Metadata.Type, msg.Data)

	switch msg.Metadata.Type {
	case message.MessageTypeJoin:
		h.handleJoinMessage(clientID)
	case message.MessageTypeMove:
		h.handleMoveMessage(clientID, msg.Data)
	}
}

func (h *ClientMessageHandler) handleJoinMessage(clientID string) {
	h.game.HandleNewJoin(clientID)
}

func (h *ClientMessageHandler) handleMoveMessage(clientID string, data interface{}) {
	moveData := data.(map[string]interface{})

	x := moveData["x"].(float64)
	y := moveData["y"].(float64)

	h.game.HandleMove(clientID, int(x), int(y))
}
