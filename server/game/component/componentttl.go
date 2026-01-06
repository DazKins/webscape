package component

import (
	"webscape/server/util"
)

const ComponentIdTtl = ComponentId("ttl")

type CTtl struct {
	Remaining uint64
}

func (c *CTtl) GetId() ComponentId {
	return ComponentIdTtl
}

func (c *CTtl) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"remaining": util.JNumber(c.Remaining),
	})
}

