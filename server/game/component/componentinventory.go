package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdInventory = ComponentId("inventory")

type CInventory struct {
	items []*model.Item
}

func NewCInventory() *CInventory {
	return &CInventory{
		items: []*model.Item{},
	}
}

func (c *CInventory) GetId() ComponentId {
	return ComponentIdInventory
}

func (c *CInventory) Serialize() util.Json {
	itemsArray := make(util.JArray, len(c.items))
	for i, item := range c.items {
		itemObj := util.JObject(map[string]util.Json{
			"id":   util.JString(item.Id.String()),
			"name": util.JString(item.Name),
			"type": util.JString(item.Type),
		})
		if item.EquipmentSlot != nil {
			itemObj["equipmentSlot"] = util.JString(string(*item.EquipmentSlot))
		}
		itemsArray[i] = itemObj
	}
	return util.JObject(map[string]util.Json{
		"items": itemsArray,
	})
}

func (c *CInventory) AddItem(item *model.Item) {
	c.items = append(c.items, item)
}

func (c *CInventory) RemoveItem(itemId model.ItemId) bool {
	for i, item := range c.items {
		if item.Id == itemId {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return true
		}
	}
	return false
}

func (c *CInventory) GetItem(itemId model.ItemId) *model.Item {
	for _, item := range c.items {
		if item.Id == itemId {
			return item
		}
	}
	return nil
}

func (c *CInventory) HasItem(itemId model.ItemId) bool {
	return c.GetItem(itemId) != nil
}

func (c *CInventory) GetAllItems() []*model.Item {
	// Return a copy to prevent external modification
	result := make([]*model.Item, len(c.items))
	copy(result, c.items)
	return result
}

func (c *CInventory) GetItemCount() int {
	return len(c.items)
}
