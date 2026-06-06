package system

import (
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/world"
)

type SpawnSystem struct {
	SystemBase
}

func (s *SpawnSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdSpawn)

	for _, entityId := range entityIds {
		spawn := s.ComponentManager.GetEntityComponent(component.ComponentIdSpawn, entityId).(*component.CSpawn)
		s.updateSpawn(spawn)
	}
}

func (s *SpawnSystem) updateSpawn(spawn *component.CSpawn) {
	if spawn.HasChildEntityId() {
		childEntityId := spawn.GetChildEntityId()
		if s.ComponentManager.HasEntity(childEntityId) {
			return
		}

		spawn.ClearChildEntityId()
		spawn.SetRemainingRespawnTicks(spawn.GetRespawnTicks())
		if spawn.GetRemainingRespawnTicks() > 0 {
			return
		}
	}

	if !spawn.HasSpawned() {
		s.spawnChild(spawn)
		return
	}

	if spawn.GetRemainingRespawnTicks() > 0 {
		spawn.SetRemainingRespawnTicks(spawn.GetRemainingRespawnTicks() - 1)
		if spawn.GetRemainingRespawnTicks() > 0 {
			return
		}
	}

	s.spawnChild(spawn)
}

func (s *SpawnSystem) spawnChild(spawn *component.CSpawn) {
	templateComponents := spawn.GetChildTemplateComponents()
	delete(templateComponents, component.ComponentIdPosition.String())
	templateComponents[component.ComponentIdPosition.String()] = map[string]any{
		"x": spawn.GetSpawnPosition().X,
		"y": spawn.GetSpawnPosition().Y,
	}

	components := entity.CreateAuthoredEntity(world.WorldEntity{
		Id:         spawn.GetChildTemplateEntityId(),
		Components: templateComponents,
	})
	childEntityId := s.ComponentManager.CreateNewEntity(components...)
	spawn.SetChildEntityId(childEntityId)
	spawn.SetRemainingRespawnTicks(0)
	spawn.MarkSpawned()
}
