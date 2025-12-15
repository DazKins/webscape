package component

import (
	"webscape/server/game/model"
)

type ComponentManager struct {
	components map[ComponentId]map[model.EntityId]Component
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		components: make(map[ComponentId]map[model.EntityId]Component),
	}
}

func (c *ComponentManager) GetComponent(componentId ComponentId) map[model.EntityId]Component {
	entities, ok := c.components[componentId]
	if !ok {
		entities = make(map[model.EntityId]Component)
		c.components[componentId] = entities
	}
	return entities
}

func (c *ComponentManager) GetEntityComponent(componentId ComponentId, entityId model.EntityId) Component {
	return c.GetComponent(componentId)[entityId]
}

func (c *ComponentManager) GetAllComponents() map[ComponentId]map[model.EntityId]Component {
	return c.components
}

func (c *ComponentManager) GetEntitiesWithComponents(componentIds ...ComponentId) []model.EntityId {
	if len(componentIds) == 0 {
		return []model.EntityId{}
	}

	// Get entities for the first component
	firstComponentEntities := c.GetComponent(componentIds[0])
	result := make([]model.EntityId, 0)

	// For each entity in the first component, check if it has all other components
	for entityId := range firstComponentEntities {
		hasAll := true
		for i := 1; i < len(componentIds); i++ {
			componentEntities := c.GetComponent(componentIds[i])
			if _, exists := componentEntities[entityId]; !exists {
				hasAll = false
				break
			}
		}
		if hasAll {
			result = append(result, entityId)
		}
	}

	return result
}

func (c *ComponentManager) SetEntityComponent(entityId model.EntityId, component Component) {
	entities := c.GetComponent(component.GetId())
	entities[entityId] = component
}

func (c *ComponentManager) SetEntityComponents(entityId model.EntityId, components ...Component) {
	for _, component := range components {
		c.SetEntityComponent(entityId, component)
	}
}

func (c *ComponentManager) CreateNewEntity(components ...Component) model.EntityId {
	entityId := model.NewEntityId()
	c.SetEntityComponents(entityId, components...)
	return entityId
}

func (c *ComponentManager) RemoveComponent(componentId ComponentId, entityId model.EntityId) {
	entities := c.GetComponent(componentId)
	delete(entities, entityId)
}

func (c *ComponentManager) RemoveEntity(entityId model.EntityId) {
	for _, components := range c.components {
		delete(components, entityId)
	}
}
