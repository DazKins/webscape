package system

import (
	"fmt"
	"math/rand"
	"webscape/server/game/collision"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

type CombatSystem struct {
	SystemBase
	World          *world.World
	SpatialIndex   SpatialCandidates
	EventEmitter   GameEventEmitter
	TickSource     TickSource
	pendingImpacts []pendingCombatImpact
}

type pendingCombatImpact struct {
	attackerId     model.EntityId
	targetId       model.EntityId
	attackerName   string
	minDamage      int
	maxDamage      int
	critChance     float64
	critMultiplier float64
	attackMethod   model.AttackMethod
	impactTick     uint64
}

func (s *CombatSystem) Update() {
	s.resolvePendingImpacts()
	s.updateCombatStates()
}

func (s *CombatSystem) updateCombatStates() {
	combatEntities := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdCombatState,
		component.ComponentIdPosition,
		component.ComponentIdHealth,
	)

	for _, attackerId := range combatEntities {
		combatState := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attackerId).(*component.CCombatState)
		targetId := combatState.GetTargetId()
		attackerHealth := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, attackerId).(*component.CHealth)
		targetHealth := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, targetId)
		targetPositionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetId)
		if attackerHealth.GetCurrentHealth() <= 0 || targetHealth == nil ||
			targetHealth.(*component.CHealth).GetCurrentHealth() <= 0 || targetPositionComponent == nil {
			s.clearCombatState(attackerId)
			continue
		}

		attackerPosition := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, attackerId).(*component.CPosition).GetPosition()
		targetPosition := targetPositionComponent.(*component.CPosition).GetPosition()

		attackerStats := s.ensureCombatStats(attackerId)
		if attackerStats == nil {
			continue
		}
		currentTick := currentTick(s.TickSource)
		if !combatState.ProfileMatches(attackerStats) {
			combatState.Restart(attackerStats, currentTick)
			s.ComponentManager.SetEntityComponent(attackerId, combatState)
		}
		if !s.hasProjectileAmmo(attackerId, attackerStats.GetAttackMethod()) {
			s.stopRangedCombatWithoutArrows(attackerId)
			continue
		}

		if combatState.GetPhase() == component.CombatPhaseCasting {
			if currentTick >= combatState.GetPhaseStartedTick()+uint64(combatState.GetWindUpTicks()) {
				if !s.launchProjectileAttack(attackerId, targetId, attackerPosition, targetPosition, attackerStats) {
					s.stopRangedCombatWithoutArrows(attackerId)
					continue
				}
				combatState.BeginRecovering(currentTick, combatState.GetNextAttackTick())
				s.ComponentManager.SetEntityComponent(attackerId, combatState)
			}
			continue
		}

		if combatState.GetPhase() == component.CombatPhaseRecovering {
			if currentTick < combatState.GetNextAttackTick() {
				continue
			}
			combatState.BeginApproaching(currentTick)
			s.ComponentManager.SetEntityComponent(attackerId, combatState)
		}

		if manhattanDistance(attackerPosition, targetPosition) > attackerStats.GetAttackRange() {
			s.setPathingToEntity(attackerId, targetId)
			continue
		}

		if manhattanDistance(attackerPosition, targetPosition) == 0 {
			if s.stepOutFromTarget(attackerId, targetPosition) {
				continue
			}
		}

		if isProjectileAttack(attackerStats.GetAttackMethod()) {
			combatState.BeginCasting(
				currentTick,
				currentTick+uint64(attackerStats.GetAttackSpeedTicks()),
			)
			s.ComponentManager.SetEntityComponent(attackerId, combatState)
			continue
		}

		targetStats := s.ensureCombatStats(targetId)
		if targetStats == nil {
			continue
		}
		attackResult := s.resolveAttack(attackerId, targetId, attackerStats, targetStats)
		if s.applyAttackResult(attackerId, targetId, attackResult) {
			s.discardPendingImpactsAgainst(targetId)
		}
		combatState.BeginRecovering(
			currentTick,
			currentTick+uint64(attackerStats.GetAttackSpeedTicks()),
		)
		s.ComponentManager.SetEntityComponent(attackerId, combatState)
	}
}

func (s *CombatSystem) launchProjectileAttack(
	attackerId model.EntityId,
	targetId model.EntityId,
	origin math.Vec2,
	targetPosition math.Vec2,
	attackerStats *component.CCombatStats,
) bool {
	if !s.consumeProjectileAmmo(attackerId, attackerStats.GetAttackMethod()) {
		return false
	}
	launchTick := currentTick(s.TickSource)
	impactTick := launchTick + uint64(attackerStats.GetTravelTicks())
	s.pendingImpacts = append(s.pendingImpacts, pendingCombatImpact{
		attackerId:     attackerId,
		targetId:       targetId,
		attackerName:   s.getEntityName(attackerId),
		minDamage:      attackerStats.GetMinDamage(),
		maxDamage:      attackerStats.GetMaxDamage(),
		critChance:     attackerStats.GetCritChance(),
		critMultiplier: attackerStats.GetCritMultiplier(),
		attackMethod:   attackerStats.GetAttackMethod(),
		impactTick:     impactTick,
	})
	if s.EventEmitter != nil {
		s.EventEmitter.EmitGameEvent(gameevent.NewCombatProjectileLaunched(
			attackerId,
			targetId,
			attackerStats.GetProjectileType(),
			origin,
			targetPosition,
			launchTick,
			impactTick,
		))
	}
	return true
}

func isProjectileAttack(attackMethod model.AttackMethod) bool {
	return attackMethod == model.AttackMethodMagic || attackMethod == model.AttackMethodRanged
}

func (s *CombatSystem) hasProjectileAmmo(attackerId model.EntityId, attackMethod model.AttackMethod) bool {
	if attackMethod != model.AttackMethodRanged {
		return true
	}
	inventory := s.ComponentManager.GetEntityComponent(component.ComponentIdInventory, attackerId)
	return inventory != nil && inventory.(*component.CInventory).HasItemType(model.ItemTypeArrow)
}

func (s *CombatSystem) consumeProjectileAmmo(attackerId model.EntityId, attackMethod model.AttackMethod) bool {
	if attackMethod != model.AttackMethodRanged {
		return true
	}
	inventory := s.ComponentManager.GetEntityComponent(component.ComponentIdInventory, attackerId)
	if inventory == nil {
		return false
	}
	inventoryComponent := inventory.(*component.CInventory)
	if inventoryComponent.RemoveFirstItemByType(model.ItemTypeArrow) == nil {
		return false
	}
	s.ComponentManager.SetEntityComponent(attackerId, inventoryComponent)
	return true
}

func (s *CombatSystem) stopRangedCombatWithoutArrows(attackerId model.EntityId) {
	s.clearCombatState(attackerId)
	s.addCombatLog(attackerId, "You have no arrows left", "miss")
}

func (s *CombatSystem) resolvePendingImpacts() {
	if len(s.pendingImpacts) == 0 {
		return
	}
	now := currentTick(s.TickSource)
	remaining := make([]pendingCombatImpact, 0, len(s.pendingImpacts))
	killedTargets := make(map[model.EntityId]bool)
	for _, impact := range s.pendingImpacts {
		if killedTargets[impact.targetId] {
			continue
		}
		if impact.impactTick > now {
			remaining = append(remaining, impact)
			continue
		}
		healthValue := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, impact.targetId)
		if healthValue == nil || healthValue.(*component.CHealth).GetCurrentHealth() <= 0 {
			continue
		}
		targetStats := s.ensureCombatStats(impact.targetId)
		if targetStats == nil {
			continue
		}
		result := s.resolveGuaranteedAttack(impact, targetStats)
		if s.applyAttackResult(impact.attackerId, impact.targetId, result) {
			killedTargets[impact.targetId] = true
		}
	}
	if len(killedTargets) > 0 {
		remaining = filterImpactsNotTargeting(remaining, killedTargets)
	}
	s.pendingImpacts = remaining
}

func filterImpactsNotTargeting(
	impacts []pendingCombatImpact,
	targets map[model.EntityId]bool,
) []pendingCombatImpact {
	result := impacts[:0]
	for _, impact := range impacts {
		if !targets[impact.targetId] {
			result = append(result, impact)
		}
	}
	return result
}

func (s *CombatSystem) discardPendingImpactsAgainst(targetId model.EntityId) {
	s.pendingImpacts = filterImpactsNotTargeting(
		s.pendingImpacts,
		map[model.EntityId]bool{targetId: true},
	)
}

func (s *CombatSystem) HandleEntityDeath(entityId model.EntityId) {
	s.discardPendingImpactsAgainst(entityId)
}

func (s *CombatSystem) emitKillEvents(attackerId model.EntityId, targetId model.EntityId) {
	if s.EventEmitter == nil {
		return
	}
	if s.ComponentManager.GetEntityComponent(component.ComponentIdPlayer, attackerId) == nil {
		return
	}

	if entityType := s.getMetadataString(targetId, "entityType"); entityType != "" {
		event := gameevent.New("kill:entity:"+gameevent.NormalizeToken(entityType), attackerId)
		event.TargetEntityId = targetId
		s.EventEmitter.EmitGameEvent(event)
	}
	if name := s.getMetadataString(targetId, "name"); name != "" {
		event := gameevent.New("kill:name:"+gameevent.NormalizeToken(name), attackerId)
		event.TargetEntityId = targetId
		s.EventEmitter.EmitGameEvent(event)
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
		s.emitCombatResolved(attackerId, targetId, attackResult{DidHit: false}, model.AttackMethodMelee)
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

	kind := "hit"
	if isCrit {
		kind = "crit"
	}
	s.emitCombatResolved(attackerId, targetId, attackResult{DidHit: true, Damage: damage, IsCrit: isCrit}, model.AttackMethodMelee)
	s.addCombatLog(attackerId, fmt.Sprintf("You hit %s for %d", targetName, damage), kind)
	s.addCombatLog(targetId, fmt.Sprintf("%s hits you for %d", attackerName, damage), kind)

	return attackResult{DidHit: true, Damage: damage, IsCrit: isCrit}
}

func (s *CombatSystem) resolveGuaranteedAttack(
	impact pendingCombatImpact,
	targetStats *component.CCombatStats,
) attackResult {
	minDamage := impact.minDamage
	maxDamage := impact.maxDamage
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	damage := minDamage + rand.Intn(maxDamage-minDamage+1)
	damage -= targetStats.GetArmor()
	if damage < 1 {
		damage = 1
	}
	isCrit := rand.Float64() < impact.critChance
	if isCrit {
		damage = int(float64(damage) * impact.critMultiplier)
	}
	result := attackResult{DidHit: true, Damage: damage, IsCrit: isCrit}
	s.emitCombatResolved(impact.attackerId, impact.targetId, result, impact.attackMethod)
	kind := "hit"
	if isCrit {
		kind = "crit"
	}
	s.addCombatLog(impact.attackerId, fmt.Sprintf("You hit %s for %d", s.getEntityName(impact.targetId), damage), kind)
	s.addCombatLog(impact.targetId, fmt.Sprintf("%s hits you for %d", impact.attackerName, damage), kind)
	return result
}

func (s *CombatSystem) applyAttackResult(
	attackerId model.EntityId,
	targetId model.EntityId,
	result attackResult,
) bool {
	if !result.DidHit {
		return false
	}
	healthValue := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, targetId)
	if healthValue == nil {
		return false
	}
	targetHealth := healthValue.(*component.CHealth)
	previousHealth := targetHealth.GetCurrentHealth()
	newHealth := max(0, previousHealth-result.Damage)
	targetHealth.SetCurrentHealth(newHealth)
	s.ComponentManager.SetEntityComponent(targetId, targetHealth)
	if previousHealth > 0 && newHealth == 0 {
		s.emitKillEvents(attackerId, targetId)
		return true
	}
	return false
}

func (s *CombatSystem) emitCombatResolved(
	attackerId model.EntityId,
	targetId model.EntityId,
	result attackResult,
	attackMethods ...model.AttackMethod,
) {
	if s.EventEmitter == nil {
		return
	}
	attackMethod := model.AttackMethodMelee
	if len(attackMethods) > 0 && attackMethods[0] != "" {
		attackMethod = attackMethods[0]
	}
	s.EventEmitter.EmitGameEvent(gameevent.NewCombatResolved(
		attackerId,
		targetId,
		result.DidHit,
		result.Damage,
		result.IsCrit,
		attackMethod,
	))
}

func (s *CombatSystem) addCombatLog(entityId model.EntityId, text string, kind string) {
	combatLog := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, entityId)
	if combatLog == nil {
		return
	}
	combatLogComponent := combatLog.(*component.CCombatLog)
	combatLogComponent.AddEntry(component.NewCombatLogEntry(text, kind))
	s.ComponentManager.SetEntityComponent(entityId, combatLogComponent)
}

func (s *CombatSystem) getEntityName(entityId model.EntityId) string {
	name := s.getMetadataString(entityId, "name")
	if name != "" {
		return name
	}
	return "Unknown"
}

func (s *CombatSystem) getMetadataString(entityId model.EntityId, key string) string {
	metadata := s.ComponentManager.GetEntityComponent(component.ComponentIdMetadata, entityId)
	if metadata == nil {
		return ""
	}
	metadataComponent := metadata.(*component.CMetadata)
	metadataObject, ok := metadataComponent.GetMetadata().(util.JObject)
	if !ok {
		return ""
	}
	value, ok := metadataObject[key].(util.JString)
	if !ok {
		return ""
	}
	return string(value)
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
	s.entityStateTransitions(s.TickSource).EndCombat(entityId)
}

func (s *CombatSystem) setPathingToEntity(entityId model.EntityId, targetId model.EntityId) {
	if value := s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, entityId); value != nil {
		target := value.(*component.CPathing).GetTarget()
		if target.EntityId.IsPresent() && target.EntityId.Unwrap() == targetId {
			return
		}
	}
	pathingComponent := component.NewCPathing(component.PathingTarget{
		EntityId: util.OptionalSome(targetId),
	})
	s.ComponentManager.SetEntityComponent(entityId, pathingComponent)
}

func (s *CombatSystem) stepOutFromTarget(attackerId model.EntityId, targetPosition math.Vec2) bool {
	directions := []math.Vec2{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
	}

	for _, direction := range directions {
		candidate := targetPosition.Add(direction)
		if s.collision().IsBlocked(candidate.X, candidate.Y) {
			continue
		}
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, attackerId).(*component.CPosition)
		s.entityStateTransitions(s.TickSource).BeginMoving(attackerId)
		positionComponent.SetPosition(candidate)
		s.ComponentManager.SetEntityComponent(attackerId, positionComponent)
		return true
	}
	return false
}

func (s *CombatSystem) collision() collision.Checker {
	return collision.Checker{
		World:            s.World,
		ComponentManager: s.ComponentManager,
		SpatialIndex:     s.SpatialIndex,
	}
}

func toEquipped(componentValue component.Component) *component.CEquipped {
	if componentValue == nil {
		return nil
	}
	return componentValue.(*component.CEquipped)
}
