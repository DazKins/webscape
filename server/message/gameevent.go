package message

import "webscape/server/game/model"

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
) Message {
	return newMessage(MessageTypeCombatResolved, combatResolvedData{
		AttackerEntityId: attackerEntityId.String(),
		TargetEntityId:   targetEntityId.String(),
		DidHit:           didHit,
		Damage:           damage,
		IsCritical:       isCritical,
	})
}
