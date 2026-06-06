package component

import (
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdSpawn = ComponentId("spawn")

type CSpawn struct {
	spawnPosition           math.Vec2
	respawnTicks            int
	remainingRespawnTicks   int
	childTemplateEntityId   string
	childTemplateComponents map[string]any
	childEntityId           util.Optional[model.EntityId]
	hasSpawned              bool
}

func NewCSpawn(
	spawnPosition math.Vec2,
	respawnTicks int,
	childTemplateEntityId string,
	childTemplateComponents map[string]any,
) *CSpawn {
	if respawnTicks < 0 {
		respawnTicks = 0
	}
	return &CSpawn{
		spawnPosition:           spawnPosition,
		respawnTicks:            respawnTicks,
		childTemplateEntityId:   childTemplateEntityId,
		childTemplateComponents: cloneComponentMap(childTemplateComponents),
		childEntityId:           util.OptionalNone[model.EntityId](),
	}
}

func (c *CSpawn) GetId() ComponentId {
	return ComponentIdSpawn
}

func (c *CSpawn) GetSpawnPosition() math.Vec2 {
	return c.spawnPosition
}

func (c *CSpawn) GetRespawnTicks() int {
	return c.respawnTicks
}

func (c *CSpawn) GetRemainingRespawnTicks() int {
	return c.remainingRespawnTicks
}

func (c *CSpawn) SetRemainingRespawnTicks(remainingRespawnTicks int) {
	if remainingRespawnTicks < 0 {
		remainingRespawnTicks = 0
	}
	c.remainingRespawnTicks = remainingRespawnTicks
}

func (c *CSpawn) GetChildTemplateEntityId() string {
	return c.childTemplateEntityId
}

func (c *CSpawn) GetChildTemplateComponents() map[string]any {
	return cloneComponentMap(c.childTemplateComponents)
}

func (c *CSpawn) HasChildEntityId() bool {
	return c.childEntityId.IsPresent()
}

func (c *CSpawn) GetChildEntityId() model.EntityId {
	return c.childEntityId.Unwrap()
}

func (c *CSpawn) SetChildEntityId(entityId model.EntityId) {
	c.childEntityId = util.OptionalSome(entityId)
}

func (c *CSpawn) ClearChildEntityId() {
	c.childEntityId = util.OptionalNone[model.EntityId]()
}

func (c *CSpawn) HasSpawned() bool {
	return c.hasSpawned
}

func (c *CSpawn) MarkSpawned() {
	c.hasSpawned = true
}

func cloneComponentMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = cloneComponentValue(value)
	}
	return result
}

func cloneComponentValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(value))
		for key, item := range value {
			result[key] = cloneComponentValue(item)
		}
		return result
	case []any:
		result := make([]any, len(value))
		for i, item := range value {
			result[i] = cloneComponentValue(item)
		}
		return result
	default:
		return value
	}
}
