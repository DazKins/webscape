package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdWoodcuttingSwing = ComponentId("woodcuttingswing")

type CWoodcuttingSwing struct {
	playerEntityId model.EntityId
	targetEntityId model.EntityId
}

func NewCWoodcuttingSwing(playerEntityId model.EntityId, targetEntityId model.EntityId) *CWoodcuttingSwing {
	return &CWoodcuttingSwing{
		playerEntityId: playerEntityId,
		targetEntityId: targetEntityId,
	}
}

func (c *CWoodcuttingSwing) GetId() ComponentId {
	return ComponentIdWoodcuttingSwing
}

func (c *CWoodcuttingSwing) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"playerEntityId": util.JString(c.playerEntityId.String()),
		"targetEntityId": util.JString(c.targetEntityId.String()),
	})
}

func (c *CWoodcuttingSwing) GetPlayerEntityId() model.EntityId {
	return c.playerEntityId
}

func (c *CWoodcuttingSwing) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}
