package component

import (
	"fmt"
	"webscape/server/game/model"
)

// ParseAuthoredEquipment resolves catalog ids into fresh, authoritative items.
// An empty object remains valid for older authored entities.
func ParseAuthoredEquipment(raw any) (*CEquipped, error) {
	value, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("equipped must be an object")
	}
	equipped := NewCEquipped()
	for key := range value {
		if key != "slots" {
			return nil, fmt.Errorf("unknown equipped field %q", key)
		}
	}
	rawSlots, exists := value["slots"]
	if !exists {
		return equipped, nil
	}
	slots, ok := rawSlots.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("equipped.slots must be an object")
	}
	for name, rawID := range slots {
		slot, valid := model.ParseEquipmentSlot(name)
		id, isString := rawID.(string)
		item := model.CreateShopItem(id)
		if !valid || !isString || item == nil || item.GetEquipmentSlot() == nil || *item.GetEquipmentSlot() != slot {
			return nil, fmt.Errorf("equipped slot %q must reference a compatible equipment catalog id", name)
		}
		equipped.EquipItem(slot, item)
	}
	return equipped, nil
}
