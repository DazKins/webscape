package game

import (
	"fmt"
	"strconv"
	"strings"
	"webscape/server/config"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/model"
	"webscape/server/game/system"
	"webscape/server/message"
)

func (g *Game) ConfigureAdminCommands(settings config.AdminCommandsConfig, devMode bool) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	g.adminCommands = settings
	g.devMode = devMode
}

// Called under stateMutex, using only the registered connection's player identity.
func (g *Game) handleAdminCommand(clientID string, id model.EntityId, command string) {
	reply := func(success bool, text string) {
		g.pendingClientEvents = append(g.pendingClientEvents, pendingClientEvent{
			clientIDs: []string{clientID}, message: message.NewAdminCommandResult(command, success, text),
		})
	}
	if !g.adminCommands.IsEnabled(g.devMode) {
		reply(false, "Admin commands are disabled.")
		return
	}
	if !g.adminCommands.Allows(id.String(), g.devMode) {
		reply(false, "You do not have permission to use admin commands.")
		return
	}
	args := strings.Fields(command)
	if len(args) > 0 && args[0] == "/give" {
		if len(args) < 2 || len(args) > 3 {
			reply(false, "Usage: /give {item_definition_id} [quantity]")
			return
		}
		quantity := 1
		if len(args) == 3 {
			parsed, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil || parsed < 1 || parsed > model.MaxStackQuantity {
				reply(false, fmt.Sprintf("Quantity must be a whole number from 1 to %d.", model.MaxStackQuantity))
				return
			}
			quantity = int(parsed)
		}
		item := model.NewItem(args[1])
		if item == nil {
			reply(false, fmt.Sprintf("Unknown item definition %q.", args[1]))
			return
		}
		inventory, ok := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
		if !ok {
			reply(false, "Your character has no inventory.")
			return
		}
		if !item.IsStackable() && quantity > inventory.AvailableSlots() {
			reply(false, "Cannot add item: your inventory or item stack is full.")
			return
		}
		// Prepare the complete grant before committing; never grant a partial batch.
		trial := inventory.Clone()
		instances := quantity
		if item.IsStackable() {
			item.Quantity = quantity
			instances = 1
		}
		for i := 0; i < instances; i++ {
			if i > 0 {
				item = model.NewItem(args[1])
			}
			if !trial.AddItem(item) {
				reply(false, "Cannot add item: your inventory or item stack is full.")
				return
			}
		}
		*inventory = *trial
		if quantity > 1 {
			reply(true, fmt.Sprintf("Added %d x %s to your inventory.", quantity, item.Name()))
			return
		}
		reply(true, fmt.Sprintf("Added %s to your inventory.", item.Name()))
		return
	}
	if command != "/reset" {
		reply(false, "Unknown admin command or invalid arguments. Usage: /reset or /give {item_definition_id} [quantity]")
		return
	}
	player := g.componentManager.GetEntityComponent(component.ComponentIdPlayer, id).(*component.CPlayer)
	for _, registeredSystem := range g.systems {
		if combat, ok := registeredSystem.(*system.CombatSystem); ok {
			combat.ResetEntity(id)
		}
	}
	// Remove every component, including transient actions and future additions,
	// before applying exactly the same defaults as a first registration.
	g.componentManager.RemoveEntity(id)
	delete(g.offlinePlayers, id)
	g.componentManager.SetEntityComponents(id, entity.CreatePlayerEntity(id, player.GetName(), g.world.GetPlayerSpawn(), g.currentTick, g.world.GetManaSettings().MaxMana)...)
	reply(true, "Character reset to a new player at spawn.")
}
