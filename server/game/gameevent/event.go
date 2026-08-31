package gameevent

import (
	"strings"
	"unicode"
	"webscape/server/game/model"
)

type Event struct {
	Id             string
	ActorEntityId  model.EntityId
	TargetEntityId model.EntityId
	Count          int
	Metadata       map[string]string
	Payload        any
}

const (
	EventIdChatSpoken       = "chat:spoken"
	EventIdCombatResolved   = "combat:resolved"
	EventIdWoodcuttingSwing = "woodcutting:swing"
	EventIdFishingCatch     = "fishing:catch"
)

type ChatSpokenPayload struct {
	Message string
}

type CombatResolvedPayload struct {
	DidHit     bool
	Damage     int
	IsCritical bool
}

type WoodcuttingSwingPayload struct{}

func New(id string, actorEntityId model.EntityId) Event {
	return Event{
		Id:            id,
		ActorEntityId: actorEntityId,
		Count:         1,
		Metadata:      map[string]string{},
	}
}

func NewChatSpoken(actorEntityId model.EntityId, message string) Event {
	event := New(EventIdChatSpoken, actorEntityId)
	event.Payload = ChatSpokenPayload{Message: message}
	return event
}

func NewCombatResolved(
	attackerEntityId model.EntityId,
	targetEntityId model.EntityId,
	didHit bool,
	damage int,
	isCritical bool,
) Event {
	event := New(EventIdCombatResolved, attackerEntityId)
	event.TargetEntityId = targetEntityId
	event.Payload = CombatResolvedPayload{
		DidHit:     didHit,
		Damage:     damage,
		IsCritical: isCritical,
	}
	return event
}

func NewWoodcuttingSwing(playerEntityId model.EntityId, targetEntityId model.EntityId) Event {
	event := New(EventIdWoodcuttingSwing, playerEntityId)
	event.TargetEntityId = targetEntityId
	event.Payload = WoodcuttingSwingPayload{}
	return event
}

func NewFishingCatch(playerEntityId model.EntityId, targetEntityId model.EntityId, count int) Event {
	event := New(EventIdFishingCatch, playerEntityId)
	event.TargetEntityId = targetEntityId
	event.Count = count
	return event
}

func NormalizeToken(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var builder strings.Builder
	previousSeparator := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			previousSeparator = false
			continue
		}
		if !previousSeparator {
			builder.WriteByte('_')
			previousSeparator = true
		}
	}
	return strings.Trim(builder.String(), "_")
}
