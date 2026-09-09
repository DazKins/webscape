package component

import (
	"fmt"
	"webscape/server/util"
)

func (c *CMetadata) Save() (SavedComponent, error) {
	return marshalSaved(c.metadata)
}
func init() { registerComponentRestore(ComponentIdMetadata, 1, restoreMetadata) }
func restoreMetadata(saved SavedComponent) (Component, error) {
	var raw any
	if err := decodeSaved(saved.Data, &raw); err != nil {
		return nil, err
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("metadata must be an object")
	}
	return NewCMetadata(saveJSON(obj)), nil
}

// saveJSON converts decoded metadata back to the game's JSON value types.
func saveJSON(v any) util.Json {
	switch v := v.(type) {
	case nil:
		return util.JNull{}
	case bool:
		return util.JBool(v)
	case float64:
		return util.JNumber(v)
	case string:
		return util.JString(v)
	case []any:
		result := make(util.JArray, len(v))
		for i, item := range v {
			result[i] = saveJSON(item)
		}
		return result
	case map[string]any:
		result := util.JObject{}
		for key, item := range v {
			result[key] = saveJSON(item)
		}
		return result
	default:
		panic("unsupported decoded JSON type")
	}
}

func (c *CMetadata) ValidateSaved(ctx SaveContext) error {
	metadata, ok := c.metadata.(util.JObject)
	if !ok {
		return fmt.Errorf("metadata must be an object")
	}
	size := ctx.ChunkSize()
	for _, dimension := range []struct {
		name  string
		limit int
	}{{"width", size.X}, {"height", size.Y}} {
		if raw, exists := metadata[dimension.name]; exists {
			n, ok := raw.(util.JNumber)
			if !ok || n < 1 || n > util.JNumber(dimension.limit) || n != util.JNumber(int(n)) {
				return fmt.Errorf("invalid saved entity footprint")
			}
		}
	}
	return nil
}
