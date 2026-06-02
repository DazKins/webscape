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
}

func New(id string, actorEntityId model.EntityId) Event {
	return Event{
		Id:            id,
		ActorEntityId: actorEntityId,
		Count:         1,
		Metadata:      map[string]string{},
	}
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
