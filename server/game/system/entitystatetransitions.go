package system

import (
	"fmt"
	"webscape/server/game/component"
	"webscape/server/game/model"
)

type entityStateTransition string

const (
	transitionPathing     entityStateTransition = "pathing"
	transitionInteraction entityStateTransition = "interaction"
	transitionMoving      entityStateTransition = "moving"
	transitionFishing     entityStateTransition = "fishing"
	transitionWoodcutting entityStateTransition = "woodcutting"
	transitionCombat      entityStateTransition = "combat"
	transitionCombatEnd   entityStateTransition = "combat-end"
	transitionPathReject  entityStateTransition = "path-reject"
	transitionDeath       entityStateTransition = "death"
)

// incompatibleComponents is the authoritative compatibility policy for
// persistent entity states. Transition methods below apply these removals and
// then install the destination state as one server-side operation.
var incompatibleComponents = map[entityStateTransition][]component.ComponentId{
	transitionPathing: {
		component.ComponentIdActiveConversation,
		component.ComponentIdInteracting,
		component.ComponentIdCombatState,
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
	},
	transitionInteraction: {
		component.ComponentIdActiveConversation,
		component.ComponentIdCombatState,
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
	},
	transitionMoving: {
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
	},
	transitionFishing: {
		component.ComponentIdActiveConversation,
		component.ComponentIdCombatState,
		component.ComponentIdPathing,
		component.ComponentIdInteracting,
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
	},
	transitionWoodcutting: {
		component.ComponentIdActiveConversation,
		component.ComponentIdCombatState,
		component.ComponentIdPathing,
		component.ComponentIdInteracting,
		component.ComponentIdFishing,
		component.ComponentIdWoodcutting,
	},
	transitionCombat: {
		component.ComponentIdActiveConversation,
		component.ComponentIdPathing,
		component.ComponentIdInteracting,
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
	},
	transitionCombatEnd: {
		component.ComponentIdCombatState,
		component.ComponentIdPathing,
		component.ComponentIdInteracting,
	},
	transitionPathReject: {
		component.ComponentIdPathing,
		component.ComponentIdInteracting,
		component.ComponentIdCombatState,
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
	},
	transitionDeath: {
		component.ComponentIdCombatState,
		component.ComponentIdPathing,
		component.ComponentIdInteracting,
		component.ComponentIdWoodcutting,
		component.ComponentIdFishing,
		component.ComponentIdActiveConversation,
	},
}

type EntityStateTransitions struct {
	ComponentManager *component.ComponentManager
	TickSource       TickSource
}

func NewEntityStateTransitions(
	manager *component.ComponentManager,
	tickSource TickSource,
) *EntityStateTransitions {
	return &EntityStateTransitions{
		ComponentManager: manager,
		TickSource:       tickSource,
	}
}

func (t *EntityStateTransitions) BeginPathing(
	entityId model.EntityId,
	pathing *component.CPathing,
) {
	t.applyPolicy(entityId, transitionPathing)
	t.ComponentManager.SetEntityComponent(entityId, pathing)
}

func (t *EntityStateTransitions) BeginInteraction(
	entityId model.EntityId,
	pathing *component.CPathing,
	interacting *component.CInteracting,
) {
	t.applyPolicy(entityId, transitionInteraction)
	t.ComponentManager.SetEntityComponent(entityId, pathing)
	t.ComponentManager.SetEntityComponent(entityId, interacting)
}

func (t *EntityStateTransitions) SettleForInteraction(entityId model.EntityId) bool {
	if t.movedThisTick(entityId) {
		return false
	}
	t.setLocomotionIdle(entityId)
	return true
}

func (t *EntityStateTransitions) BeginMoving(entityId model.EntityId) {
	t.applyPolicy(entityId, transitionMoving)
	value := t.ComponentManager.GetEntityComponent(component.ComponentIdLocomotion, entityId)
	if value == nil {
		return
	}
	locomotion := value.(*component.CLocomotion)
	if locomotion.MarkMoving(t.currentTick()) {
		t.ComponentManager.SetEntityComponent(entityId, locomotion)
	}
}

func (t *EntityStateTransitions) BeginFishing(
	entityId model.EntityId,
	fishing *component.CFishing,
) bool {
	if t.movedThisTick(entityId) {
		return false
	}
	t.applyPolicy(entityId, transitionFishing)
	t.setLocomotionIdle(entityId)
	t.ComponentManager.SetEntityComponent(entityId, fishing)
	return true
}

func (t *EntityStateTransitions) BeginWoodcutting(
	entityId model.EntityId,
	woodcutting *component.CWoodcutting,
) bool {
	if t.movedThisTick(entityId) {
		return false
	}
	t.applyPolicy(entityId, transitionWoodcutting)
	t.setLocomotionIdle(entityId)
	t.ComponentManager.SetEntityComponent(entityId, woodcutting)
	return true
}

func (t *EntityStateTransitions) BeginCombat(
	entityId model.EntityId,
	combat *component.CCombatState,
) {
	t.applyPolicy(entityId, transitionCombat)
	t.setLocomotionIdle(entityId)
	t.ComponentManager.SetEntityComponent(entityId, combat)
}

func (t *EntityStateTransitions) EndCombat(entityId model.EntityId) {
	t.applyPolicy(entityId, transitionCombatEnd)
}

func (t *EntityStateTransitions) RejectPathing(entityId model.EntityId) {
	t.applyPolicy(entityId, transitionPathReject)
}

func (t *EntityStateTransitions) HandleDeath(entityId model.EntityId) {
	t.applyPolicy(entityId, transitionDeath)
	t.setLocomotionIdle(entityId)
}

func (t *EntityStateTransitions) FinishInteraction(entityId model.EntityId) {
	t.removeIfPresent(entityId, component.ComponentIdInteracting)
}

func (t *EntityStateTransitions) applyPolicy(
	entityId model.EntityId,
	transition entityStateTransition,
) {
	for _, componentId := range incompatibleComponents[transition] {
		t.removeIfPresent(entityId, componentId)
	}
}

func (t *EntityStateTransitions) removeIfPresent(
	entityId model.EntityId,
	componentId component.ComponentId,
) {
	if t.ComponentManager.GetEntityComponent(componentId, entityId) != nil {
		t.ComponentManager.RemoveComponent(componentId, entityId)
	}
}

func (t *EntityStateTransitions) movedThisTick(entityId model.EntityId) bool {
	value := t.ComponentManager.GetEntityComponent(component.ComponentIdLocomotion, entityId)
	if value == nil {
		return false
	}
	locomotion := value.(*component.CLocomotion)
	return locomotion.GetPhase() == component.LocomotionPhaseMoving &&
		locomotion.GetLastMovementTick() == t.currentTick()
}

func (t *EntityStateTransitions) setLocomotionIdle(entityId model.EntityId) {
	value := t.ComponentManager.GetEntityComponent(component.ComponentIdLocomotion, entityId)
	if value == nil {
		return
	}
	locomotion := value.(*component.CLocomotion)
	if locomotion.SetIdle(t.currentTick()) {
		t.ComponentManager.SetEntityComponent(entityId, locomotion)
	}
}

func (t *EntityStateTransitions) currentTick() uint64 {
	return currentTick(t.TickSource)
}

func ValidateEntityState(
	manager *component.ComponentManager,
	entityId model.EntityId,
) error {
	locomotionValue := manager.GetEntityComponent(component.ComponentIdLocomotion, entityId)
	isMoving := locomotionValue != nil &&
		locomotionValue.(*component.CLocomotion).GetPhase() == component.LocomotionPhaseMoving

	if manager.GetEntityComponent(component.ComponentIdFishing, entityId) != nil {
		if locomotionValue == nil || isMoving {
			return fmt.Errorf("entity %s is fishing without idle locomotion", entityId)
		}
		if conflicting := firstPresentComponent(manager, entityId,
			component.ComponentIdActiveConversation,
			component.ComponentIdPathing,
			component.ComponentIdInteracting,
			component.ComponentIdCombatState,
			component.ComponentIdWoodcutting,
		); conflicting != "" {
			return fmt.Errorf("entity %s is fishing with incompatible %s state", entityId, conflicting)
		}
	}

	if manager.GetEntityComponent(component.ComponentIdWoodcutting, entityId) != nil {
		if locomotionValue == nil || isMoving {
			return fmt.Errorf("entity %s is woodcutting without idle locomotion", entityId)
		}
		if conflicting := firstPresentComponent(manager, entityId,
			component.ComponentIdActiveConversation,
			component.ComponentIdPathing,
			component.ComponentIdInteracting,
			component.ComponentIdCombatState,
			component.ComponentIdFishing,
		); conflicting != "" {
			return fmt.Errorf("entity %s is woodcutting with incompatible %s state", entityId, conflicting)
		}
	}

	if manager.GetEntityComponent(component.ComponentIdCombatState, entityId) != nil {
		if conflicting := firstPresentComponent(manager, entityId,
			component.ComponentIdActiveConversation,
			component.ComponentIdInteracting,
			component.ComponentIdWoodcutting,
			component.ComponentIdFishing,
		); conflicting != "" {
			return fmt.Errorf("entity %s is in combat with incompatible %s state", entityId, conflicting)
		}
	}

	if manager.GetEntityComponent(component.ComponentIdActiveConversation, entityId) != nil {
		if conflicting := firstPresentComponent(manager, entityId,
			component.ComponentIdPathing,
			component.ComponentIdInteracting,
			component.ComponentIdCombatState,
			component.ComponentIdWoodcutting,
			component.ComponentIdFishing,
		); conflicting != "" {
			return fmt.Errorf("entity %s is conversing with incompatible %s state", entityId, conflicting)
		}
	}

	if healthValue := manager.GetEntityComponent(component.ComponentIdHealth, entityId); healthValue != nil &&
		healthValue.(*component.CHealth).GetCurrentHealth() <= 0 {
		if conflicting := firstPresentComponent(manager, entityId,
			component.ComponentIdActiveConversation,
			component.ComponentIdPathing,
			component.ComponentIdInteracting,
			component.ComponentIdCombatState,
			component.ComponentIdWoodcutting,
			component.ComponentIdFishing,
		); conflicting != "" {
			return fmt.Errorf("dead entity %s has incompatible %s state", entityId, conflicting)
		}
	}

	return nil
}

func firstPresentComponent(
	manager *component.ComponentManager,
	entityId model.EntityId,
	componentIds ...component.ComponentId,
) component.ComponentId {
	for _, componentId := range componentIds {
		if manager.GetEntityComponent(componentId, entityId) != nil {
			return componentId
		}
	}
	return ""
}
