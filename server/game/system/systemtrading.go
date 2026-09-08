package system

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

type TradingValidator interface {
	CanTradeWith(player, target model.EntityId) bool
}
type TradingSystem struct {
	SystemBase
	Validator TradingValidator
}

func (s *TradingSystem) Update() {
	for id, raw := range s.ComponentManager.GetComponent(component.ComponentIdTrading) {
		if !s.Validator.CanTradeWith(id, raw.(*component.CTrading).TargetEntityId) {
			s.ComponentManager.RemoveComponent(component.ComponentIdTrading, id)
		}
	}
}
