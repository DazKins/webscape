package system

import (
	"math/rand"
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
)

const WALK_TIMER = 2
const WALK_TIMER_VARIANCE = 1

var DIRECTIONS = []math.Vec2{
	{X: 0, Y: 1},
	{X: 1, Y: 0},
	{X: 0, Y: -1},
	{X: -1, Y: 0},
}

type RandomWalkSystem struct {
	SystemBase
	World *world.World
}

func (s *RandomWalkSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdPosition, component.ComponentIdRandomWalk)

	for _, entityId := range entityIds {
		randomwalkComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdRandomWalk, entityId).(*component.CRandomWalk)
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)

		newTimer := randomwalkComponent.GetWalkTimer() - 1
		randomwalkComponent.SetWalkTimer(newTimer)
		if newTimer <= 0 {
			direction := DIRECTIONS[rand.Intn(len(DIRECTIONS))]
			currentPos := positionComponent.GetPosition()
			newPosition := currentPos.Add(direction)

			if !s.World.GetWall(newPosition.X, newPosition.Y) {
				positionComponent.SetPosition(newPosition)
			}

			randomwalkComponent.SetWalkTimer(WALK_TIMER + rand.Intn(WALK_TIMER_VARIANCE*2) - WALK_TIMER_VARIANCE)
		}
	}
}
