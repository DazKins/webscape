package game

import (
	"webscape/server/game/component"
	"webscape/server/game/entity"
)

func (g *Game) getPathingEntities() []entity.Entity {
	entities := make([]entity.Entity, 0)
	for _, entity := range g.entities.Values() {
		if _, ok := entity.GetComponent(component.ComponentIdPosition); !ok {
			continue
		}
		if _, ok := entity.GetComponent(component.ComponentIdPathing); !ok {
			continue
		}
		entities = append(entities, *entity)
	}
	return entities
}

func (g *Game) updatePathing() {
	entities := g.getPathingEntities()

	for _, entity := range entities {
		positionComponentI, _ := entity.GetComponent(component.ComponentIdPosition)
		pathingComponentI, _ := entity.GetComponent(component.ComponentIdPathing)

		positionComponent := positionComponentI.(*component.CPosition)
		pathingComponent := pathingComponentI.(*component.CPathing)

		path := pathingComponent.GetPath()

		if path.Size() == 0 {
			entity.RemoveComponent(component.ComponentIdPathing)
			continue
		}

		nextPosition := path.Pop()
		positionComponent.SetPosition(*nextPosition)
	}
}
