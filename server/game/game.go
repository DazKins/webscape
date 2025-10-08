package game

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"webscape/server/game/entity"
	"webscape/server/game/entity/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"

	"github.com/google/uuid"
)

type MessageBroadcaster func(message message.Message)
type MessageSender func(clientID string, message message.Message)

type Game struct {
	world              *world.World
	clientIdToEntityId *util.BiMap[string, entity.EntityId]
	ticker             *time.Ticker
	done               chan bool
	entities           *util.IdMap[*entity.Entity, entity.EntityId]
	savedEntities      *util.IdMap[*entity.Entity, entity.EntityId]
	sendMessage        MessageSender
	broadcastMessage   MessageBroadcaster
}

var NAMES = []string{"Bob", "Alice", "Charlie", "David", "Eve", "Frank", "George", "Hannah", "Isaac", "Jack"}

func NewGame() *Game {
	world := world.NewWorld(10, 10)
	entities := util.NewIdMap[*entity.Entity]()

	for i := range 3 {
		entity := entity.NewEntity(entity.EntityId(uuid.New()))
		entity.AddComponent(component.NewCPosition(math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}))
		entity.AddComponent(component.NewCMetadata(map[string]any{
			"name":  NAMES[i],
			"color": fmt.Sprintf("#%06X", rand.Intn(0xFFFFFF)),
		}))
		entities.Put(entity)
	}

	return &Game{
		clientIdToEntityId: util.NewBiMap[string, entity.EntityId](),
		world:              world,
		done:               make(chan bool),
		entities:           entities,
		savedEntities:      util.NewIdMap[*entity.Entity](),
	}
}

func (g *Game) AddEntity(e *entity.Entity) {
	g.entities.Put(e)
	g.broadcastMessage(message.NewEntityUpdateMessage(e))
}

func (g *Game) RemoveEntity(e *entity.Entity) {
	g.entities.Delete(e)
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
	updatedEntities := make([]*entity.Entity, 0)

	for _, entity := range g.entities.Values() {
		if entity.Update() {
			updatedEntities = append(updatedEntities, entity)
		}
	}

	for _, entity := range updatedEntities {
		g.broadcastMessage(message.NewEntityUpdateMessage(entity))
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

	if _, ok := g.entities.GetById(id); ok {
		g.sendMessage(clientID, message.NewJoinFailedMessage("you are already connected in another session"))
		return
	}

	if _, ok := g.clientIdToEntityId.Get(clientID); ok {
		g.sendMessage(clientID, message.NewJoinFailedMessage("you are already connected in another session"))
		return
	}

	playerEntity := (*entity.Entity)(nil)
	if savedEntity, ok := g.savedEntities.GetById(id); ok {
		playerEntity = savedEntity
		g.savedEntities.Delete(savedEntity)
	} else {
		x := rand.Intn(g.world.GetSizeX()) - g.world.GetSizeX()/2
		y := rand.Intn(g.world.GetSizeY()) - g.world.GetSizeY()/2

		playerEntity = entity.NewEntity(id)
		positionComponent := component.NewCPosition(math.Vec2{X: x, Y: y})
		playerEntity.AddComponent(positionComponent)
		playerEntity.AddComponent(component.NewCPathing(
			g.world,
			positionComponent,
		))
		playerEntity.AddComponent(component.NewCMetadata(map[string]any{
			"name": name,
		}))
	}

	g.AddEntity(playerEntity)
	g.clientIdToEntityId.Put(clientID, playerEntity.GetId())

	g.sendMessage(clientID, message.NewJoinedMessage(playerEntity.GetId().String()))
	g.sendMessage(clientID, message.NewWorldMessage(g.world))

	for _, e := range g.entities.Values() {
		if e.GetId() == id {
			continue
		}
		g.sendMessage(clientID, message.NewEntityUpdateMessage(e))
	}
}

func (g *Game) HandleMove(clientID string, x int, y int) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		panic("Client ID not found in clientIdToEntityId")
	}

	entity, ok := g.entities.GetById(entityId)
	if !ok {
		panic("Entity not found in entities")
	}

	pathingComponent := entity.GetComponent(component.ComponentIdPathing).(*component.CPathing)
	if pathingComponent == nil {
		panic("Pathing component not found in entity")
	}

	pathingComponent.PathTo(math.Vec2{X: x, Y: y})
}

func (g *Game) HandleLeave(clientID string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	entity, ok := g.entities.GetById(entityId)
	if !ok {
		panic("Entity not found in entities")
	}

	g.clientIdToEntityId.Delete(clientID)
	g.RemoveEntity(entity)
	g.savedEntities.Put(entity)

	g.broadcastMessage(message.NewEntityRemoveMessage(entityId))
}
