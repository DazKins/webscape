package component

import "webscape/server/util"

type ComponentId string

func (c ComponentId) String() string {
	return string(c)
}

type Component interface {
	GetId() ComponentId
}

type SerializeableComponent interface {
	Serialize() util.Json
}
