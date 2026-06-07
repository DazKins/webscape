package component

import "webscape/server/util"

const ComponentIdOpenable = ComponentId("openable")

type COpenable struct {
	isOpen bool
}

func NewCOpenable(isOpen bool) *COpenable {
	return &COpenable{
		isOpen: isOpen,
	}
}

func (c *COpenable) GetId() ComponentId {
	return ComponentIdOpenable
}

func (c *COpenable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"isOpen": util.JBool(c.isOpen),
	})
}

func (c *COpenable) IsOpen() bool {
	return c.isOpen
}

func (c *COpenable) SetOpen(isOpen bool) {
	c.isOpen = isOpen
}
