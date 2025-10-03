package entity

import (
	"fmt"
	"math/rand"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"

	"github.com/google/uuid"
)

type EntityId uuid.UUID

func (id EntityId) String() string {
	return uuid.UUID(id).String()
}

type Entity struct {
	ID       EntityId
	Position math.Vec2
	Color    string
	Name     string

	path  *util.Path
	world *world.World

	updated bool
}

func NewEntity(id EntityId, position math.Vec2, name string, world *world.World) *Entity {
	randomColor := fmt.Sprintf("#%06x", rand.Intn(0xffffff))
	return &Entity{
		ID:       id,
		Position: position,
		Color:    randomColor,
		Name:     name,
		path:     nil,
		updated:  false,
		world:    world,
	}
}

func (e *Entity) HandleMove(targetPosition math.Vec2) {
	e.path = util.Ptr(e.getPath(targetPosition))
}

func (e *Entity) Update() {
	e.updated = false

	if e.path != nil {
		e.Position = e.path.Pop()
		e.updated = true
		if e.path.Size() == 0 {
			e.path = nil
		}
	}
}

func (e *Entity) getPath(targetPosition math.Vec2) util.Path {
	frontier := util.NewQueue[math.Vec2]()
	frontier.Enqueue(math.Vec2{X: e.Position.X, Y: e.Position.Y})

	cameFrom := make(map[math.Vec2]math.Vec2)
	costSoFar := make(map[math.Vec2]float64)

	costSoFar[math.Vec2{X: e.Position.X, Y: e.Position.Y}] = 0.0

	for frontier.Size() > 0 {
		current := frontier.Dequeue()

		if current.X == targetPosition.X && current.Y == targetPosition.Y {
			break
		}

		canEast := !e.world.GetWall(current.X+1, current.Y)
		canWest := !e.world.GetWall(current.X-1, current.Y)
		canNorth := !e.world.GetWall(current.X, current.Y+1)
		canSouth := !e.world.GetWall(current.X, current.Y-1)

		checkStepXMin := -1
		checkStepXMax := 1
		checkStepYMin := -1
		checkStepYMax := 1
		if !canEast {
			checkStepXMax = 0
		}
		if !canWest {
			checkStepXMin = 0
		}
		if !canNorth {
			checkStepYMax = 0
		}
		if !canSouth {
			checkStepYMin = 0
		}

		for x := checkStepXMin; x <= checkStepXMax; x++ {
			for y := checkStepYMin; y <= checkStepYMax; y++ {
				if x == 0 && y == 0 {
					continue
				}

				posX := current.X + x
				posY := current.Y + y

				if e.world.GetWall(posX, posY) {
					continue
				}

				neighbor := math.Vec2{X: posX, Y: posY}
				dist := 1.0
				if x != 0 && y != 0 {
					dist = 1.0001 // slightly prefer straight line movement
				}
				newCost := costSoFar[current] + dist
				if _, ok := costSoFar[neighbor]; !ok || newCost < costSoFar[neighbor] {
					costSoFar[neighbor] = newCost
					frontier.Enqueue(neighbor)
					cameFrom[neighbor] = current
				}
			}
		}
	}

	curPos := math.Vec2{X: e.Position.X, Y: e.Position.Y}
	path := util.Path{}
	current := math.Vec2{X: targetPosition.X, Y: targetPosition.Y}
	for current != curPos {
		path.Append(current)
		current = cameFrom[current]
	}
	return path.Reversed()
}

func (e *Entity) WasUpdated() bool {
	return e.updated
}
