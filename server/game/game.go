package game

import (
	"log"
	"time"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/gameevent"
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
		EventEmitter:        game,
		LootHandler:         game,
	})
	game.RegisterSystem(&system.CombatSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World:        world,
		EventEmitter: game,
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
	for _, authoredEntity := range g.world.GetEntities() {
		if _, ok := authoredEntity.Components["playerSpawn"]; ok {
			continue
		}
		if conversation := authoredConversationId(authoredEntity); conversation != "" {
			if _, ok := g.world.GetConversation(conversation); !ok {
				log.Printf("entity %q references unknown conversation %q", authoredEntity.Id, conversation)
				continue
			}
		}
		components := entity.CreateAuthoredEntity(authoredEntity)
		g.componentManager.CreateNewEntity(components...)
	}
}

func authoredConversationId(entity world.WorldEntity) string {
	conversation, ok := entity.Components["conversation"].(map[string]any)
	if !ok {
		return ""
	}
	conversationId, _ := conversation["conversationId"].(string)
	return conversationId
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
		g.broadcastMessage(message.NewGameUpdateMessage(
			updatedComponents,
			removedComponents,
			g.availableInteractionsForGameUpdate(updatedComponents, removedComponents),
		))
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

func (g *Game) EmitGameEvent(event gameevent.Event) {
	if event.Count < 1 {
		event.Count = 1
	}
	log.Printf(
		"game event id=%q actor=%s target=%s count=%d metadata=%v",
		event.Id,
		event.ActorEntityId.String(),
		event.TargetEntityId.String(),
		event.Count,
		event.Metadata,
	)

	questLogComponent := g.componentManager.GetEntityComponent(component.ComponentIdQuestLog, event.ActorEntityId)
	if questLogComponent == nil {
		return
	}

	questLog := questLogComponent.(*component.CQuestLog)
	for _, quest := range g.world.GetQuestRegistry().All() {
		if quest.StartEventId == "" || quest.StartEventId != event.Id {
			continue
		}
		if questLog.IsActive(quest.Id) || questLog.IsCompleted(quest.Id) || len(quest.Steps) == 0 {
			continue
		}
		questLog.StartQuest(quest.Id, quest.Steps[0].Id)
	}

	for _, progress := range questLog.GetActiveProgress() {
		quest, ok := g.world.GetQuest(progress.QuestId)
		if !ok || progress.CurrentStepIndex < 0 || progress.CurrentStepIndex >= len(quest.Steps) {
			questLog.CompleteQuest(progress.QuestId)
			continue
		}

		step := quest.Steps[progress.CurrentStepIndex]
		if step.Requirement.EventId != event.Id {
			continue
		}

		nextCount := progress.CurrentCount + event.Count
		if nextCount < step.Requirement.Count {
			questLog.SetProgress(progress.QuestId, progress.CurrentStepIndex, step.Id, nextCount)
			continue
		}

		nextStepIndex := progress.CurrentStepIndex + 1
		if nextStepIndex >= len(quest.Steps) {
			questLog.CompleteQuest(progress.QuestId)
			continue
		}

		nextStep := quest.Steps[nextStepIndex]
		questLog.SetProgress(progress.QuestId, nextStepIndex, nextStep.Id, 0)
	}

	g.componentManager.SetEntityComponent(event.ActorEntityId, questLog)
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
	snapshot := g.serializedComponentsSnapshot()
	g.sendMessage(clientID, message.NewGameUpdateMessage(
		snapshot,
		make(map[component.ComponentId][]model.EntityId),
		g.availableInteractionsForGameUpdate(
			snapshot,
			make(map[component.ComponentId][]model.EntityId),
		),
	))
}

func (g *Game) AddItemToPlayerInventory(playerEntityId model.EntityId, item *model.Item) {
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId)
	if inventory == nil || item == nil {
		return
	}

	inventoryComponent := inventory.(*component.CInventory)
	inventoryComponent.AddItem(item)
	g.componentManager.SetEntityComponent(playerEntityId, inventoryComponent)

	if item.Type != "" {
		g.EmitGameEvent(gameevent.New("collect:item:"+gameevent.NormalizeToken(item.Type), playerEntityId))
	}
	if item.Name != "" {
		g.EmitGameEvent(gameevent.New("collect:name:"+gameevent.NormalizeToken(item.Name), playerEntityId))
	}
}

func (g *Game) serializedComponentsSnapshot() map[component.ComponentId]map[model.EntityId]util.Json {
	result := make(map[component.ComponentId]map[model.EntityId]util.Json)
	for componentId, entities := range g.componentManager.GetAllComponents() {
		for entityId, comp := range entities {
			serializeableComponent, ok := comp.(component.SerializeableComponent)
			if !ok {
				continue
			}
			if result[componentId] == nil {
				result[componentId] = make(map[model.EntityId]util.Json)
			}
			result[componentId][entityId] = serializeableComponent.Serialize()
		}
	}
	return result
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
	optionAllowed := false
	for _, interactionOption := range g.getInteractionOptionsForEntity(entityId) {
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

func (g *Game) LootEntityFor(playerEntityId model.EntityId, targetEntityId model.EntityId) {
	lootable := g.componentManager.GetEntityComponent(component.ComponentIdLootable, targetEntityId)
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId)
	if lootable == nil || inventory == nil {
		return
	}

	lootableComponent := lootable.(*component.CLootable)
	if !lootableComponent.CanLoot() {
		return
	}

	for _, lootItem := range lootableComponent.GetItems() {
		count := lootItem.Count
		if count < 1 {
			count = 1
		}
		for i := 0; i < count; i++ {
			g.AddItemToPlayerInventory(playerEntityId, lootItem.CreateItem())
		}
	}

	if lootableComponent.IsOnce() {
		lootableComponent.SetLooted(true)
		g.componentManager.SetEntityComponent(targetEntityId, lootableComponent)
	}
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
	g.sendConversationNode(playerEntityId, targetEntityId, conversationId, conversation.StartNodeId)
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
		g.sendConversationNode(
			entityId,
			activeConversation.GetTargetEntityId(),
			conversationId,
			option.NextNodeId,
		)
		return
	}
}

func (g *Game) sendConversationNode(
	playerEntityId model.EntityId,
	targetEntityId model.EntityId,
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

	g.sendMessage(clientID, message.NewConversationMessage(conversationId, targetEntityId, *node))
	g.EmitGameEvent(gameevent.New(
		"conversation:node:"+gameevent.NormalizeToken(conversationId)+":"+gameevent.NormalizeToken(node.Id),
		playerEntityId,
	))
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

func (g *Game) getInteractionOptionsForEntity(entityId model.EntityId) []component.InteractionOption {
	options := []component.InteractionOption{}

	if g.componentManager.GetEntityComponent(component.ComponentIdConversation, entityId) != nil {
		options = append(options, component.InteractionOptionTalk)
	}

	lootable := g.componentManager.GetEntityComponent(component.ComponentIdLootable, entityId)
	if lootable != nil && lootable.(*component.CLootable).CanLoot() {
		options = append(options, component.InteractionOptionLoot)
	}

	if g.componentManager.GetEntityComponent(component.ComponentIdPlayer, entityId) == nil &&
		g.componentManager.GetEntityComponent(component.ComponentIdHealth, entityId) != nil &&
		g.componentManager.GetEntityComponent(component.ComponentIdCombatStats, entityId) != nil {
		options = append(options, component.InteractionOptionAttack)
	}

	return options
}

func (g *Game) availableInteractionsForGameUpdate(
	updatedComponents map[component.ComponentId]map[model.EntityId]util.Json,
	removedComponents map[component.ComponentId][]model.EntityId,
) map[model.EntityId][]component.InteractionOption {
	entityIds := make(map[model.EntityId]bool)
	for _, entities := range updatedComponents {
		for entityId := range entities {
			entityIds[entityId] = true
		}
	}
	for _, entities := range removedComponents {
		for _, entityId := range entities {
			entityIds[entityId] = true
		}
	}

	result := make(map[model.EntityId][]component.InteractionOption, len(entityIds))
	for entityId := range entityIds {
		result[entityId] = g.getInteractionOptionsForEntity(entityId)
	}
	return result
}
