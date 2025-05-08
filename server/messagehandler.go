package server

import (
	"log"
	"webscape/server/game"
	"webscape/server/game/entity"
	"webscape/server/message"

	"github.com/google/uuid"
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
		h.handleJoinMessage(clientID, msg)
	case message.MessageTypeMove:
		h.handleMoveMessage(clientID, msg.Data)
	}
}

func (h *ClientMessageHandler) handleJoinMessage(clientID string, msg message.Message) {
	joinData := msg.Data.(map[string]interface{})
	id := joinData["id"].(string)

	uuid, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID: %v", err)
		return
	}

	h.game.HandleJoin(clientID, entity.EntityId(uuid))
}

func (h *ClientMessageHandler) handleMoveMessage(clientID string, data interface{}) {
	moveData := data.(map[string]interface{})

	x := moveData["x"].(float64)
	y := moveData["y"].(float64)

	h.game.HandleMove(clientID, int(x), int(y))
}
