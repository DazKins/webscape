package game

import (
	"log"
	"time"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/model"
	"webscape/server/game/system"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

type MessageBroadcaster func(message message.Message)
type MessageSender func(clientID string, message message.Message)

type Game struct {
	world              *world.World
	clientIdToEntityId *util.BiMap[string, model.EntityId]
	ticker             *time.Ticker
	done               chan bool
	sendMessage        MessageSender
	broadcastMessage   MessageBroadcaster

	systems []system.System

	componentManager         *component.ComponentManager
	prevSerialisedComponents map[component.ComponentId]map[model.EntityId]util.Json
}

func NewGameWithWorld(world *world.World) *Game {
	game := &Game{
		clientIdToEntityId: util.NewBiMap[string, model.EntityId](),
		world:              world,
		done:               make(chan bool),

		componentManager: component.NewComponentManager(),
		systems:          []system.System{},

		prevSerialisedComponents: make(map[component.ComponentId]map[model.EntityId]util.Json),
	}

	game.loadWorldEntities()

	game.RegisterSystem(&system.PathingSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world,
	})
	game.RegisterSystem(&system.InteractionSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		ConversationStarter: game,
	})
	game.RegisterSystem(&system.CombatSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world,
	})
	game.RegisterSystem(&system.RandomWalkSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world,
	})
	game.RegisterSystem(&system.TtlSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	})
	game.RegisterSystem(&system.HealthSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	})

	return game
}

func (g *Game) loadWorldEntities() {
	for _, object := range g.world.GetObjects() {
		components := entity.CreateWorldObjectEntity(object)
		g.componentManager.CreateNewEntity(components...)
	}

	for _, spawn := range g.world.GetSpawns() {
		if spawn.Type == "player" {
			continue
		}
		name := spawn.Name
		if name == "" {
			name = spawn.EntityType
		}
		if name == "" {
			name = spawn.Type
		}
		components := entity.CreateDudeEntity(
			name,
			math.Vec2{X: spawn.X, Y: spawn.Y},
		)
		if spawn.ConversationId != "" {
			if _, ok := g.world.GetConversation(spawn.ConversationId); ok {
				components = append(components, component.NewCConversation(spawn.ConversationId))
				for _, comp := range components {
					interactable, ok := comp.(*component.CInteractable)
					if ok {
						interactable.SetInteractionOptions([]component.InteractionOption{
							component.InteractionOptionTalk,
							component.InteractionOptionTrade,
							component.InteractionOptionAttack,
						})
						break
					}
				}
			} else {
				log.Printf("spawn references unknown conversation %q", spawn.ConversationId)
			}
		}
		g.componentManager.CreateNewEntity(components...)
	}
}

func (g *Game) RegisterSystem(system system.System) {
	g.systems = append(g.systems, system)
}

func (g *Game) GetEntitiesWithComponents(componentIds ...component.ComponentId) []model.EntityId {
	if len(componentIds) == 0 {
		return []model.EntityId{}
	}

	// Count how many of the requested components each entity has
	entityCounts := make(map[model.EntityId]int)
	for _, componentId := range componentIds {
		for entityId := range g.componentManager.GetComponent(componentId) {
			entityCounts[entityId]++
		}
	}

	// Only keep entities that have all requested components
	result := make([]model.EntityId, 0)
	requiredCount := len(componentIds)
	for entityId, count := range entityCounts {
		if count == requiredCount {
			result = append(result, entityId)
		}
	}

	return result
}

func (g *Game) StartUpdateLoop() {
	g.ticker = time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.update()
			case <-g.done:
				g.ticker.Stop()
				return
			}
		}
	}()
}

func (g *Game) update() {
	for _, system := range g.systems {
		system.Update()
	}

	updatedComponents := make(map[component.ComponentId]map[model.EntityId]util.Json)

	for componentId, entities := range g.componentManager.GetAllComponents() {
		prevSerialisedEntities := g.prevSerialisedComponents[componentId]
		if prevSerialisedEntities == nil {
			prevSerialisedEntities = make(map[model.EntityId]util.Json)
			g.prevSerialisedComponents[componentId] = prevSerialisedEntities
		}

		for entityId, comp := range entities {
			serializeableComponent, ok := comp.(component.SerializeableComponent)
			if !ok {
				continue
			}

			serialized := serializeableComponent.Serialize()

			if util.JsonEqual(prevSerialisedEntities[entityId], serialized) {
				continue
			}

			prevSerialisedEntities[entityId] = serialized

			if updatedComponents[componentId] == nil {
				updatedComponents[componentId] = make(map[model.EntityId]util.Json)
			}
			updatedComponents[componentId][entityId] = serialized
		}
	}

	removedComponents := make(map[component.ComponentId][]model.EntityId)
	removedEntities := make(map[model.EntityId]bool)

	for componentId, prevSerialisedEntities := range g.prevSerialisedComponents {
		for entityId := range prevSerialisedEntities {
			if g.componentManager.GetEntityComponent(componentId, entityId) == nil {
				removedComponents[componentId] = append(removedComponents[componentId], entityId)
				delete(prevSerialisedEntities, entityId)
				removedEntities[entityId] = true
			}
		}
	}

	// Check if any entities have been completely removed (no components left)
	completelyRemovedEntities := make([]model.EntityId, 0)
	for entityId := range removedEntities {
		// Check if entity has any components left
		hasComponents := false
		for _, components := range g.componentManager.GetAllComponents() {
			if _, exists := components[entityId]; exists {
				hasComponents = true
				break
			}
		}
		if !hasComponents {
			completelyRemovedEntities = append(completelyRemovedEntities, entityId)
		}
	}

	if len(updatedComponents) > 0 || len(removedComponents) > 0 {
		g.broadcastMessage(message.NewGameUpdateMessage(updatedComponents, removedComponents))
	}

	// Send entity removal messages for completely removed entities
	for _, entityId := range completelyRemovedEntities {
		g.broadcastMessage(message.NewEntityRemoveMessage(entityId))
	}
}

func (g *Game) Stop() {
	g.done <- true
}

func (g *Game) RegisterBroadcaster(messageBroadcaster MessageBroadcaster) {
	g.broadcastMessage = messageBroadcaster
}

func (g *Game) RegisterSender(messageSender MessageSender) {
	g.sendMessage = messageSender
}

func (g *Game) HandleJoin(clientID string, id model.EntityId, name string) {
	if _, ok := g.clientIdToEntityId.Get(clientID); ok {
		g.sendMessage(clientID, message.NewJoinFailedMessage("you are already connected in another session"))
		return
	}

	components := entity.CreatePlayerEntity(id, name, g.world.GetPlayerSpawn())
	g.componentManager.SetEntityComponents(id, components...)

	g.clientIdToEntityId.Put(clientID, id)

	// Serialize the newly created player's components and add them to prevSerialisedComponents
	// so they're included in the initial game update
	for _, comp := range components {
		serializeableComponent, ok := comp.(component.SerializeableComponent)
		if !ok {
			continue
		}

		componentId := comp.GetId()
		if g.prevSerialisedComponents[componentId] == nil {
			g.prevSerialisedComponents[componentId] = make(map[model.EntityId]util.Json)
		}

		serialized := serializeableComponent.Serialize()
		g.prevSerialisedComponents[componentId][id] = serialized
	}

	g.sendMessage(clientID, message.NewJoinedMessage(id.String()))
	g.sendMessage(clientID, message.NewWorldMessage(g.world))

	// This seems hacky. Could be race conditions if we try
	// to send the new client the state of all entities while
	// a game tick is in progress. Works for now.
	g.sendMessage(clientID, message.NewGameUpdateMessage(
		g.prevSerialisedComponents,
		make(map[component.ComponentId][]model.EntityId),
	))
}

func (g *Game) HandleMove(clientID string, x int, y int) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		panic("client ID not found in clientIdToEntityId")
	}

	positionComponent := g.componentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
	if !ok || positionComponent == nil {
		panic("position component not found")
	}

	pathingComponent := component.NewCPathing(component.PathingTarget{
		Position: util.OptionalSome(math.Vec2{X: x, Y: y}),
	})
	g.componentManager.SetEntityComponent(entityId, pathingComponent)
}

func (g *Game) HandleLeave(clientID string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.componentManager.RemoveEntity(entityId)
	g.clientIdToEntityId.Delete(clientID)
}

// SendChatMessageEntityFor sends a new chat message entity for the given entity,
// removing any existing chat messages from that entity first to ensure only one
// chat message exists per entity at a time.
func (g *Game) SendChatMessageEntityFor(fromEntityId model.EntityId, message string) model.EntityId {
	// Remove any existing chat messages from this entity
	chatMessageEntities := g.componentManager.GetComponent(component.ComponentIdChatMessage)
	for existingEntityId, comp := range chatMessageEntities {
		chatMessageComp := comp.(*component.CChatMessage)
		if chatMessageComp.GetFromEntityId() == fromEntityId {
			g.componentManager.RemoveEntity(existingEntityId)
		}
	}

	// Create the new chat message entity
	chatMessageComponents := entity.CreateChatMessageEntity(fromEntityId, message)
	return g.componentManager.CreateNewEntity(chatMessageComponents...)
}

func (g *Game) HandleChat(clientID string, chatMessage string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.SendChatMessageEntityFor(entityId, chatMessage)
}

func (g *Game) HandleInteract(clientID string, entityId model.EntityId, option component.InteractionOption) {
	interactable := g.componentManager.GetEntityComponent(component.ComponentIdInteractable, entityId)
	if interactable == nil {
		return
	}

	interactableComponent := interactable.(*component.CInteractable)
	optionAllowed := false
	for _, interactionOption := range interactableComponent.GetInteractionOptions() {
		if interactionOption == option {
			optionAllowed = true
			break
		}
	}
	if !optionAllowed {
		return
	}
	if option == component.InteractionOptionTalk &&
		g.componentManager.GetEntityComponent(component.ComponentIdConversation, entityId) == nil {
		return
	}

	interactingEntityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	// Set pathing component to path to the target entity
	pathingComponent := component.NewCPathing(component.PathingTarget{
		EntityId: util.OptionalSome(entityId),
	})
	g.componentManager.SetEntityComponent(interactingEntityId, pathingComponent)

	// Set interacting component to track the interaction
	interactingComponent := component.NewCInteracting(entityId, option)
	g.componentManager.SetEntityComponent(interactingEntityId, interactingComponent)
}

func (g *Game) StartConversationFor(playerEntityId model.EntityId, targetEntityId model.EntityId) {
	conversationComponent := g.componentManager.GetEntityComponent(
		component.ComponentIdConversation,
		targetEntityId,
	)
	if conversationComponent == nil {
		return
	}

	conversationId := conversationComponent.(*component.CConversation).GetConversationId()
	conversation, ok := g.world.GetConversation(conversationId)
	if !ok {
		return
	}

	activeConversation := component.NewCActiveConversation(
		conversationId,
		targetEntityId,
		conversation.StartNodeId,
	)
	g.componentManager.SetEntityComponent(playerEntityId, activeConversation)
	g.sendConversationNode(playerEntityId, conversationId, conversation.StartNodeId)
}

func (g *Game) HandleConversationOption(
	clientID string,
	conversationId string,
	nodeId string,
	optionId string,
) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	active := g.componentManager.GetEntityComponent(component.ComponentIdActiveConversation, entityId)
	if active == nil {
		return
	}

	activeConversation := active.(*component.CActiveConversation)
	if activeConversation.GetConversationId() != conversationId ||
		activeConversation.GetCurrentNodeId() != nodeId {
		return
	}

	conversation, ok := g.world.GetConversation(conversationId)
	if !ok {
		g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, entityId)
		return
	}

	node, ok := conversation.GetNode(nodeId)
	if !ok {
		g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, entityId)
		return
	}

	for _, option := range node.Options {
		if option.Id != optionId {
			continue
		}

		activeConversation.SetCurrentNodeId(option.NextNodeId)
		g.componentManager.SetEntityComponent(entityId, activeConversation)
		g.sendConversationNode(entityId, conversationId, option.NextNodeId)
		return
	}
}

func (g *Game) sendConversationNode(
	playerEntityId model.EntityId,
	conversationId string,
	nodeId string,
) {
	conversation, ok := g.world.GetConversation(conversationId)
	if !ok {
		g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, playerEntityId)
		return
	}

	node, ok := conversation.GetNode(nodeId)
	if !ok {
		g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, playerEntityId)
		return
	}

	clientID, ok := g.clientIdToEntityId.GetKey(playerEntityId)
	if !ok {
		return
	}

	g.sendMessage(clientID, message.NewConversationMessage(conversationId, *node))
	if node.EndConversation {
		g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, playerEntityId)
	}
}

func (g *Game) HandleEquip(clientID string, itemId model.ItemId) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, entityId)
	equipped := g.componentManager.GetEntityComponent(component.ComponentIdEquipped, entityId)
	baseStats := g.componentManager.GetEntityComponent(component.ComponentIdBaseStats, entityId)
	if inventory == nil || equipped == nil || baseStats == nil {
		return
	}

	inventoryComponent := inventory.(*component.CInventory)
	equippedComponent := equipped.(*component.CEquipped)
	baseStatsComponent := baseStats.(*component.CBaseStats)

	item := inventoryComponent.GetItem(itemId)
	if item == nil || !item.IsEquipable() || item.GetEquipmentSlot() == nil {
		return
	}

	slot := *item.GetEquipmentSlot()
	previousItem := equippedComponent.EquipItem(slot, item)
	inventoryComponent.RemoveItem(itemId)
	if previousItem != nil {
		inventoryComponent.AddItem(previousItem)
	}

	g.componentManager.SetEntityComponent(entityId, inventoryComponent)
	g.componentManager.SetEntityComponent(entityId, equippedComponent)

	combatStats := component.CalculateCombatStats(baseStatsComponent, equippedComponent)
	g.componentManager.SetEntityComponent(entityId, combatStats)
}

func (g *Game) HandleUnequip(clientID string, slot model.EquipmentSlot) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, entityId)
	equipped := g.componentManager.GetEntityComponent(component.ComponentIdEquipped, entityId)
	baseStats := g.componentManager.GetEntityComponent(component.ComponentIdBaseStats, entityId)
	if inventory == nil || equipped == nil || baseStats == nil {
		return
	}

	inventoryComponent := inventory.(*component.CInventory)
	equippedComponent := equipped.(*component.CEquipped)
	baseStatsComponent := baseStats.(*component.CBaseStats)

	item := equippedComponent.UnequipItem(slot)
	if item != nil {
		inventoryComponent.AddItem(item)
	}

	g.componentManager.SetEntityComponent(entityId, inventoryComponent)
	g.componentManager.SetEntityComponent(entityId, equippedComponent)

	combatStats := component.CalculateCombatStats(baseStatsComponent, equippedComponent)
	g.componentManager.SetEntityComponent(entityId, combatStats)
}
