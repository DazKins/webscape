package component

import "webscape/server/util"

const ComponentIdRenderable = ComponentId("renderable")

type CRenderable struct {
	renderType string
}

func NewCRenderable(renderType string) *CRenderable {
	return &CRenderable{
		renderType: renderType,
	}
}

func (c *CRenderable) GetId() ComponentId {
	return ComponentIdRenderable
}

func (c *CRenderable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"type": util.JString(c.renderType),
	})
}

func (c *CRenderable) GetType() string {
	return c.renderType
}

func (c *CRenderable) SetType(renderType string) {
	c.renderType = renderType
}
