package component

import (
	"webscape/server/util"
)

const ComponentIdTtl = ComponentId("ttl")

type CTtl struct {
	remaining uint64
}

func NewCTtl(remaining uint64) *CTtl {
	return &CTtl{
		remaining: remaining,
	}
}

func (c *CTtl) GetId() ComponentId {
	return ComponentIdTtl
}

func (c *CTtl) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"remaining": util.JNumber(c.remaining),
	})
}

func (c *CTtl) GetRemaining() uint64 {
	return c.remaining
}

func (c *CTtl) SetRemaining(remaining uint64) {
	c.remaining = remaining
}

