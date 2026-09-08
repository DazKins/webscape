package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdInventory = ComponentId("inventory")
const InventoryCapacity = 20

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
		itemsArray[i] = SerializeItem(item)
	}
	return util.JObject(map[string]util.Json{
		"items": itemsArray,
	})
}

func (c *CInventory) AddItem(item *model.Item) bool {
	if item == nil || !item.ValidQuantity() || c.HasItem(item.Id) {
		return false
	}
	if item.IsStackable() {
		for _, existing := range c.items {
			if existing.Type == item.Type {
				if item.Quantity > model.MaxStackQuantity-existing.Quantity {
					return false
				}
				existing.Quantity += item.Quantity
				return true
			}
		}
	}
	if c.IsFull() {
		return false
	}
	c.items = append(c.items, item)
	return true
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

func (c *CInventory) RemoveFirstItemByType(itemType string) *model.Item {
	for i, item := range c.items {
		if item.Type == itemType {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return item
		}
	}
	return nil
}

func (c *CInventory) HasItemType(itemType string) bool {
	for _, item := range c.items {
		if item.Type == itemType {
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

func (c *CInventory) IsFull() bool {
	return len(c.items) >= InventoryCapacity
}

func (c *CInventory) AvailableSlots() int {
	available := InventoryCapacity - len(c.items)
	if available < 0 {
		return 0
	}
	return available
}

// Clone isolates quantity mutations while preparing atomic inventory operations.
func (c *CInventory) Clone() *CInventory {
	result := NewCInventory()
	for _, item := range c.items {
		copied := *item
		result.items = append(result.items, &copied)
	}
	return result
}

func (c *CInventory) FindByType(itemType string) *model.Item {
	for _, item := range c.items {
		if item.Type == itemType {
			return item
		}
	}
	return nil
}

// Exchange commits only if payment and the resulting slot/stack layout both fit.
func (c *CInventory) Exchange(paymentID model.ItemId, quantity int, received *model.Item) bool {
	trial := c.Clone()
	payment := trial.GetItem(paymentID)
	if payment == nil || quantity <= 0 || payment.Quantity < quantity {
		return false
	}
	if payment.Quantity == quantity {
		trial.RemoveItem(paymentID)
	} else {
		payment.Quantity -= quantity
	}
	if !trial.AddItem(received) {
		return false
	}
	c.items = trial.items
	return true
}
