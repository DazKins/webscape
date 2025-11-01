package entity

import (
	"webscape/server/game/component"
	"webscape/server/util"

	"github.com/google/uuid"
)

type EntityId uuid.UUID

func NewEntityId() EntityId {
	return EntityId(uuid.New())
}

func (e EntityId) String() string {
	return uuid.UUID(e).String()
}

type Entity struct {
	id         EntityId
	components *util.IdMap[component.Component, component.ComponentId]
}

func NewEntity(id EntityId) *Entity {
	return &Entity{
		id:         id,
		components: util.NewIdMap[component.Component](),
	}
}

func (e *Entity) GetId() EntityId {
	return e.id
}

func (e *Entity) GetComponents() *util.IdMap[component.Component, component.ComponentId] {
	return e.components
}

func (e *Entity) GetComponent(componentId component.ComponentId) component.Component {
	component, ok := e.components.GetById(componentId)
	if !ok {
		return nil
	}
	return component
}

func (e *Entity) SetComponent(component component.Component) *Entity {
	e.components.Put(component)
	return e
}

func (e *Entity) RemoveComponent(componentId component.ComponentId) *Entity {
	e.components.DeleteById(componentId)
	return e
}
