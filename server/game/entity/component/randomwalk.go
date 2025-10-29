package component

import (
	"math/rand"
	"webscape/server/game/world"
	"webscape/server/math"
)

const ComponentIdRandomWalk ComponentId = "randomwalk"

const BASE_WALK_TIMER = 10
const WALK_TIMER_VARIANCE = 5

type CRandomWalk struct {
	world             *world.World
	positionComponent *CPosition
	walkTimer         int
}

func generateWalkTimer() int {
	return BASE_WALK_TIMER + rand.Intn(WALK_TIMER_VARIANCE*2) - WALK_TIMER_VARIANCE
}

func NewCRandomWalk(world *world.World, positionComponent *CPosition) *CRandomWalk {
	return &CRandomWalk{
		world:             world,
		positionComponent: positionComponent,
		walkTimer:         generateWalkTimer(),
	}
}

func (c *CRandomWalk) GetId() ComponentId {
	return ComponentIdRandomWalk
}

func (c *CRandomWalk) Update() bool {
	c.walkTimer--
	if c.walkTimer > 0 {
		return false
	}
	c.walkTimer = generateWalkTimer()

	position := c.positionComponent.GetPosition()

	var randomX, randomY int
	if rand.Intn(2) == 0 {
		randomX = rand.Intn(3) - 1
		randomY = 0
	} else {
		randomX = 0
		randomY = rand.Intn(3) - 1
	}

	newPosition := math.Vec2{X: position.X + randomX, Y: position.Y + randomY}

	if c.world.GetWall(newPosition.X, newPosition.Y) {
		return false
	}

	c.positionComponent.SetPosition(newPosition)

	return true
}
