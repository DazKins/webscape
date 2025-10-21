package server

import (
	"log"
	"webscape/server/command"
	"webscape/server/game"
	"webscape/server/game/entity"
	"webscape/server/message"

	"github.com/google/uuid"
)

type MessageSender func(clientId string, message message.Message)

type ClientCommandHandler struct {
	game *game.Game
}

func NewClientCommandHandler(game *game.Game) *ClientCommandHandler {
	return &ClientCommandHandler{
		game,
	}
}

func (h *ClientCommandHandler) HandleCommand(clientID string, cmd command.Command) {
	log.Printf("Received %s command: %v\n", cmd.Type, cmd.Data)

	switch cmd.Type {
	case command.CommandTypeJoin:
		h.handleJoinCommand(clientID, cmd)
	case command.CommandTypeMove:
		h.handleMoveCommand(clientID, cmd)
	}
}

func (h *ClientCommandHandler) handleJoinCommand(clientID string, cmd command.Command) {
	id := cmd.Data["id"].(string)
	name, ok := cmd.Data["name"].(string)
	if !ok {
		name = "Heboblobus"
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID: %v", err)
		return
	}

	h.game.HandleJoin(clientID, entity.EntityId(uuid), name)
}

func (h *ClientCommandHandler) handleMoveCommand(clientID string, cmd command.Command) {
	x := cmd.Data["x"].(float64)
	y := cmd.Data["y"].(float64)

	h.game.HandleMove(clientID, int(x), int(y))
}
