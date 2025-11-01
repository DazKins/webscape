package component

import (
	"webscape/server/util"
)

const ComponentIdMetadata = ComponentId("metadata")

type CMetadata struct {
	Metadata util.Json
}

func (c *CMetadata) GetId() ComponentId {
	return ComponentIdMetadata
}

func (c *CMetadata) Serialize() util.Json {
	return c.Metadata
}
