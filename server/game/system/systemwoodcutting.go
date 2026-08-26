package system

import (
	"fmt"
	"math/rand"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
)

const woodcuttingAttemptCooldownTicks = 2

type WoodcuttingYieldHandler interface {
	AddItemToPlayerInventory(playerEntityId model.EntityId, item *model.Item) bool
}

type WoodcuttingSystem struct {
	SystemBase
	YieldHandler WoodcuttingYieldHandler
	EventEmitter GameEventEmitter
	RollSource   func() int
}

func (s *WoodcuttingSystem) StartWoodcuttingFor(
	playerEntityId model.EntityId,
	targetEntityId model.EntityId,
) bool {
	woodcuttable := s.woodcuttable(targetEntityId)
	position := s.position(playerEntityId)
	targetPosition := s.position(targetEntityId)
	if woodcuttable == nil || woodcuttable.IsDepleted() || position == nil || targetPosition == nil || !s.isAlive(playerEntityId) {
		return false
	}
	if manhattanDistance(position.GetPosition(), targetPosition.GetPosition()) != 1 {
		return false
	}
	if !s.hasEquippedAxe(playerEntityId) {
		s.addActivityLog(playerEntityId, "You need an equipped axe to chop this tree.", "woodcutting-info")
		return false
	}
	if !s.inventoryCanFit(playerEntityId, woodcuttable.GetYield().Count) {
		s.addActivityLog(playerEntityId, "Your inventory is full.", "woodcutting-info")
		return false
	}

	s.ComponentManager.RemoveComponent(component.ComponentIdCombatState, playerEntityId)
	s.ComponentManager.SetEntityComponent(
		playerEntityId,
		component.NewCWoodcutting(targetEntityId, position.GetPosition()),
	)
	return true
}

func (s *WoodcuttingSystem) Update() {
	s.updateRespawns()
	s.updateWoodcuttingStates()
}

func (s *WoodcuttingSystem) updateRespawns() {
	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdWoodcuttable) {
		woodcuttable := value.(*component.CWoodcuttable)
		if !woodcuttable.IsDepleted() {
			continue
		}

		remaining := woodcuttable.GetRemainingRespawnTicks()
		if remaining > 0 {
			remaining--
			woodcuttable.SetRemainingRespawnTicks(remaining)
		}
		if remaining == 0 {
			woodcuttable.SetCurrentDurability(woodcuttable.GetMaxDurability())
			woodcuttable.SetDepleted(false)
			s.addActivityLog(woodcuttable.GetLastFellerEntityId(), "A tree has regrown.", "woodcutting-info")
			woodcuttable.SetLastFellerEntityId(model.EntityId{})
		}
		s.ComponentManager.SetEntityComponent(entityId, woodcuttable)
	}
}

func (s *WoodcuttingSystem) updateWoodcuttingStates() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdWoodcutting,
		component.ComponentIdPosition,
	)
	for _, playerEntityId := range entityIds {
		s.updateWoodcuttingState(playerEntityId)
	}
}

func (s *WoodcuttingSystem) updateWoodcuttingState(playerEntityId model.EntityId) {
	stateValue := s.ComponentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerEntityId)
	if stateValue == nil {
		return
	}
	state := stateValue.(*component.CWoodcutting)
	position := s.position(playerEntityId)
	targetPosition := s.position(state.GetTargetEntityId())
	woodcuttable := s.woodcuttable(state.GetTargetEntityId())

	if position == nil || targetPosition == nil || woodcuttable == nil || woodcuttable.IsDepleted() ||
		!s.isAlive(playerEntityId) ||
		position.GetPosition() != state.GetStartPosition() ||
		manhattanDistance(position.GetPosition(), targetPosition.GetPosition()) != 1 {
		s.cancelWoodcutting(playerEntityId)
		return
	}
	if !s.hasEquippedAxe(playerEntityId) {
		s.addActivityLog(playerEntityId, "You need an equipped axe to keep chopping.", "woodcutting-info")
		s.cancelWoodcutting(playerEntityId)
		return
	}
	if !s.inventoryCanFit(playerEntityId, woodcuttable.GetYield().Count) {
		s.addActivityLog(playerEntityId, "Your inventory is full.", "woodcutting-info")
		s.cancelWoodcutting(playerEntityId)
		return
	}

	if state.GetCooldownRemaining() > 0 {
		state.SetCooldownRemaining(state.GetCooldownRemaining() - 1)
		if state.GetCooldownRemaining() > 0 {
			return
		}
	}

	damage, description, kind := s.rollOutcome()
	s.emitSwing(playerEntityId, state.GetTargetEntityId())
	if damage == 0 {
		s.addActivityLog(playerEntityId, "You miss the tree.", kind)
		state.SetCooldownRemaining(woodcuttingAttemptCooldownTicks)
		return
	}

	remaining := woodcuttable.GetCurrentDurability() - damage
	s.addActivityLog(playerEntityId, fmt.Sprintf("You make a %s chop for %d damage.", description, damage), kind)
	if remaining > 0 {
		woodcuttable.SetCurrentDurability(remaining)
		s.ComponentManager.SetEntityComponent(state.GetTargetEntityId(), woodcuttable)
		state.SetCooldownRemaining(woodcuttingAttemptCooldownTicks)
		return
	}

	yield := woodcuttable.GetYield()
	if s.YieldHandler == nil || !s.addYield(playerEntityId, yield) {
		s.addActivityLog(playerEntityId, "Your inventory is full.", "woodcutting-info")
		s.cancelWoodcutting(playerEntityId)
		return
	}

	woodcuttable.SetCurrentDurability(0)
	woodcuttable.SetDepleted(true)
	woodcuttable.SetRemainingRespawnTicks(woodcuttable.GetRespawnTicks())
	woodcuttable.SetLastFellerEntityId(playerEntityId)
	s.ComponentManager.SetEntityComponent(state.GetTargetEntityId(), woodcuttable)
	s.addActivityLog(playerEntityId, yieldReceivedMessage(yield), "woodcutting-reward")
	s.cancelWoodcuttersTargeting(state.GetTargetEntityId())
}

func (s *WoodcuttingSystem) emitSwing(playerEntityId model.EntityId, targetEntityId model.EntityId) {
	if s.EventEmitter == nil {
		return
	}
	s.EventEmitter.EmitGameEvent(gameevent.NewWoodcuttingSwing(playerEntityId, targetEntityId))
}

func (s *WoodcuttingSystem) addYield(playerEntityId model.EntityId, yield component.LootItem) bool {
	for i := 0; i < yield.Count; i++ {
		if !s.YieldHandler.AddItemToPlayerInventory(playerEntityId, yield.CreateItem()) {
			return false
		}
	}
	return true
}

func yieldReceivedMessage(yield component.LootItem) string {
	if yield.Count == 1 {
		return fmt.Sprintf("You receive 1 %s.", yield.Name)
	}
	return fmt.Sprintf("You receive %d %s.", yield.Count, yield.Name)
}

func (s *WoodcuttingSystem) rollOutcome() (damage int, description string, kind string) {
	roll := rand.Intn(100)
	if s.RollSource != nil {
		roll = s.RollSource()
	}
	if roll < 0 {
		roll = 0
	}
	if roll > 99 {
		roll = 99
	}
	switch {
	case roll < 25:
		return 0, "missed", "woodcutting-miss"
	case roll < 75:
		return 1, "bad", "woodcutting-bad"
	default:
		return 2, "good", "woodcutting-good"
	}
}

func (s *WoodcuttingSystem) hasEquippedAxe(entityId model.EntityId) bool {
	equipped := s.ComponentManager.GetEntityComponent(component.ComponentIdEquipped, entityId)
	if equipped == nil {
		return false
	}
	item := equipped.(*component.CEquipped).GetEquippedItem(model.SlotWeapon)
	return item != nil && item.Type == "axe"
}

func (s *WoodcuttingSystem) isAlive(entityId model.EntityId) bool {
	health := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, entityId)
	return health != nil && health.(*component.CHealth).GetCurrentHealth() > 0
}

func (s *WoodcuttingSystem) inventoryCanFit(entityId model.EntityId, count int) bool {
	inventory := s.ComponentManager.GetEntityComponent(component.ComponentIdInventory, entityId)
	return inventory != nil && count > 0 && inventory.(*component.CInventory).AvailableSlots() >= count
}

func (s *WoodcuttingSystem) woodcuttable(entityId model.EntityId) *component.CWoodcuttable {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdWoodcuttable, entityId)
	if value == nil {
		return nil
	}
	return value.(*component.CWoodcuttable)
}

func (s *WoodcuttingSystem) position(entityId model.EntityId) *component.CPosition {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId)
	if value == nil {
		return nil
	}
	return value.(*component.CPosition)
}

func (s *WoodcuttingSystem) cancelWoodcutting(entityId model.EntityId) {
	s.ComponentManager.RemoveComponent(component.ComponentIdWoodcutting, entityId)
}

func (s *WoodcuttingSystem) cancelWoodcuttersTargeting(targetEntityId model.EntityId) {
	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdWoodcutting) {
		if value.(*component.CWoodcutting).GetTargetEntityId() == targetEntityId {
			s.cancelWoodcutting(entityId)
		}
	}
}

func (s *WoodcuttingSystem) addActivityLog(entityId model.EntityId, text string, kind string) {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, entityId)
	if value == nil {
		return
	}
	activityLog := value.(*component.CCombatLog)
	activityLog.AddEntry(component.NewCombatLogEntry(text, kind))
	s.ComponentManager.SetEntityComponent(entityId, activityLog)
}
