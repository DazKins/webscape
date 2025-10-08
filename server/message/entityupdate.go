package message

import (
	"webscape/server/game/entity"
)

type entityUpdateData struct {
	Id         string         `json:"id"`
	Components map[string]any `json:"components"`
}

type SerializeableComponent interface {
	Serialize() map[string]any
}

func NewEntityUpdateMessage(entity *entity.Entity) Message {
	serializedComponents := make(map[string]any)
	for _, component := range entity.GetComponents() {
		if serializeableComponent, ok := component.(SerializeableComponent); ok {
			name := component.GetId().String()
			serializedComponents[name] = serializeableComponent.Serialize()
		}
	}

	return newMessage(
		MessageTypeEntityUpdate,
		entityUpdateData{
			Id:         entity.GetId().String(),
			Components: serializedComponents,
		},
	)
}
