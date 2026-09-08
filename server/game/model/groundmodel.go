package model

import "strings"

// GroundRenderModel also covers authored loot, which is defined by name/type.
func (item *Item) GroundRenderModel() string {
	if item.RenderModel != "" {
		return item.RenderModel
	}
	switch strings.ToLower(item.Name) {
	case "health potion":
		return "healthPotion"
	case "bread":
		return "bread"
	case "apple":
		return "apple"
	case "iron ore":
		return "ironOre"
	case "wood":
		return "wood"
	case "logs":
		return "logs"
	case "stone":
		return "stone"
	case "mysterious key":
		return "mysteriousKey"
	case "ancient scroll":
		return "ancientScroll"
	}
	switch item.Type {
	case ItemTypeGold:
		return "gold"
	case ItemTypeArrow:
		return "arrow"
	case "fish":
		return "fish"
	}
	return "unknownItem"
}
