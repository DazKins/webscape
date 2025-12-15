package system

import (
	"webscape/server/game/component"
)

type System interface {
	Update()
}

type SystemBase struct {
	ComponentManager *component.ComponentManager
}
