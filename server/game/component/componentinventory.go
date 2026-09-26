package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdInventory = ComponentId("inventory")
const InventoryWidth = 4
const InventoryHeight = 5
const InventoryCapacity = InventoryWidth * InventoryHeight

type CInventory struct {
	items []*model.Item
	slots map[model.ItemId]int
}

func NewCInventory() *CInventory {
	return &CInventory{
		items: []*model.Item{},
		slots: make(map[model.ItemId]int),
	}
}

func (c *CInventory) GetId() ComponentId {
	return ComponentIdInventory
}

func (c *CInventory) Serialize() util.Json {
	itemsArray := make(util.JArray, len(c.items))
	for i, item := range c.items {
		value := SerializeItem(item).(util.JObject)
		value["slot"] = util.JNumber(c.slots[item.Id])
		itemsArray[i] = value
	}
	return util.JObject(map[string]util.Json{
		"items":  itemsArray,
		"width":  util.JNumber(InventoryWidth),
		"height": util.JNumber(InventoryHeight),
	})
}

func (c *CInventory) AddItem(item *model.Item) bool {
	return c.addItem(item, InventoryCapacity)
}

func (c *CInventory) addItem(item *model.Item, capacity int) bool {
	if item == nil || item.ValidateSaved() != nil || c.HasItem(item.Id) {
		return false
	}
	if item.IsStackable() {
		for _, existing := range c.items {
			if existing.DefinitionID == item.DefinitionID && existing.IsStackable() {
				if item.Quantity > model.MaxStackQuantity-existing.Quantity {
					return false
				}
				existing.Quantity += item.Quantity
				return true
			}
		}
	}
	if capacity > 0 && len(c.items) >= capacity {
		return false
	}
	occupied := make(map[int]bool, len(c.slots))
	for _, slot := range c.slots {
		occupied[slot] = true
	}
	slot := 0
	for occupied[slot] {
		slot++
	}
	if c.slots == nil {
		c.slots = make(map[model.ItemId]int)
	}
	c.slots[item.Id] = slot
	c.items = append(c.items, item)
	return true
}

func (c *CInventory) RemoveItem(itemId model.ItemId) bool {
	for i, item := range c.items {
		if item.Id == itemId {
			delete(c.slots, item.Id)
			c.items = append(c.items[:i], c.items[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveFirstItemByDefinition removes one unit, preserving any remaining stack.
func (c *CInventory) RemoveFirstItemByDefinition(definitionID string) *model.Item {
	for i, item := range c.items {
		if item.DefinitionID == definitionID {
			if item.Quantity > 1 {
				item.Quantity--
				removed := item.Clone()
				removed.Id = model.NewItemId()
				removed.Quantity = 1
				return removed
			}
			delete(c.slots, item.Id)
			c.items = append(c.items[:i], c.items[i+1:]...)
			return item
		}
	}
	return nil
}

func (c *CInventory) HasItemDefinition(definitionID string) bool {
	for _, item := range c.items {
		if item.DefinitionID == definitionID {
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
		result.items = append(result.items, item.Clone())
		result.slots[item.Id] = c.slots[item.Id]
	}
	return result
}

func (c *CInventory) FindByDefinition(definitionID string) *model.Item {
	for _, item := range c.items {
		if item.DefinitionID == definitionID {
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
	c.items, c.slots = trial.items, trial.slots
	return true
}

func (c *CInventory) itemAtSlot(slot int) *model.Item {
	for _, item := range c.items {
		if c.slots[item.Id] == slot {
			return item
		}
	}
	return nil
}

// MoveItem preserves other positions, swapping occupied slots or merging stacks.
func (c *CInventory) MoveItem(id model.ItemId, slot int) bool {
	item := c.GetItem(id)
	if item == nil || slot < 0 || slot >= InventoryCapacity {
		return false
	}
	source := c.slots[id]
	if source == slot {
		return true
	}
	if target := c.itemAtSlot(slot); target != nil {
		if item.IsStackable() && target.IsStackable() && item.DefinitionID == target.DefinitionID {
			if item.Quantity > model.MaxStackQuantity-target.Quantity {
				return false
			}
			target.Quantity += item.Quantity
			c.RemoveItem(id)
			return true
		}
		c.slots[target.Id] = source
	}
	c.slots[id] = slot
	return true
}
