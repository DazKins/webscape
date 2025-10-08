package entity

import (
	"webscape/server/game/entity/component"

	"github.com/google/uuid"
)

type EntityId uuid.UUID

func (id EntityId) String() string {
	return uuid.UUID(id).String()
}

type Entity struct {
	id         EntityId
	components []component.Component
}

func NewEntity(id EntityId) *Entity {
	return &Entity{
		id:         id,
		components: []component.Component{},
	}
}

func (e *Entity) GetId() EntityId {
	return e.id
}

func (e *Entity) AddComponent(component component.Component) {
	e.components = append(e.components, component)
}

func (e *Entity) GetComponent(id component.ComponentId) component.Component {
	for _, component := range e.components {
		if component.GetId() == id {
			return component
		}
	}
	return nil
}

func (e *Entity) GetComponents() []component.Component {
	return e.components
}

func (e *Entity) Update() bool {
	anyUpdated := false
	for _, component := range e.components {
		if component.Update() {
			anyUpdated = true
		}
	}
	return anyUpdated
}
