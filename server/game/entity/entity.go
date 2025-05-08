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
	Color    string

	targetPosition *math.Vec2

	updated bool
}

func NewEntity(id EntityId, position math.Vec2) *Entity {
	randomColor := fmt.Sprintf("#%06x", rand.Intn(0xffffff))
	return &Entity{
		ID:             id,
		Position:       position,
		Color:          randomColor,
		targetPosition: nil,
		updated:        false,
	}
}

func (e *Entity) SetTargetPosition(targetPosition math.Vec2) {
	e.targetPosition = &targetPosition
}

func (e *Entity) Update() {
	e.updated = false

	if e.targetPosition != nil {
		if e.Position.X < e.targetPosition.X {
			e.Position.X += 1
		} else if e.Position.X > e.targetPosition.X {
			e.Position.X -= 1
		}

		if e.Position.Y < e.targetPosition.Y {
			e.Position.Y += 1
		} else if e.Position.Y > e.targetPosition.Y {
			e.Position.Y -= 1
		}

		e.updated = true

		if e.Position.X == e.targetPosition.X && e.Position.Y == e.targetPosition.Y {
			e.targetPosition = nil
		}
	}
}

func (e *Entity) WasUpdated() bool {
	return e.updated
}
