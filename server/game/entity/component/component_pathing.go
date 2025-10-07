package component

import (
	"log"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdPathing ComponentId = "pathing"

type CPathing struct {
	path              *util.Path
	world             *world.World
	positionComponent *CPosition
}

func NewCPathing(world *world.World, positionComponent *CPosition) *CPathing {
	if positionComponent == nil {
		panic("positionComponent is nil")
	}
	if world == nil {
		panic("world is nil")
	}

	return &CPathing{
		path:              nil,
		world:             world,
		positionComponent: positionComponent,
	}
}

func (c *CPathing) GetId() ComponentId {
	return ComponentIdPathing
}

func (c *CPathing) Update() bool {
	if c.path != nil {
		nextPos := c.path.Pop()
		c.positionComponent.Position = nextPos

		if c.path.Size() == 0 {
			c.path = nil
		}

		return true
	}

	return false
}

func (c *CPathing) PathTo(targetPosition math.Vec2) {
	frontier := util.NewQueue[math.Vec2]()
	frontier.Enqueue(c.positionComponent.Position)

	cameFrom := make(map[math.Vec2]math.Vec2)
	costSoFar := make(map[math.Vec2]float64)

	costSoFar[c.positionComponent.Position] = 0.0

	for frontier.Size() > 0 {
		current := frontier.Dequeue()

		if current.X == targetPosition.X && current.Y == targetPosition.Y {
			break
		}

		canEast := !c.world.GetWall(current.X+1, current.Y)
		canWest := !c.world.GetWall(current.X-1, current.Y)
		canNorth := !c.world.GetWall(current.X, current.Y+1)
		canSouth := !c.world.GetWall(current.X, current.Y-1)

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

				if c.world.GetWall(posX, posY) {
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

	path := util.Path{}
	current := targetPosition
	for current != c.positionComponent.Position {
		path.Append(current)
		cameFrom, ok := cameFrom[current]
		if !ok {
			log.Println("no path found")
			return
		}
		current = cameFrom
	}
	c.path = util.Ptr(path.Reversed())
}
