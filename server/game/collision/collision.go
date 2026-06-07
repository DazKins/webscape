package collision

import (
	"errors"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

type StaticWorld interface {
	GetStaticWall(x int, y int) bool
}

type Checker struct {
	World            StaticWorld
	ComponentManager *component.ComponentManager
}

func (c Checker) IsBlocked(x int, y int) bool {
	if c.World.GetStaticWall(x, y) {
		return true
	}
	if c.ComponentManager == nil {
		return false
	}

	for entityId, comp := range c.ComponentManager.GetComponent(component.ComponentIdPosition) {
		position := comp.(*component.CPosition).GetPosition()
		width, height := c.entitySize(entityId)
		if x < position.X || y < position.Y || x >= position.X+width || y >= position.Y+height {
			continue
		}
		if c.entityBlocksMovement(entityId) {
			return true
		}
	}
	return false
}

func (c Checker) GetPath(from math.Vec2, to math.Vec2) (util.Path, error) {
	frontier := util.NewQueue[math.Vec2]()
	cameFrom := make(map[math.Vec2]math.Vec2)
	costSoFar := make(map[math.Vec2]float64)

	frontier.Enqueue(from)
	costSoFar[from] = 0.0

	for frontier.Size() > 0 {
		current := frontier.Dequeue()

		if current.X == to.X && current.Y == to.Y {
			break
		}

		canEast := !c.IsBlocked(current.X+1, current.Y)
		canWest := !c.IsBlocked(current.X-1, current.Y)
		canNorth := !c.IsBlocked(current.X, current.Y+1)
		canSouth := !c.IsBlocked(current.X, current.Y-1)

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

				if c.IsBlocked(posX, posY) {
					continue
				}

				neighbor := math.Vec2{X: posX, Y: posY}
				dist := 1.0
				if x != 0 && y != 0 {
					dist = 1.0001
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
	current := to
	for current != from {
		path.Append(current)
		cameFrom, ok := cameFrom[current]
		if !ok {
			return util.Path{}, errors.New("no path found")
		}
		current = cameFrom
	}
	return path.Reversed(), nil
}

func (c Checker) entityBlocksMovement(entityId model.EntityId) bool {
	metadataComponent := c.ComponentManager.GetEntityComponent(component.ComponentIdMetadata, entityId)
	if metadataComponent == nil {
		return false
	}
	metadata, ok := metadataComponent.(*component.CMetadata).GetMetadata().(util.JObject)
	if !ok {
		return false
	}
	blocksMovement, ok := metadata["blocksMovement"].(util.JBool)
	if !ok || !bool(blocksMovement) {
		return false
	}

	openable := c.ComponentManager.GetEntityComponent(component.ComponentIdOpenable, entityId)
	if openable == nil {
		return true
	}
	return !openable.(*component.COpenable).IsOpen()
}

func (c Checker) entitySize(entityId model.EntityId) (int, int) {
	width := 1
	height := 1
	metadataComponent := c.ComponentManager.GetEntityComponent(component.ComponentIdMetadata, entityId)
	if metadataComponent == nil {
		return width, height
	}
	metadata, ok := metadataComponent.(*component.CMetadata).GetMetadata().(util.JObject)
	if !ok {
		return width, height
	}
	if value, ok := metadata["width"].(util.JNumber); ok && value >= 1 {
		width = int(value)
	}
	if value, ok := metadata["height"].(util.JNumber); ok && value >= 1 {
		height = int(value)
	}
	return width, height
}
