package component

import (
	"webscape/server/util"
)

const ComponentIdMetadata = ComponentId("metadata")

type CMetadata struct {
	metadata util.Json
}

func NewCMetadata(metadata util.Json) *CMetadata {
	return &CMetadata{
		metadata: metadata,
	}
}

func (c *CMetadata) GetId() ComponentId {
	return ComponentIdMetadata
}

func (c *CMetadata) Serialize() util.Json {
	return c.metadata
}

func (c *CMetadata) GetMetadata() util.Json {
	return c.metadata
}

func (c *CMetadata) SetMetadata(metadata util.Json) {
	c.metadata = metadata
}
