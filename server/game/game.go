package game

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"webscape/server/game/collision"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/spatial"
	"webscape/server/game/system"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

type MessageBroadcaster func(message message.Message)
type MessageSender func(clientID string, message message.Message)

type Game struct {
	// stateMutex serializes update ticks with commands arriving from WebSocket goroutines.
	stateMutex          sync.Mutex
	world               *world.World
	clientIdToEntityId  *util.BiMap[string, model.EntityId]
	ticker              *time.Ticker
	done                chan bool
	sendMessage         MessageSender
	broadcastMessage    MessageBroadcaster
	chunkRadius         int
	clients             map[string]*clientStreamState
	spatialIndex        *spatial.Index
	eventDispatcher     *gameevent.Dispatcher
	pendingClientEvents []pendingClientEvent

	systems           []system.System
	woodcuttingSystem *system.WoodcuttingSystem

	componentManager *component.ComponentManager
}

type clientStreamState struct {
	loadedChunks map[world.ChunkCoord]bool
	baseline     map[component.ComponentId]map[model.EntityId]util.Json
}

type pendingClientEvent struct {
	clientIDs []string
	message   message.Message
}

func NewGameWithWorld(world *world.World) *Game {
	return NewGameWithWorldAndChunkRadius(world, 1)
}

func NewGameWithWorldAndChunkRadius(world *world.World, chunkRadius int) *Game {
	if chunkRadius < 0 {
		chunkRadius = 0
	}
	game := &Game{
		clientIdToEntityId: util.NewBiMap[string, model.EntityId](),
		world:              world,
		done:               make(chan bool),

		componentManager: component.NewComponentManager(),
		systems:          []system.System{},
		chunkRadius:      chunkRadius,
		clients:          make(map[string]*clientStreamState),
		eventDispatcher:  gameevent.NewDispatcher(),
	}
	questHandler := gameevent.HandlerFunc(game.handleQuestEvent)
	for eventId := range game.questEventIds() {
		game.RegisterGameEventHandlerFor(eventId, questHandler)
	}
	clientHandler := gameevent.HandlerFunc(game.handleClientEvent)
	game.RegisterGameEventHandlerFor(gameevent.EventIdChatSpoken, clientHandler)
	game.RegisterGameEventHandlerFor(gameevent.EventIdCombatResolved, clientHandler)
	game.RegisterGameEventHandlerFor(gameevent.EventIdWoodcuttingSwing, clientHandler)

	game.loadWorldEntities()
	game.spatialIndex = spatial.NewIndex(world, game.componentManager)

	game.RegisterSystem(&system.PathingSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world, SpatialIndex: game.spatialIndex, EventEmitter: game,
	})
	woodcuttingSystem := &system.WoodcuttingSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		YieldHandler: game,
		EventEmitter: game,
	}
	game.woodcuttingSystem = woodcuttingSystem
	game.RegisterSystem(&system.InteractionSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		ConversationStarter: game,
		EventEmitter:        game,
		LootHandler:         game,
		WoodcuttingStarter:  woodcuttingSystem,
	})
	game.RegisterSystem(woodcuttingSystem)
	game.RegisterSystem(&system.CombatSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world, SpatialIndex: game.spatialIndex,
		EventEmitter: game,
	})
	game.RegisterSystem(&system.RandomWalkSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world, SpatialIndex: game.spatialIndex,
	})
	game.RegisterSystem(&system.HealthSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	})
	spawnSystem := &system.SpawnSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	}
	game.RegisterSystem(spawnSystem)
	game.RegisterSystem(&system.FacingSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	})
	spawnSystem.Update()

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
		if spawnChildConversation := authoredSpawnChildConversationId(authoredEntity); spawnChildConversation != "" {
			if _, ok := g.world.GetConversation(spawnChildConversation); !ok {
				log.Printf("spawn entity %q references unknown conversation %q", authoredEntity.Id, spawnChildConversation)
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

func authoredSpawnChildConversationId(entity world.WorldEntity) string {
	spawn, ok := entity.Components["spawn"].(map[string]any)
	if !ok {
		return ""
	}
	rawTemplate, ok := spawn["entity"].(map[string]any)
	if !ok {
		return ""
	}
	components, ok := rawTemplate["components"].(map[string]any)
	if !ok {
		return ""
	}
	return authoredConversationId(world.WorldEntity{Components: components})
}

func (g *Game) RegisterSystem(system system.System) {
	g.systems = append(g.systems, system)
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
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	for _, system := range g.systems {
		system.Update()
	}

	for clientID := range g.clients {
		g.syncClient(clientID)
	}
	g.flushPendingClientEvents()
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

func questCompletedEventId(questId string) string {
	return "quest:completed:" + gameevent.NormalizeToken(questId)
}

func (g *Game) RegisterGameEventHandler(handler gameevent.Handler) {
	g.eventDispatcher.Register(handler)
}

func (g *Game) RegisterGameEventHandlerFor(eventId string, handler gameevent.Handler) {
	g.eventDispatcher.Subscribe(eventId, handler)
}

func (g *Game) EmitGameEvent(event gameevent.Event) {
	if event.Count < 1 {
		event.Count = 1
	}
	if !isHighFrequencyClientEvent(event.Id) {
		log.Printf(
			"game event id=%q actor=%s target=%s count=%d metadata=%v",
			event.Id,
			event.ActorEntityId.String(),
			event.TargetEntityId.String(),
			event.Count,
			event.Metadata,
		)
	}
	g.eventDispatcher.Emit(event)
}

func isHighFrequencyClientEvent(eventId string) bool {
	switch eventId {
	case gameevent.EventIdChatSpoken,
		gameevent.EventIdCombatResolved,
		gameevent.EventIdWoodcuttingSwing:
		return true
	default:
		return false
	}
}

func (g *Game) questEventIds() map[string]bool {
	result := map[string]bool{}
	for _, quest := range g.world.GetQuestRegistry().All() {
		if quest.StartEventId != "" {
			result[quest.StartEventId] = true
		}
		for _, step := range quest.Steps {
			if step.Requirement.EventId != "" {
				result[step.Requirement.EventId] = true
			}
		}
	}
	return result
}

func (g *Game) handleQuestEvent(event gameevent.Event) {
	questLogComponent := g.componentManager.GetEntityComponent(component.ComponentIdQuestLog, event.ActorEntityId)
	if questLogComponent == nil {
		return
	}

	questLog := questLogComponent.(*component.CQuestLog)
	questLogChanged := false
	completedQuestEvents := []gameevent.Event{}
	completeQuest := func(quest world.Quest, completedStep world.QuestStep) {
		completedEvent, ok := g.completeQuestForPlayer(event.ActorEntityId, event.TargetEntityId, questLog, quest, completedStep)
		if !ok {
			return
		}
		questLogChanged = true
		completedQuestEvents = append(completedQuestEvents, completedEvent)
	}

	for _, quest := range g.world.GetQuestRegistry().All() {
		if quest.StartEventId == "" || quest.StartEventId != event.Id {
			continue
		}
		if questLog.IsActive(quest.Id) || questLog.IsCompleted(quest.Id) || len(quest.Steps) == 0 {
			continue
		}
		questLog.StartQuest(quest.Id, quest.Steps[0].Id)
		questLogChanged = true
	}

	for _, progress := range questLog.GetActiveProgress() {
		quest, ok := g.world.GetQuest(progress.QuestId)
		if !ok {
			questLog.CompleteQuest(progress.QuestId)
			questLogChanged = true
			continue
		}
		if progress.CurrentStepIndex < 0 || progress.CurrentStepIndex >= len(quest.Steps) {
			completeQuest(*quest, world.QuestStep{})
			continue
		}

		step := quest.Steps[progress.CurrentStepIndex]
		if step.Requirement.EventId != event.Id {
			continue
		}

		nextCount := progress.CurrentCount + event.Count
		if nextCount < step.Requirement.Count {
			questLog.SetProgress(progress.QuestId, progress.CurrentStepIndex, step.Id, nextCount)
			questLogChanged = true
			continue
		}

		nextStepIndex := progress.CurrentStepIndex + 1
		if nextStepIndex >= len(quest.Steps) {
			completeQuest(*quest, step)
			continue
		}

		nextStep := quest.Steps[nextStepIndex]
		questLog.SetProgress(progress.QuestId, nextStepIndex, nextStep.Id, 0)
		questLogChanged = true
	}

	if questLogChanged {
		g.componentManager.SetEntityComponent(event.ActorEntityId, questLog)
	}
	for _, completedEvent := range completedQuestEvents {
		g.EmitGameEvent(completedEvent)
	}
}

func (g *Game) handleClientEvent(event gameevent.Event) {
	var clientMessage message.Message
	entityIDs := []model.EntityId{event.ActorEntityId}

	switch event.Id {
	case gameevent.EventIdChatSpoken:
		payload, ok := event.Payload.(gameevent.ChatSpokenPayload)
		if !ok {
			return
		}
		clientMessage = message.NewChatMessage(event.ActorEntityId, payload.Message)
	case gameevent.EventIdCombatResolved:
		payload, ok := event.Payload.(gameevent.CombatResolvedPayload)
		if !ok {
			return
		}
		entityIDs = append(entityIDs, event.TargetEntityId)
		clientMessage = message.NewCombatResolvedMessage(
			event.ActorEntityId,
			event.TargetEntityId,
			payload.DidHit,
			payload.Damage,
			payload.IsCritical,
		)
	case gameevent.EventIdWoodcuttingSwing:
		if _, ok := event.Payload.(gameevent.WoodcuttingSwingPayload); !ok {
			return
		}
		entityIDs = append(entityIDs, event.TargetEntityId)
		clientMessage = message.NewWoodcuttingSwingMessage(event.ActorEntityId, event.TargetEntityId)
	default:
		return
	}

	clientIDs := g.clientIDsViewingAny(entityIDs...)
	if len(clientIDs) == 0 {
		return
	}
	g.pendingClientEvents = append(g.pendingClientEvents, pendingClientEvent{
		clientIDs: clientIDs,
		message:   clientMessage,
	})
}

func (g *Game) clientIDsViewingAny(entityIDs ...model.EntityId) []string {
	clientIDs := make([]string, 0)
	for clientID := range g.clients {
		for _, entityID := range entityIDs {
			if entityID == (model.EntityId{}) || !g.entityVisibleToClient(clientID, entityID) {
				continue
			}
			clientIDs = append(clientIDs, clientID)
			break
		}
	}
	sort.Strings(clientIDs)
	return clientIDs
}

func (g *Game) flushPendingClientEvents() {
	pending := g.pendingClientEvents
	g.pendingClientEvents = nil
	if g.sendMessage == nil {
		return
	}
	for _, event := range pending {
		for _, clientID := range event.clientIDs {
			if g.clients[clientID] != nil {
				g.sendMessage(clientID, event.message)
			}
		}
	}
}

func (g *Game) completeQuestForPlayer(
	playerEntityId model.EntityId,
	targetEntityId model.EntityId,
	questLog *component.CQuestLog,
	quest world.Quest,
	completedStep world.QuestStep,
) (gameevent.Event, bool) {
	if questLog.IsCompleted(quest.Id) {
		return gameevent.Event{}, false
	}

	questLog.CompleteQuest(quest.Id)
	rewardDeliveries := g.deliverQuestRewards(playerEntityId, quest.Rewards.Items)
	if clientID, ok := g.clientIdToEntityId.GetKey(playerEntityId); ok {
		g.sendMessage(clientID, message.NewQuestCompletedMessage(quest, completedStep, rewardDeliveries))
	}

	completedEvent := gameevent.New(questCompletedEventId(quest.Id), playerEntityId)
	completedEvent.TargetEntityId = targetEntityId
	completedEvent.Metadata["questId"] = quest.Id
	return completedEvent, true
}

func (g *Game) deliverQuestRewards(
	playerEntityId model.EntityId,
	rewardItems []world.QuestRewardItem,
) []message.QuestRewardDelivery {
	droppedItems := []component.LootItem{}
	deliveries := []message.QuestRewardDelivery{}

	for _, rewardItem := range rewardItems {
		count := rewardItem.Count
		if count < 1 {
			count = 1
		}

		addedCount := 0
		droppedCount := 0
		for i := 0; i < count; i++ {
			item := model.NewItem(rewardItem.Name, rewardItem.Type)
			if g.addItemToPlayerInventory(playerEntityId, item, false) {
				addedCount++
				continue
			}
			droppedCount++
		}
		if addedCount > 0 {
			deliveries = append(deliveries, message.QuestRewardDelivery{
				Name:     rewardItem.Name,
				Type:     rewardItem.Type,
				Count:    addedCount,
				Delivery: message.QuestRewardDeliveryInventory,
			})
		}
		if droppedCount > 0 {
			droppedItems = append(droppedItems, component.LootItem{
				Name:  rewardItem.Name,
				Type:  rewardItem.Type,
				Count: droppedCount,
			})
			deliveries = append(deliveries, message.QuestRewardDelivery{
				Name:     rewardItem.Name,
				Type:     rewardItem.Type,
				Count:    droppedCount,
				Delivery: message.QuestRewardDeliveryDropped,
			})
		}
	}

	if len(droppedItems) > 0 {
		g.spawnRewardDrop(playerEntityId, droppedItems)
	}
	return deliveries
}

func (g *Game) HandleRegister(clientID string, id model.EntityId, name string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	if _, ok := g.clientIdToEntityId.Get(clientID); ok {
		g.sendMessage(clientID, message.NewRegistrationFailedMessage("this connection is already registered"))
		return
	}
	if _, ok := g.clientIdToEntityId.GetKey(id); ok {
		g.sendMessage(clientID, message.NewRegistrationFailedMessage("this player is already active"))
		return
	}

	normalizedName := strings.TrimSpace(name)
	nameLength := utf8.RuneCountInString(normalizedName)
	if nameLength == 0 {
		g.sendMessage(clientID, message.NewRegistrationFailedMessage("name cannot be blank"))
		return
	}
	if nameLength > 24 {
		g.sendMessage(clientID, message.NewRegistrationFailedMessage("name must be 24 characters or fewer"))
		return
	}

	components := entity.CreatePlayerEntity(id, normalizedName, g.world.GetPlayerSpawn())
	g.componentManager.SetEntityComponents(id, components...)

	g.clientIdToEntityId.Put(clientID, id)
	state := g.clients[clientID]
	hadViewerStream := state != nil
	if state == nil {
		state = newClientStreamState()
		g.clients[clientID] = state
	}
	// Re-send the visible snapshot after admission so private player components
	// and gameplay interactions replace the observer-only view.
	state.baseline = make(map[component.ComponentId]map[model.EntityId]util.Json)

	g.sendMessage(clientID, message.NewRegisteredMessage(id.String(), normalizedName))
	if !hadViewerStream {
		g.sendMessage(clientID, message.NewWorldMessage(g.world))
	}
	g.syncClient(clientID)
}

func newClientStreamState() *clientStreamState {
	return &clientStreamState{
		loadedChunks: make(map[world.ChunkCoord]bool),
		baseline:     make(map[component.ComponentId]map[model.EntityId]util.Json),
	}
}

func (g *Game) HandleConnect(clientID string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	if g.clients[clientID] != nil {
		return
	}
	g.clients[clientID] = newClientStreamState()
	g.sendMessage(clientID, message.NewWorldMessage(g.world))
	g.syncClient(clientID)
}

func (g *Game) RejectRegistration(clientID string, reason string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	g.sendMessage(clientID, message.NewRegistrationFailedMessage(reason))
}

func (g *Game) IsRegistered(clientID string) bool {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	_, ok := g.clientIdToEntityId.Get(clientID)
	return ok
}

func (g *Game) syncClient(clientID string) {
	state := g.clients[clientID]
	if state == nil {
		return
	}
	playerID, registered := g.clientIdToEntityId.Get(clientID)
	position := g.world.GetPlayerSpawn()
	if registered {
		positionComponent := g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerID)
		if positionComponent == nil {
			return
		}
		position = positionComponent.(*component.CPosition).GetPosition()
	}
	center, _ := g.world.GlobalToChunk(position.X, position.Y)
	nextChunks := g.world.ChunksWithin(center, g.chunkRadius)
	loads := make([]world.Chunk, 0)
	unloads := make([]world.ChunkCoord, 0)
	for coord := range nextChunks {
		if state.loadedChunks[coord] {
			continue
		}
		if chunk, exists := g.world.GetChunk(coord); exists {
			loads = append(loads, *chunk)
		}
	}
	for coord := range state.loadedChunks {
		if !nextChunks[coord] {
			unloads = append(unloads, coord)
		}
	}
	sort.Slice(loads, func(i, j int) bool {
		if loads[i].Coordinate.Y == loads[j].Coordinate.Y {
			return loads[i].Coordinate.X < loads[j].Coordinate.X
		}
		return loads[i].Coordinate.Y < loads[j].Coordinate.Y
	})
	sort.Slice(unloads, func(i, j int) bool {
		if unloads[i].Y == unloads[j].Y {
			return unloads[i].X < unloads[j].X
		}
		return unloads[i].Y < unloads[j].Y
	})
	if len(loads) > 0 || len(unloads) > 0 {
		g.sendMessage(clientID, message.NewChunkUpdateMessage(loads, unloads))
	}
	state.loadedChunks = nextChunks

	visible := g.spatialIndex.EntitiesInChunks(nextChunks)
	if registered {
		visible[playerID] = true
	}
	updated := make(map[component.ComponentId]map[model.EntityId]util.Json)
	removed := make(map[component.ComponentId][]model.EntityId)
	for componentID, entities := range g.componentManager.GetAllComponents() {
		if !registered && !isPublicObserverComponent(componentID) {
			continue
		}
		if state.baseline[componentID] == nil {
			state.baseline[componentID] = make(map[model.EntityId]util.Json)
		}
		for entityID, value := range entities {
			if !visible[entityID] {
				continue
			}
			serializable, ok := value.(component.SerializeableComponent)
			if !ok {
				continue
			}
			serialized := serializable.Serialize()
			if util.JsonEqual(state.baseline[componentID][entityID], serialized) {
				continue
			}
			if updated[componentID] == nil {
				updated[componentID] = make(map[model.EntityId]util.Json)
			}
			updated[componentID][entityID] = serialized
			state.baseline[componentID][entityID] = serialized
		}
	}
	for componentID, entities := range state.baseline {
		for entityID := range entities {
			if visible[entityID] && g.componentManager.GetEntityComponent(componentID, entityID) != nil {
				continue
			}
			removed[componentID] = append(removed[componentID], entityID)
			delete(entities, entityID)
		}
	}
	if len(updated) > 0 || len(removed) > 0 {
		interactions := map[model.EntityId][]component.InteractionOption{}
		if registered {
			interactions = g.availableInteractionsForGameUpdate(updated, removed)
		}
		g.sendMessage(clientID, message.NewGameUpdateMessage(updated, removed, interactions))
	}
}

func isPublicObserverComponent(componentID component.ComponentId) bool {
	switch componentID {
	case component.ComponentIdPosition,
		component.ComponentIdMetadata,
		component.ComponentIdRenderable,
		component.ComponentIdHealth,
		component.ComponentIdFacing,
		component.ComponentIdOpenable:
		return true
	default:
		return false
	}
}

func (g *Game) AddItemToPlayerInventory(playerEntityId model.EntityId, item *model.Item) bool {
	return g.addItemToPlayerInventory(playerEntityId, item, true)
}

func (g *Game) addItemToPlayerInventory(playerEntityId model.EntityId, item *model.Item, emitEvents bool) bool {
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId)
	if inventory == nil || item == nil {
		return false
	}

	inventoryComponent := inventory.(*component.CInventory)
	if !inventoryComponent.AddItem(item) {
		return false
	}
	g.componentManager.SetEntityComponent(playerEntityId, inventoryComponent)

	if emitEvents && item.Type != "" {
		g.EmitGameEvent(gameevent.New("collect:item:"+gameevent.NormalizeToken(item.Type), playerEntityId))
	}
	if emitEvents && item.Name != "" {
		g.EmitGameEvent(gameevent.New("collect:name:"+gameevent.NormalizeToken(item.Name), playerEntityId))
	}
	return true
}

func (g *Game) spawnRewardDrop(playerEntityId model.EntityId, items []component.LootItem) model.EntityId {
	position := g.rewardDropPosition(playerEntityId)
	metadata := util.JObject(map[string]util.Json{
		"name":           util.JString("Reward Parcel"),
		"type":           util.JString("rewarddrop"),
		"width":          util.JNumber(1),
		"height":         util.JNumber(1),
		"blocksMovement": util.JBool(false),
	})
	return g.componentManager.CreateNewEntity(
		component.NewCPosition(position),
		component.NewCMetadata(metadata),
		component.NewCRenderable("rewarddrop"),
		component.NewCLootable(true, items),
		component.NewCRewardDrop(),
	)
}

func (g *Game) rewardDropPosition(playerEntityId model.EntityId) math.Vec2 {
	positionComponent := g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerEntityId)
	if positionComponent == nil {
		return g.world.GetPlayerSpawn()
	}

	playerPosition := positionComponent.(*component.CPosition).GetPosition()
	checker := collision.Checker{
		World:            g.world,
		ComponentManager: g.componentManager,
		SpatialIndex:     g.spatialIndex,
	}
	directions := []math.Vec2{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
	}
	for _, direction := range directions {
		candidate := playerPosition.Add(direction)
		if !checker.IsBlocked(candidate.X, candidate.Y) {
			return candidate
		}
	}
	return playerPosition
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
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		return
	}

	positionComponent := g.componentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
	if !ok || positionComponent == nil {
		panic("position component not found")
	}
	targetChunk, _ := g.world.GlobalToChunk(x, y)
	state := g.clients[clientID]
	if state == nil || !state.loadedChunks[targetChunk] || g.world.GetStaticWall(x, y) {
		return
	}

	pathingComponent := component.NewCPathing(component.PathingTarget{
		Position: util.OptionalSome(math.Vec2{X: x, Y: y}),
	})
	g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, entityId)
	g.componentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
	g.componentManager.RemoveComponent(component.ComponentIdWoodcutting, entityId)
	g.componentManager.SetEntityComponent(entityId, pathingComponent)
}

func (g *Game) HandleLeave(clientID string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if ok {
		g.componentManager.RemoveEntity(entityId)
		g.clientIdToEntityId.Delete(clientID)
	}
	delete(g.clients, clientID)
}

func (g *Game) HandleChat(clientID string, chatMessage string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		return
	}

	g.EmitGameEvent(gameevent.NewChatSpoken(entityId, chatMessage))
}

func (g *Game) HandleInteract(clientID string, entityId model.EntityId, option component.InteractionOption) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	if !g.entityVisibleToClient(clientID, entityId) {
		return
	}
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
		return
	}
	g.componentManager.RemoveComponent(component.ComponentIdWoodcutting, interactingEntityId)

	// Set pathing component to path to the target entity
	pathingComponent := component.NewCPathing(component.PathingTarget{
		EntityId: util.OptionalSome(entityId),
	})
	g.componentManager.SetEntityComponent(interactingEntityId, pathingComponent)

	// Set interacting component to track the interaction
	interactingComponent := component.NewCInteracting(entityId, option)
	g.componentManager.SetEntityComponent(interactingEntityId, interactingComponent)
}

func (g *Game) entityVisibleToClient(clientID string, entityID model.EntityId) bool {
	state := g.clients[clientID]
	if state == nil {
		return false
	}
	if ownID, ok := g.clientIdToEntityId.Get(clientID); ok && ownID == entityID {
		return true
	}
	for coord := range g.spatialIndex.EntityChunks(entityID) {
		if state.loadedChunks[coord] {
			return true
		}
	}
	return false
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

	inventoryComponent := inventory.(*component.CInventory)
	totalItemCount := 0
	for _, lootItem := range lootableComponent.GetItems() {
		count := lootItem.Count
		if count < 1 {
			count = 1
		}
		totalItemCount += count
	}
	if totalItemCount > inventoryComponent.AvailableSlots() {
		return
	}

	allItemsAdded := true
	for _, lootItem := range lootableComponent.GetItems() {
		count := lootItem.Count
		if count < 1 {
			count = 1
		}
		for i := 0; i < count; i++ {
			if !g.AddItemToPlayerInventory(playerEntityId, lootItem.CreateItem()) {
				allItemsAdded = false
				break
			}
		}
		if !allItemsAdded {
			break
		}
	}

	if allItemsAdded && lootableComponent.IsOnce() {
		lootableComponent.SetLooted(true)
		g.componentManager.SetEntityComponent(targetEntityId, lootableComponent)
		if g.componentManager.GetEntityComponent(component.ComponentIdRewardDrop, targetEntityId) != nil {
			g.componentManager.RemoveEntity(targetEntityId)
		}
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
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		return
	}

	active := g.componentManager.GetEntityComponent(component.ComponentIdActiveConversation, entityId)
	if active == nil {
		return
	}

	activeConversation := active.(*component.CActiveConversation)
	if !g.entityVisibleToClient(clientID, activeConversation.GetTargetEntityId()) {
		g.componentManager.RemoveComponent(component.ComponentIdActiveConversation, entityId)
		return
	}
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
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
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
	if !inventoryComponent.RemoveItem(itemId) {
		return
	}
	previousItem := equippedComponent.EquipItem(slot, item)
	if previousItem != nil {
		inventoryComponent.AddItem(previousItem)
	}

	g.componentManager.SetEntityComponent(entityId, inventoryComponent)
	g.componentManager.SetEntityComponent(entityId, equippedComponent)

	combatStats := component.CalculateCombatStats(baseStatsComponent, equippedComponent)
	g.componentManager.SetEntityComponent(entityId, combatStats)
}

func (g *Game) HandleUnequip(clientID string, slot model.EquipmentSlot) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
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

	item := equippedComponent.GetEquippedItem(slot)
	if item == nil || inventoryComponent.IsFull() {
		return
	}
	equippedComponent.UnequipItem(slot)
	inventoryComponent.AddItem(item)

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

	openable := g.componentManager.GetEntityComponent(component.ComponentIdOpenable, entityId)
	if openable != nil {
		if openable.(*component.COpenable).IsOpen() {
			options = append(options, component.InteractionOptionClose)
		} else {
			options = append(options, component.InteractionOptionOpen)
		}
	}

	woodcuttable := g.componentManager.GetEntityComponent(component.ComponentIdWoodcuttable, entityId)
	if woodcuttable != nil && !woodcuttable.(*component.CWoodcuttable).IsDepleted() {
		options = append(options, component.InteractionOptionChop)
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
