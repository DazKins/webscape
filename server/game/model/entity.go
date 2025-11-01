package model

import "github.com/google/uuid"

type EntityId uuid.UUID

func NewEntityId() EntityId {
	return EntityId(uuid.New())
}

func (e EntityId) String() string {
	return uuid.UUID(e).String()
}
