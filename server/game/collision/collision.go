package collision

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/util"
)

type StaticWorld interface {
	GetStaticWall(x int, y int) bool
}

type Checker struct {
	World            StaticWorld
	ComponentManager *component.ComponentManager
	SpatialIndex     interface {
		EntitiesAt(x int, y int) []model.EntityId
	}
}

func (c Checker) IsBlocked(x int, y int) bool {
	if c.World.GetStaticWall(x, y) {
		return true
	}
	if c.ComponentManager == nil {
		return false
	}

	if index, ok := c.SpatialIndex.(interface{ BlocksMovement(int, int) bool }); ok {
		return index.BlocksMovement(x, y)
	}

	candidates := make([]model.EntityId, 0)
	if c.SpatialIndex != nil {
		candidates = c.SpatialIndex.EntitiesAt(x, y)
	} else {
		for entityId := range c.ComponentManager.GetComponent(component.ComponentIdPosition) {
			candidates = append(candidates, entityId)
		}
	}
	for _, entityId := range candidates {
		comp := c.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId)
		if comp == nil {
			continue
		}
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
