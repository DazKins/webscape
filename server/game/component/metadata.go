package component

import "webscape/server/game/component/componentheader"

const ComponentIdMetadata = ComponentId("metadata")

type CMetadata struct {
	componentheader.ComponentHeader
	metadata map[string]any
}

func (c *CMetadata) GetId() ComponentId {
	return ComponentIdMetadata
}

func (c *CMetadata) Serialize() map[string]any {
	return c.metadata
}

func (c *CMetadata) GetMetadata() map[string]any {
	return c.metadata
}

func (c *CMetadata) SetMetadata(metadata map[string]any) {
	c.metadata = metadata
	c.Update()
}
