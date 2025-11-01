package component

type ComponentId string

func (c ComponentId) String() string {
	return string(c)
}

type Component interface {
	GetId() ComponentId
	ShouldSend() bool
	MarkSent()
}

type SerializeableComponent interface {
	Serialize() map[string]any
}
