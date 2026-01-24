package system

import (
	"fmt"
	"math/rand"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

const (
	combatTextTtlTicks = 4
)

type CombatSystem struct {
	SystemBase
	World        *world.World
	TickCounter  int
	entryCounter int
}

func (s *CombatSystem) Update() {
	s.TickCounter++
	s.updateCombatAI()
	s.updateCombatStates()
}

func (s *CombatSystem) updateCombatAI() {
	aiEntities := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdCombatAI,
		component.ComponentIdPosition,
	)
	playerEntities := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdPlayer,
		component.ComponentIdPosition,
	)

	for _, entityId := range aiEntities {
		aiComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatAI, entityId).(*component.CCombatAI)
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
		combatState := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, entityId)

		if combatState != nil {
			state := combatState.(*component.CCombatState)
			targetId := state.GetTargetId()
			targetPositionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetId)
			if targetPositionComponent == nil {
				s.clearCombatState(entityId)
				continue
			}

			targetPosition := targetPositionComponent.(*component.CPosition).GetPosition()
			if manhattanDistance(positionComponent.GetPosition(), targetPosition) > aiComponent.GetLeashRadius() {
				s.clearCombatState(entityId)
				s.setPathingToPosition(entityId, aiComponent.GetHomePosition())
			}
			continue
		}

		var closestPlayer model.EntityId
		closestDistance := aiComponent.GetAggroRadius() + 1
		foundTarget := false
		for _, playerId := range playerEntities {
			playerPosition := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, playerId).(*component.CPosition).GetPosition()
			distance := manhattanDistance(positionComponent.GetPosition(), playerPosition)
			if distance <= aiComponent.GetAggroRadius() && distance < closestDistance {
				closestDistance = distance
				closestPlayer = playerId
				foundTarget = true
			}
		}

		if foundTarget {
			s.ComponentManager.SetEntityComponent(entityId, component.NewCCombatState(closestPlayer))
			s.setPathingToEntity(entityId, closestPlayer)
		}
	}
}

func (s *CombatSystem) updateCombatStates() {
	combatEntities := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdCombatState,
		component.ComponentIdPosition,
		component.ComponentIdHealth,
	)

	for _, attackerId := range combatEntities {
		combatState := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attackerId).(*component.CCombatState)
		if combatState.GetCooldownRemaining() > 0 {
			combatState.SetCooldownRemaining(combatState.GetCooldownRemaining() - 1)
		}

		targetId := combatState.GetTargetId()
		targetHealth := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, targetId)
		targetPositionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetId)
		if targetHealth == nil || targetPositionComponent == nil {
			s.clearCombatState(attackerId)
			continue
		}

		attackerPosition := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, attackerId).(*component.CPosition).GetPosition()
		targetPosition := targetPositionComponent.(*component.CPosition).GetPosition()

		attackerStats := s.ensureCombatStats(attackerId)
		if attackerStats == nil {
			continue
		}
		targetStats := s.ensureCombatStats(targetId)
		if targetStats == nil {
			continue
		}

		if manhattanDistance(attackerPosition, targetPosition) > attackerStats.GetAttackRange() {
			s.setPathingToEntity(attackerId, targetId)
			continue
		}

		if combatState.GetCooldownRemaining() > 0 {
			continue
		}

		attackResult := s.resolveAttack(attackerId, targetId, attackerStats, targetStats)
		if attackResult.DidHit {
			targetHealthComponent := targetHealth.(*component.CHealth)
			newHealth := targetHealthComponent.GetCurrentHealth() - attackResult.Damage
			if newHealth < 0 {
				newHealth = 0
			}
			targetHealthComponent.SetCurrentHealth(newHealth)
			s.ComponentManager.SetEntityComponent(targetId, targetHealthComponent)
		}

		combatState.SetCooldownRemaining(attackerStats.GetAttackSpeedTicks())
		combatState.SetLastAttackTick(s.TickCounter)
	}
}

type attackResult struct {
	DidHit bool
	Damage int
	IsCrit bool
}

func (s *CombatSystem) resolveAttack(
	attackerId model.EntityId,
	targetId model.EntityId,
	attackerStats *component.CCombatStats,
	targetStats *component.CCombatStats,
) attackResult {
	attackerName := s.getEntityName(attackerId)
	targetName := s.getEntityName(targetId)

	hitChance := 70 + attackerStats.GetAccuracy() - targetStats.GetEvasion()
	if hitChance < 5 {
		hitChance = 5
	}
	if hitChance > 95 {
		hitChance = 95
	}

	if rand.Intn(100) >= hitChance {
		s.emitCombatText(targetId, "MISS", "miss")
		s.addCombatLog(attackerId, fmt.Sprintf("You miss %s", targetName), "miss")
		s.addCombatLog(targetId, fmt.Sprintf("%s misses you", attackerName), "miss")
		return attackResult{DidHit: false, Damage: 0, IsCrit: false}
	}

	minDamage := attackerStats.GetMinDamage()
	maxDamage := attackerStats.GetMaxDamage()
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	damage := minDamage + rand.Intn(maxDamage-minDamage+1)
	damage -= targetStats.GetArmor()
	if damage < 1 {
		damage = 1
	}

	isCrit := rand.Float64() < attackerStats.GetCritChance()
	if isCrit {
		damage = int(float64(damage) * attackerStats.GetCritMultiplier())
	}

	text := fmt.Sprintf("%d", damage)
	kind := "hit"
	if isCrit {
		text = fmt.Sprintf("CRIT %d", damage)
		kind = "crit"
	}
	s.emitCombatText(targetId, text, kind)
	s.addCombatLog(attackerId, fmt.Sprintf("You hit %s for %d", targetName, damage), kind)
	s.addCombatLog(targetId, fmt.Sprintf("%s hits you for %d", attackerName, damage), kind)

	return attackResult{DidHit: true, Damage: damage, IsCrit: isCrit}
}

func (s *CombatSystem) emitCombatText(targetId model.EntityId, text string, kind string) {
	renderable := component.NewCRenderable("combattext")
	combatText := component.NewCCombatText(targetId, text, kind)
	ttl := component.NewCTtl(combatTextTtlTicks)
	s.ComponentManager.CreateNewEntity(renderable, combatText, ttl)
}

func (s *CombatSystem) addCombatLog(entityId model.EntityId, text string, kind string) {
	combatLog := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, entityId)
	if combatLog == nil {
		return
	}
	s.entryCounter++
	combatLogComponent := combatLog.(*component.CCombatLog)
	combatLogComponent.AddEntry(component.NewCombatLogEntry(text, kind))
	s.ComponentManager.SetEntityComponent(entityId, combatLogComponent)
}

func (s *CombatSystem) getEntityName(entityId model.EntityId) string {
	metadata := s.ComponentManager.GetEntityComponent(component.ComponentIdMetadata, entityId)
	if metadata == nil {
		return "Unknown"
	}
	metadataComponent := metadata.(*component.CMetadata)
	metadataObject, ok := metadataComponent.GetMetadata().(util.JObject)
	if !ok {
		return "Unknown"
	}
	nameValue, ok := metadataObject["name"].(util.JString)
	if !ok {
		return "Unknown"
	}
	return string(nameValue)
}

func (s *CombatSystem) ensureCombatStats(entityId model.EntityId) *component.CCombatStats {
	stats := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatStats, entityId)
	if stats != nil {
		return stats.(*component.CCombatStats)
	}

	baseStats := s.ComponentManager.GetEntityComponent(component.ComponentIdBaseStats, entityId)
	equipped := s.ComponentManager.GetEntityComponent(component.ComponentIdEquipped, entityId)
	if baseStats == nil {
		return nil
	}

	computed := component.CalculateCombatStats(baseStats.(*component.CBaseStats), toEquipped(equipped))
	s.ComponentManager.SetEntityComponent(entityId, computed)
	return computed
}

func (s *CombatSystem) clearCombatState(entityId model.EntityId) {
	s.ComponentManager.RemoveComponent(component.ComponentIdCombatState, entityId)
	s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
	s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
}

func (s *CombatSystem) setPathingToEntity(entityId model.EntityId, targetId model.EntityId) {
	pathingComponent := component.NewCPathing(component.PathingTarget{
		EntityId: util.OptionalSome(targetId),
	})
	s.ComponentManager.SetEntityComponent(entityId, pathingComponent)
}

func (s *CombatSystem) setPathingToPosition(entityId model.EntityId, target math.Vec2) {
	pathingComponent := component.NewCPathing(component.PathingTarget{
		Position: util.OptionalSome(target),
	})
	s.ComponentManager.SetEntityComponent(entityId, pathingComponent)
}

func manhattanDistance(a math.Vec2, b math.Vec2) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func toEquipped(componentValue component.Component) *component.CEquipped {
	if componentValue == nil {
		return nil
	}
	return componentValue.(*component.CEquipped)
}
