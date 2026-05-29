package server

import (
	"log"
	"webscape/server/command"
	"webscape/server/game"
	"webscape/server/game/component"
	"webscape/server/game/model"
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
	case command.CommandTypeChat:
		h.handleChatCommand(clientID, cmd)
	case command.CommandTypeInteract:
		h.handleInteractCommand(clientID, cmd)
	case command.CommandTypeEquip:
		h.handleEquipCommand(clientID, cmd)
	case command.CommandTypeUnequip:
		h.handleUnequipCommand(clientID, cmd)
	case command.CommandTypeConversationOption:
		h.handleConversationOptionCommand(clientID, cmd)
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

	h.game.HandleJoin(clientID, model.EntityId(uuid), name)
}

func (h *ClientCommandHandler) handleMoveCommand(clientID string, cmd command.Command) {
	x := cmd.Data["x"].(float64)
	y := cmd.Data["y"].(float64)

	h.game.HandleMove(clientID, int(x), int(y))
}

func (h *ClientCommandHandler) handleChatCommand(clientID string, cmd command.Command) {
	message := cmd.Data["message"].(string)

	h.game.HandleChat(clientID, message)
}

func (h *ClientCommandHandler) handleInteractCommand(clientID string, cmd command.Command) {
	entityId := cmd.Data["entityId"].(string)
	option := cmd.Data["option"].(string)

	uuid, err := uuid.Parse(entityId)
	if err != nil {
		log.Printf("Invalid UUID: %v", err)
		return
	}

	h.game.HandleInteract(clientID, model.EntityId(uuid), component.InteractionOption(option))
}

func (h *ClientCommandHandler) handleEquipCommand(clientID string, cmd command.Command) {
	itemId := cmd.Data["itemId"].(string)
	uuidValue, err := uuid.Parse(itemId)
	if err != nil {
		log.Printf("Invalid item UUID: %v", err)
		return
	}

	h.game.HandleEquip(clientID, model.ItemId(uuidValue))
}

func (h *ClientCommandHandler) handleUnequipCommand(clientID string, cmd command.Command) {
	slotValue, ok := cmd.Data["slot"].(string)
	if !ok {
		log.Printf("Invalid equipment slot")
		return
	}

	slot, valid := model.ParseEquipmentSlot(slotValue)
	if !valid {
		log.Printf("Invalid equipment slot: %s", slotValue)
		return
	}

	h.game.HandleUnequip(clientID, slot)
}

func (h *ClientCommandHandler) handleConversationOptionCommand(clientID string, cmd command.Command) {
	conversationId, ok := cmd.Data["conversationId"].(string)
	if !ok {
		log.Printf("Invalid conversation id")
		return
	}
	nodeId, ok := cmd.Data["nodeId"].(string)
	if !ok {
		log.Printf("Invalid conversation node id")
		return
	}
	optionId, ok := cmd.Data["optionId"].(string)
	if !ok {
		log.Printf("Invalid conversation option id")
		return
	}

	h.game.HandleConversationOption(clientID, conversationId, nodeId, optionId)
}
