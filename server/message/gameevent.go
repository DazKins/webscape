package message

import (
	"webscape/server/game/model"
	"webscape/server/math"
)

type chatMessageData struct {
	FromEntityId string `json:"fromEntityId"`
	Message      string `json:"message"`
}

type combatResolvedData struct {
	AttackerEntityId string `json:"attackerEntityId"`
	TargetEntityId   string `json:"targetEntityId"`
	DidHit           bool   `json:"didHit"`
	Damage           int    `json:"damage"`
	IsCritical       bool   `json:"isCritical"`
	AttackMethod     string `json:"attackMethod"`
}

type combatProjectilePositionData struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type combatProjectileLaunchedData struct {
	AttackerEntityId string                       `json:"attackerEntityId"`
	TargetEntityId   string                       `json:"targetEntityId"`
	ProjectileType   string                       `json:"projectileType"`
	Origin           combatProjectilePositionData `json:"origin"`
	TargetPosition   combatProjectilePositionData `json:"targetPosition"`
	LaunchTick       uint64                       `json:"launchTick"`
	ImpactTick       uint64                       `json:"impactTick"`
}

func NewChatMessage(fromEntityId model.EntityId, text string) Message {
	return newMessage(MessageTypeChatMessage, chatMessageData{
		FromEntityId: fromEntityId.String(),
		Message:      text,
	})
}

func NewCombatResolvedMessage(
	attackerEntityId model.EntityId,
	targetEntityId model.EntityId,
	didHit bool,
	damage int,
	isCritical bool,
	attackMethods ...model.AttackMethod,
) Message {
	attackMethod := model.AttackMethodMelee
	if len(attackMethods) > 0 && attackMethods[0] != "" {
		attackMethod = attackMethods[0]
	}
	return newMessage(MessageTypeCombatResolved, combatResolvedData{
		AttackerEntityId: attackerEntityId.String(),
		TargetEntityId:   targetEntityId.String(),
		DidHit:           didHit,
		Damage:           damage,
		IsCritical:       isCritical,
		AttackMethod:     string(attackMethod),
	})
}

func NewCombatProjectileLaunchedMessage(
	attackerEntityId model.EntityId,
	targetEntityId model.EntityId,
	projectileType string,
	origin math.Vec2,
	targetPosition math.Vec2,
	launchTick uint64,
	impactTick uint64,
) Message {
	return newMessage(MessageTypeCombatProjectileLaunched, combatProjectileLaunchedData{
		AttackerEntityId: attackerEntityId.String(),
		TargetEntityId:   targetEntityId.String(),
		ProjectileType:   projectileType,
		Origin:           combatProjectilePositionData{X: origin.X, Y: origin.Y},
		TargetPosition:   combatProjectilePositionData{X: targetPosition.X, Y: targetPosition.Y},
		LaunchTick:       launchTick,
		ImpactTick:       impactTick,
	})
}
