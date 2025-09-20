package game

import (
	"log"
	"math/rand"
	"time"
	"webscape/server/game/entity"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"

	"github.com/google/uuid"
)

type MessageBroadcaster func(message message.Message)
type MessageSender func(clientID string, message message.Message)

type Game struct {
	entities           map[entity.EntityId]*entity.Entity
	world              *world.World
	clientIdToEntityId *util.BiMap[string, entity.EntityId]

	broadcastMessage MessageBroadcaster
	sendMessage      MessageSender
	ticker           *time.Ticker
	done             chan bool
}

var NAMES = []string{"Bob", "Alice", "Charlie", "David", "Eve", "Frank", "George", "Hannah", "Isaac", "Jack"}

func NewGame() *Game {
	entities := make(map[entity.EntityId]*entity.Entity)
	for i := 0; i < 2; i++ {
		entities[entity.EntityId(uuid.New())] = entity.NewEntity(
			entity.EntityId(uuid.New()),
			math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5},
			NAMES[i],
		)
	}

	return &Game{
		entities:           entities,
		clientIdToEntityId: util.NewBiMap[string, entity.EntityId](),
		world:              world.NewWorld(10, 10),
		done:               make(chan bool),
	}
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
	for _, e := range g.entities {
		e.Update()

		if e.WasUpdated() {
			g.broadcastMessage(message.NewEntityUpdateMessage(e))
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

func (g *Game) AddEntity(e *entity.Entity) {
	g.entities[e.ID] = e
	g.broadcastMessage(message.NewEntityUpdateMessage(e))
}

func (g *Game) RemoveEntity(e *entity.Entity) {
	log.Println("Removing entity", e.ID)
	delete(g.entities, e.ID)
	log.Println("Entities", g.entities)
	g.broadcastMessage(message.NewEntityRemoveMessage(e.ID.String()))
	log.Println("Removed entity", e.ID)
}

func (g *Game) HandleJoin(clientID string, id entity.EntityId, name string) {
	rand.Seed(time.Now().UnixNano())

	if _, ok := g.entities[id]; !ok {
		x := rand.Intn(g.world.GetSizeX()) - g.world.GetSizeX()/2
		y := rand.Intn(g.world.GetSizeY()) - g.world.GetSizeY()/2

		playerEntity := entity.NewEntity(id, math.Vec2{X: x, Y: y}, name)
		g.AddEntity(playerEntity)
		g.sendMessage(clientID, message.NewJoinedMessage(playerEntity.ID.String()))
		g.clientIdToEntityId.Put(clientID, playerEntity.ID)
	}

	g.sendMessage(clientID, message.NewWorldMessage(g.world))

	for _, e := range g.entities {
		if e.ID == id {
			continue
		}
		g.sendMessage(clientID, message.NewEntityUpdateMessage(e))
	}
}

func (g *Game) HandleMove(clientID string, x int, y int) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
	}

	entity, ok := g.entities[entityId]
	if !ok {
		log.Println("Entity not found in entities")
	}

	entity.SetTargetPosition(math.Vec2{X: x, Y: y})
}

func (g *Game) HandleLeave(clientID string) {
	log.Println("Handling leave for clientID", clientID)

	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
	}

	entity, ok := g.entities[entityId]
	if !ok {
		log.Println("Entity not found in entities")
	}

	g.RemoveEntity(entity)
}
