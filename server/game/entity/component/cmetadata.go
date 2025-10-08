package component

const ComponentIdMetadata ComponentId = "metadata"

type CMetadata struct {
	metadata map[string]any
}

func NewCMetadata(metadata map[string]any) *CMetadata {
	return &CMetadata{metadata: metadata}
}

func (c *CMetadata) Update() bool {
	return false
}

func (c *CMetadata) GetId() ComponentId {
	return ComponentIdMetadata
}

func (c *CMetadata) Serialize() map[string]any {
	return c.metadata
}
