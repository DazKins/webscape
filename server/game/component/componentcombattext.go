package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdCombatText = ComponentId("combattext")

type CCombatText struct {
	fromEntityId model.EntityId
	text         string
	kind         string
}

func NewCCombatText(fromEntityId model.EntityId, text string, kind string) *CCombatText {
	return &CCombatText{
		fromEntityId: fromEntityId,
		text:         text,
		kind:         kind,
	}
}

func (c *CCombatText) GetId() ComponentId {
	return ComponentIdCombatText
}

func (c *CCombatText) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"fromEntityId": util.JString(c.fromEntityId.String()),
		"text":         util.JString(c.text),
		"kind":         util.JString(c.kind),
	})
}

func (c *CCombatText) GetFromEntityId() model.EntityId {
	return c.fromEntityId
}

func (c *CCombatText) GetText() string {
	return c.text
}

func (c *CCombatText) GetKind() string {
	return c.kind
}
