package entity

import (
	"fmt"
	"math/rand"
	"webscape/server/math"

	"github.com/google/uuid"
)

type EntityId uuid.UUID

func (id EntityId) String() string {
	return uuid.UUID(id).String()
}

type Entity struct {
	ID       EntityId
	Position math.Vec2
	Velocity math.Vec2
	Color    string
	Name     string

	targetPosition *math.Vec2

	updated bool
}

func NewEntity(id EntityId, position math.Vec2, name string) *Entity {
	randomColor := fmt.Sprintf("#%06x", rand.Intn(0xffffff))
	return &Entity{
		ID:             id,
		Position:       position,
		Velocity:       math.Vec2Zero(),
		Color:          randomColor,
		Name:           name,
		targetPosition: nil,
		updated:        false,
	}
}

func (e *Entity) SetTargetPosition(targetPosition math.Vec2) {
	e.targetPosition = &targetPosition
}

func getNextMove(position math.Vec2, targetPosition math.Vec2) math.Vec2 {
	ret := math.Vec2Zero()

	if position.X < targetPosition.X {
		ret.X += 1
	} else if position.X > targetPosition.X {
		ret.X -= 1
	}

	if position.Y < targetPosition.Y {
		ret.Y += 1
	} else if position.Y > targetPosition.Y {
		ret.Y -= 1
	}

	return ret
}

func (e *Entity) Update() {
	e.updated = false

	if e.targetPosition != nil {
		e.Velocity = getNextMove(e.Position, *e.targetPosition)
		e.updated = true

		if e.Position.X == e.targetPosition.X && e.Position.Y == e.targetPosition.Y {
			e.targetPosition = nil
		}
	}

	if !e.Velocity.IsZero() {
		e.Position = e.Position.Add(e.Velocity)
		e.updated = true
	}
}

func (e *Entity) WasUpdated() bool {
	return e.updated
}
