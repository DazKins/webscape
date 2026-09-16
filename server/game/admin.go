package game

import (
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
	if command != "/reset" {
		reply(false, "Unknown admin command or invalid arguments. Usage: /reset")
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
	g.componentManager.SetEntityComponents(id, entity.CreatePlayerEntity(id, player.GetName(), g.world.GetPlayerSpawn(), g.currentTick)...)
	reply(true, "Character reset to a new player at spawn.")
}
