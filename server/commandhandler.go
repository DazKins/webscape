package server

import (
	"log"
	"math"
	"webscape/server/command"
	"webscape/server/game"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/message"

	"github.com/google/uuid"
)

type MessageSender func(clientId string, message message.Message)

type ClientCommandHandler struct {
	game     *game.Game
	playerID func(string) (model.EntityId, bool)
}

func NewClientCommandHandler(game *game.Game, playerID func(string) (model.EntityId, bool)) *ClientCommandHandler {
	return &ClientCommandHandler{
		game: game, playerID: playerID,
	}
}

func (h *ClientCommandHandler) HandleCommand(clientID string, cmd command.Command) {
	log.Printf("Received %s command", cmd.Type)

	if cmd.Type == command.CommandTypeRegister {
		h.handleRegisterCommand(clientID, cmd)
		return
	}
	if !h.game.IsRegistered(clientID) {
		return
	}

	switch cmd.Type {
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
	case command.CommandTypeTrade, command.CommandTypeTradeClose:
		h.handleTradeCommand(clientID, cmd)
	case command.CommandTypeDrop:
		h.handleDropCommand(clientID, cmd)
	case command.CommandTypeConversationOption:
		h.handleConversationOptionCommand(clientID, cmd)
	}
}

func (h *ClientCommandHandler) handleRegisterCommand(clientID string, cmd command.Command) {
	id, ok := h.playerID(clientID)
	if !ok {
		h.game.RejectRegistration(clientID, "sign in required")
		return
	}
	if _, supplied := cmd.Data["id"]; supplied {
		h.game.RejectRegistration(clientID, "player id must not be supplied")
		return
	}
	name, ok := cmd.Data["name"].(string)
	if !ok {
		h.game.RejectRegistration(clientID, "name is required")
		return
	}

	h.game.HandleRegister(clientID, id, name)
}

func (h *ClientCommandHandler) handleMoveCommand(clientID string, cmd command.Command) {
	x, xok := cmd.Data["x"].(float64)
	y, yok := cmd.Data["y"].(float64)
	if !xok || !yok || math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x != math.Trunc(x) || y != math.Trunc(y) || x < math.MinInt32 || x > math.MaxInt32 || y < math.MinInt32 || y > math.MaxInt32 {
		return
	}

	h.game.HandleMove(clientID, int(x), int(y))
}

func (h *ClientCommandHandler) handleChatCommand(clientID string, cmd command.Command) {
	message, ok := cmd.Data["message"].(string)
	if !ok {
		return
	}

	h.game.HandleChat(clientID, message)
}

func (h *ClientCommandHandler) handleInteractCommand(clientID string, cmd command.Command) {
	entityId, idOK := cmd.Data["entityId"].(string)
	option, optionOK := cmd.Data["option"].(string)
	if !idOK || !optionOK {
		return
	}

	uuid, err := uuid.Parse(entityId)
	if err != nil {
		log.Printf("Invalid UUID: %v", err)
		return
	}

	h.game.HandleInteract(clientID, model.EntityId(uuid), component.InteractionOption(option))
}

func (h *ClientCommandHandler) handleDropCommand(clientID string, cmd command.Command) {
	itemID, ok := cmd.Data["itemId"].(string)
	if !ok {
		return
	}
	id, err := uuid.Parse(itemID)
	if err != nil {
		return
	}
	h.game.HandleDrop(clientID, model.ItemId(id))
}

func (h *ClientCommandHandler) handleEquipCommand(clientID string, cmd command.Command) {
	itemId, ok := cmd.Data["itemId"].(string)
	if !ok {
		return
	}
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

func (h *ClientCommandHandler) handleTradeCommand(clientID string, cmd command.Command) {
	target, ok := cmd.Data["targetEntityId"].(string)
	if !ok {
		return
	}
	id, err := uuid.Parse(target)
	if err != nil {
		return
	}
	if cmd.Type == command.CommandTypeTradeClose {
		h.game.HandleTradeClose(clientID, model.EntityId(id))
		return
	}
	action, ok := cmd.Data["action"].(string)
	if !ok || (action != "buy" && action != "sell") {
		return
	}
	itemID, ok := cmd.Data["itemId"].(string)
	if !ok || itemID == "" {
		return
	}
	h.game.HandleTrade(clientID, model.EntityId(id), action, itemID)
}
