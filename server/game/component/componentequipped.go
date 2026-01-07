package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdEquipped = ComponentId("equipped")

type EquipmentSlot string

const (
	SlotHead    EquipmentSlot = "head"
	SlotChest   EquipmentSlot = "chest"
	SlotLegs    EquipmentSlot = "legs"
	SlotFeet    EquipmentSlot = "feet"
	SlotWeapon  EquipmentSlot = "weapon"
	SlotOffhand EquipmentSlot = "offhand"
)

type CEquipped struct {
	slots map[EquipmentSlot]*model.Item
}

func NewCEquipped() *CEquipped {
	return &CEquipped{
		slots: make(map[EquipmentSlot]*model.Item),
	}
}

func (c *CEquipped) GetId() ComponentId {
	return ComponentIdEquipped
}

func (c *CEquipped) Serialize() util.Json {
	slotsObject := make(util.JObject)
	for slot, item := range c.slots {
		if item != nil {
			slotsObject[string(slot)] = util.JObject(map[string]util.Json{
				"id":   util.JString(item.Id.String()),
				"name": util.JString(item.Name),
				"type": util.JString(item.Type),
			})
		} else {
			slotsObject[string(slot)] = util.JNull{}
		}
	}
	return util.JObject(map[string]util.Json{
		"slots": slotsObject,
	})
}

func (c *CEquipped) EquipItem(slot EquipmentSlot, item *model.Item) *model.Item {
	// Store the previously equipped item (if any) to return it
	previouslyEquipped := c.slots[slot]
	c.slots[slot] = item
	return previouslyEquipped
}

func (c *CEquipped) UnequipItem(slot EquipmentSlot) *model.Item {
	item := c.slots[slot]
	if item != nil {
		delete(c.slots, slot)
	}
	return item
}

func (c *CEquipped) GetEquippedItem(slot EquipmentSlot) *model.Item {
	return c.slots[slot]
}

func (c *CEquipped) IsSlotEquipped(slot EquipmentSlot) bool {
	return c.slots[slot] != nil
}

func (c *CEquipped) GetAllEquippedItems() map[EquipmentSlot]*model.Item {
	// Return a copy to prevent external modification
	result := make(map[EquipmentSlot]*model.Item)
	for slot, item := range c.slots {
		result[slot] = item
	}
	return result
}
