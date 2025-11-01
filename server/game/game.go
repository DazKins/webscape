package game

import (
	"log"
	"math/rand"
	"time"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

type MessageBroadcaster func(message message.Message)
type MessageSender func(clientID string, message message.Message)

type Game struct {
	world              *world.World
	clientIdToEntityId *util.BiMap[string, entity.EntityId]
	ticker             *time.Ticker
	done               chan bool
	sendMessage        MessageSender
	broadcastMessage   MessageBroadcaster

	entities *util.IdMap[*entity.Entity, entity.EntityId]
}

var NAMES = []string{"Bob", "Alice", "Charlie", "David", "Eve", "Frank", "George", "Hannah", "Isaac", "Jack"}

func NewGame() *Game {
	world := world.NewWorld(10, 10)

	game := &Game{
		clientIdToEntityId: util.NewBiMap[string, entity.EntityId](),
		world:              world,
		done:               make(chan bool),

		entities: util.NewIdMap[*entity.Entity, entity.EntityId](),
	}

	for i := range 3 {
		position := math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}
		for world.GetWall(position.X, position.Y) {
			position = math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}
		}

		game.AddEntity(entity.CreateDudeEntity(
			NAMES[i],
			position,
		))
	}

	return game
}

func (g *Game) AddEntity(entity *entity.Entity) {
	g.entities.Put(entity)
}

func (g *Game) GetEntity(entityId entity.EntityId) (*entity.Entity, bool) {
	return g.entities.GetById(entityId)
}

func (g *Game) RemoveEntity(entityId entity.EntityId) {
	g.entities.DeleteById(entityId)
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
	g.updatePathing()

	for _, entity := range g.entities.Values() {
		for _, comp := range entity.GetComponents().Values() {
			if comp.ShouldSend() {
				serializeableComp, ok := comp.(component.SerializeableComponent)
				if !ok {
					continue
				}

				g.broadcastMessage(message.NewComponentUpdateMessage(
					entity.GetId(),
					comp.GetId(),
					serializeableComp.Serialize(),
				))

				comp.MarkSent()
			}
		}
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

func (g *Game) HandleJoin(clientID string, id entity.EntityId, name string) {
	rand.Seed(time.Now().UnixNano())

	if _, ok := g.clientIdToEntityId.Get(clientID); ok {
		g.sendMessage(clientID, message.NewJoinFailedMessage("you are already connected in another session"))
		return
	}

	playerEntity := entity.CreatePlayerEntity(id, name, g.world)
	g.AddEntity(playerEntity)

	g.clientIdToEntityId.Put(clientID, id)

	g.sendMessage(clientID, message.NewJoinedMessage(id.String()))
	g.sendMessage(clientID, message.NewWorldMessage(g.world))

	for _, e := range g.entities.Values() {
		components := e.GetComponents()
		for _, comp := range components.Values() {
			if serializeableComp, ok := comp.(component.SerializeableComponent); ok {
				g.sendMessage(clientID, message.NewComponentUpdateMessage(
					e.GetId(),
					comp.GetId(),
					serializeableComp.Serialize(),
				))
			}
		}
	}
}

func (g *Game) HandleMove(clientID string, x int, y int) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		panic("client ID not found in clientIdToEntityId")
	}

	entity, ok := g.GetEntity(entityId)
	if !ok {
		panic("entity not found")
	}

	positionComponentI, ok := entity.GetComponent(component.ComponentIdPosition)
	if !ok {
		panic("position component not found")
	}
	positionComponent := positionComponentI.(*component.CPosition)

	path, err := g.getPath(positionComponent.GetPosition(), math.Vec2{X: x, Y: y})
	if err != nil {
		log.Printf("failed to get path: %v\n", err)
		return
	}

	pathingComponent := &component.CPathing{}
	pathingComponent.SetPath(&path)
	entity.SetComponent(pathingComponent)
}

func (g *Game) HandleLeave(clientID string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.clientIdToEntityId.Delete(clientID)
	g.RemoveEntity(entityId)

	g.broadcastMessage(message.NewEntityRemoveMessage(entityId))
}

func (g *Game) HandleChat(clientID string, chatMessage string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.broadcastMessage(message.NewChatMessage(entityId.String(), chatMessage))
}

func (g *Game) HandleInteract(clientID string, entityId entity.EntityId, option component.InteractionOption) {
	entity, ok := g.GetEntity(entityId)
	if !ok {
		panic("entity not found")
	}

	interactableComponentI, ok := entity.GetComponent(component.ComponentIdInteractable)
	if !ok {
		panic("interactable component not found")
	}
	interactableComponent := interactableComponentI.(*component.CInteractable)

	for _, interactionOption := range interactableComponent.GetInteractionOptions() {
		if interactionOption == option {
			// TODO
		}
	}

	log.Println("Invalid interaction option")
}
