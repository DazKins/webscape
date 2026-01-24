package component

import "webscape/server/util"

const ComponentIdPlayer = ComponentId("player")

type CPlayer struct {
	name string
}

func NewCPlayer(name string) *CPlayer {
	return &CPlayer{
		name: name,
	}
}

func (c *CPlayer) GetId() ComponentId {
	return ComponentIdPlayer
}

func (c *CPlayer) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"name": util.JString(c.name),
	})
}

func (c *CPlayer) GetName() string {
	return c.name
}
