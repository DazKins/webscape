package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func CreateWorldObjectEntity(object world.WorldObject) []component.Component {
	width := object.Width
	if width == 0 {
		width = 1
	}
	height := object.Height
	if height == 0 {
		height = 1
	}

	metadata := util.JObject(map[string]util.Json{
		"name":           util.JString(object.Id),
		"type":           util.JString(object.Type),
		"width":          util.JNumber(width),
		"height":         util.JNumber(height),
		"blocksMovement": util.JBool(object.BlocksMovement),
	})
	if object.State != nil {
		metadata["state"] = jsonToUtil(object.State)
	}

	components := []component.Component{
		component.NewCPosition(math.Vec2{X: object.X, Y: object.Y}),
		component.NewCMetadata(metadata),
		component.NewCRenderable(object.Type),
	}

	if len(object.Interactable) > 0 {
		options := make([]component.InteractionOption, len(object.Interactable))
		for i, option := range object.Interactable {
			options[i] = component.InteractionOption(option)
		}
		components = append(components, component.NewCInteractable(options))
	}

	return components
}

func jsonToUtil(value any) util.Json {
	switch value := value.(type) {
	case nil:
		return util.JNull{}
	case bool:
		return util.JBool(value)
	case float64:
		return util.JNumber(value)
	case string:
		return util.JString(value)
	case []any:
		result := make(util.JArray, len(value))
		for i, item := range value {
			result[i] = jsonToUtil(item)
		}
		return result
	case map[string]any:
		result := make(util.JObject)
		for key, item := range value {
			result[key] = jsonToUtil(item)
		}
		return result
	default:
		return util.JString("")
	}
}
