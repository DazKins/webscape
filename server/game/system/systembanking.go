package system

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

type BankingValidator interface {
	CanBankWith(player, target model.EntityId) bool
}
type BankingSystem struct {
	SystemBase
	Validator BankingValidator
}

func (s *BankingSystem) Update() {
	for id, raw := range s.ComponentManager.GetComponent(component.ComponentIdBanking) {
		if !s.Validator.CanBankWith(id, raw.(*component.CBanking).TargetEntityId) {
			s.ComponentManager.RemoveComponent(component.ComponentIdBanking, id)
		}
	}
}
