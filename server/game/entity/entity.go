package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/util"
)

type Entity struct {
	id         model.EntityId
	components *util.IdMap[component.Component, component.ComponentId]
}

func NewEntity(id model.EntityId) *Entity {
	return &Entity{
		id:         id,
		components: util.NewIdMap[component.Component](),
	}
}

func (e *Entity) GetId() model.EntityId {
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

func (e *Entity) RemoveAllComponents() *Entity {
	e.components = util.NewIdMap[component.Component]()
	return e
}
