package system

import (
	"fmt"
	"math/rand"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
)

type FishingYieldHandler interface {
	AddItemToPlayerInventory(playerEntityId model.EntityId, item *model.Item) bool
}

type FishingSystem struct {
	SystemBase
	YieldHandler FishingYieldHandler
	EventEmitter GameEventEmitter
	TickSource   TickSource
	RollSource   func() int
}

func (s *FishingSystem) StartFishingFor(
	playerEntityId model.EntityId,
	targetEntityId model.EntityId,
) bool {
	fishable := s.fishable(targetEntityId)
	position := s.position(playerEntityId)
	targetPosition := s.position(targetEntityId)
	if fishable == nil || position == nil || targetPosition == nil || !s.isAlive(playerEntityId) {
		return false
	}
	if manhattanDistance(position.GetPosition(), targetPosition.GetPosition()) != 1 {
		return false
	}
	if !s.hasEquippedFishingRod(playerEntityId) {
		s.addActivityLog(playerEntityId, "You need an equipped fishing rod to fish here.", "fishing-info")
		return false
	}
	if !s.inventoryCanFit(playerEntityId, fishable.GetYield().Count) {
		s.addActivityLog(playerEntityId, "Your inventory is full. Fishing canceled.", "fishing-info")
		return false
	}
	tick := s.currentTick()
	return s.entityStateTransitions(s.TickSource).BeginFishing(
		playerEntityId,
		component.NewCFishing(targetEntityId, tick, position.GetPosition()),
	)
}

func (s *FishingSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdFishing,
		component.ComponentIdPosition,
	)
	for _, playerEntityId := range entityIds {
		state := s.fishing(playerEntityId)
		if state != nil {
			s.updateFishingState(playerEntityId, state)
		}
	}
}

func (s *FishingSystem) updateFishingState(playerEntityId model.EntityId, state *component.CFishing) {
	position := s.position(playerEntityId)
	targetPosition := s.position(state.GetTargetEntityId())
	fishable := s.fishable(state.GetTargetEntityId())
	if position == nil || targetPosition == nil || fishable == nil || !s.isAlive(playerEntityId) ||
		position.GetPosition() != state.GetOriginPosition() ||
		manhattanDistance(position.GetPosition(), targetPosition.GetPosition()) != 1 ||
		s.hasConflictingActivity(playerEntityId) {
		s.cancelFishing(playerEntityId)
		return
	}
	if !s.hasEquippedFishingRod(playerEntityId) {
		s.addActivityLog(playerEntityId, "You need an equipped fishing rod to keep fishing.", "fishing-info")
		s.cancelFishing(playerEntityId)
		return
	}
	yield := fishable.GetYield()
	if state.GetPhase() != component.FishingPhaseReeling &&
		!s.inventoryCanFit(playerEntityId, yield.Count) {
		s.cancelForFullInventory(playerEntityId)
		return
	}
	tick := s.currentTick()
	switch state.GetPhase() {
	case component.FishingPhaseCasting:
		if tick > state.GetPhaseStartedTick() {
			s.startPhase(playerEntityId, state, component.FishingPhaseWaiting, tick)
		}
	case component.FishingPhaseWaiting:
		if tick <= state.GetPhaseStartedTick() || s.roll() >= fishable.GetCatchChancePercent() {
			return
		}
		if s.YieldHandler == nil || !s.addYield(playerEntityId, yield) {
			s.cancelForFullInventory(playerEntityId)
			return
		}

		s.addActivityLog(playerEntityId, fishingCatchMessage(yield), "fishing-reward")
		s.startPhase(playerEntityId, state, component.FishingPhaseReeling, tick)
		s.emitCatch(playerEntityId, state.GetTargetEntityId(), yield.Count)
	case component.FishingPhaseReeling:
		if tick <= state.GetPhaseStartedTick() {
			return
		}
		if !s.inventoryCanFit(playerEntityId, fishable.GetYield().Count) {
			s.cancelForFullInventory(playerEntityId)
			return
		}
		s.startPhase(playerEntityId, state, component.FishingPhaseWaiting, tick)
	default:
		s.cancelFishing(playerEntityId)
	}
}

func (s *FishingSystem) startPhase(
	playerEntityId model.EntityId,
	state *component.CFishing,
	phase component.FishingPhase,
	tick uint64,
) {
	state.StartPhase(phase, tick)
	s.ComponentManager.SetEntityComponent(playerEntityId, state)
}

func (s *FishingSystem) cancelForFullInventory(playerEntityId model.EntityId) {
	s.addActivityLog(playerEntityId, "Your inventory is full. Fishing canceled.", "fishing-info")
	s.cancelFishing(playerEntityId)
}

func (s *FishingSystem) roll() int {
	roll := rand.Intn(100)
	if s.RollSource != nil {
		roll = s.RollSource()
	}
	if roll < 0 {
		return 0
	}
	if roll > 99 {
		return 99
	}
	return roll
}

func (s *FishingSystem) addYield(playerEntityId model.EntityId, yield component.LootItem) bool {
	for i := 0; i < yield.Count; i++ {
		if !s.YieldHandler.AddItemToPlayerInventory(playerEntityId, yield.CreateItem()) {
			return false
		}
	}
	return true
}

func (s *FishingSystem) emitCatch(playerEntityId model.EntityId, targetEntityId model.EntityId, count int) {
	if s.EventEmitter != nil {
		s.EventEmitter.EmitGameEvent(gameevent.NewFishingCatch(playerEntityId, targetEntityId, count))
	}
}

func (s *FishingSystem) hasConflictingActivity(entityId model.EntityId) bool {
	return s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, entityId) != nil ||
		s.ComponentManager.GetEntityComponent(component.ComponentIdInteracting, entityId) != nil ||
		s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, entityId) != nil ||
		s.ComponentManager.GetEntityComponent(component.ComponentIdWoodcutting, entityId) != nil
}

func (s *FishingSystem) fishing(entityId model.EntityId) *component.CFishing {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdFishing, entityId)
	if value == nil {
		return nil
	}
	return value.(*component.CFishing)
}

func (s *FishingSystem) currentTick() uint64 {
	return currentTick(s.TickSource)
}

func (s *FishingSystem) hasEquippedFishingRod(entityId model.EntityId) bool {
	equipped := s.ComponentManager.GetEntityComponent(component.ComponentIdEquipped, entityId)
	if equipped == nil {
		return false
	}
	item := equipped.(*component.CEquipped).GetEquippedItem(model.SlotWeapon)
	return item != nil && item.Type == "fishingRod"
}

func (s *FishingSystem) isAlive(entityId model.EntityId) bool {
	health := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, entityId)
	return health != nil && health.(*component.CHealth).GetCurrentHealth() > 0
}

func (s *FishingSystem) inventoryCanFit(entityId model.EntityId, count int) bool {
	inventory := s.ComponentManager.GetEntityComponent(component.ComponentIdInventory, entityId)
	return inventory != nil && count > 0 && inventory.(*component.CInventory).AvailableSlots() >= count
}

func (s *FishingSystem) fishable(entityId model.EntityId) *component.CFishable {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdFishable, entityId)
	if value == nil {
		return nil
	}
	return value.(*component.CFishable)
}

func (s *FishingSystem) position(entityId model.EntityId) *component.CPosition {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId)
	if value == nil {
		return nil
	}
	return value.(*component.CPosition)
}

func (s *FishingSystem) cancelFishing(entityId model.EntityId) {
	s.ComponentManager.RemoveComponent(component.ComponentIdFishing, entityId)
}

func (s *FishingSystem) addActivityLog(entityId model.EntityId, text string, kind string) {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, entityId)
	if value == nil {
		return
	}
	activityLog := value.(*component.CCombatLog)
	activityLog.AddEntry(component.NewCombatLogEntry(text, kind))
	s.ComponentManager.SetEntityComponent(entityId, activityLog)
}

func fishingCatchMessage(yield component.LootItem) string {
	if yield.Count == 1 {
		return fmt.Sprintf("You catch 1 %s.", yield.Name)
	}
	return fmt.Sprintf("You catch %d %s.", yield.Count, yield.Name)
}
