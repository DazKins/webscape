package system

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

type ManaSystem struct {
	SystemBase
	TickSource TickSource
	Settings   model.ManaSettings
}

func (s *ManaSystem) Update() {
	if s.Settings.RegenIntervalTicks <= 0 || currentTick(s.TickSource)%uint64(s.Settings.RegenIntervalTicks) != 0 {
		return
	}
	for _, id := range s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdMana, component.ComponentIdHealth) {
		health := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, id).(*component.CHealth)
		mana := s.ComponentManager.GetEntityComponent(component.ComponentIdMana, id).(*component.CMana)
		if health.GetCurrentHealth() > 0 && mana.GetCurrentMana() < mana.GetMaxMana() {
			mana.Restore(s.Settings.RegenAmount)
			s.ComponentManager.SetEntityComponent(id, mana)
		}
	}
}
