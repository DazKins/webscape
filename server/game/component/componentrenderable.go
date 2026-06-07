package component

import "webscape/server/util"

const ComponentIdRenderable = ComponentId("renderable")

type CRenderable struct {
	renderType  string
	orientation string
}

func NewCRenderable(renderType string, orientation ...string) *CRenderable {
	renderOrientation := ""
	if len(orientation) > 0 {
		renderOrientation = orientation[0]
	}
	return &CRenderable{
		renderType:  renderType,
		orientation: renderOrientation,
	}
}

func (c *CRenderable) GetId() ComponentId {
	return ComponentIdRenderable
}

func (c *CRenderable) Serialize() util.Json {
	result := util.JObject(map[string]util.Json{
		"type": util.JString(c.renderType),
	})
	if c.orientation != "" {
		result["orientation"] = util.JString(c.orientation)
	}
	return result
}

func (c *CRenderable) GetType() string {
	return c.renderType
}

func (c *CRenderable) SetType(renderType string) {
	c.renderType = renderType
}

func (c *CRenderable) GetOrientation() string {
	return c.orientation
}

func (c *CRenderable) SetOrientation(orientation string) {
	c.orientation = orientation
}
