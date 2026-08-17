package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdFacing = ComponentId("facing")

type CFacing struct {
	targetEntityId model.EntityId
}

func NewCFacing(targetEntityId model.EntityId) *CFacing {
	return &CFacing{targetEntityId: targetEntityId}
}

func (c *CFacing) GetId() ComponentId {
	return ComponentIdFacing
}

func (c *CFacing) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"targetEntityId": util.JString(c.targetEntityId.String()),
	})
}

func (c *CFacing) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}
