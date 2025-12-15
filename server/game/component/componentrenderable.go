package component

import "webscape/server/util"

const ComponentIdRenderable = ComponentId("renderable")

type CRenderable struct {
	Type string
}

func (c *CRenderable) GetId() ComponentId {
	return ComponentIdRenderable
}

func (c *CRenderable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"type": util.JString(c.Type),
	})
}
