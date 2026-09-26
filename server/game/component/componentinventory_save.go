package component

import (
	"fmt"
	"webscape/server/game/model"
)

type savedInventory struct {
	Items []*model.Item `json:"items"`
}

func (c *CInventory) Save() (SavedComponent, error) {
	if err := c.validateSlots(); err != nil {
		return SavedComponent{}, err
	}
	slots := make([]int, len(c.items))
	for i, item := range c.items {
		slots[i] = c.slots[item.Id]
	}
	return marshalSaved(savedInventoryGrid{Items: c.items, Slots: slots}, 3)
}
func init() {
	registerComponentRestore(ComponentIdInventory, 3, restoreInventory)
	registerComponentRestore(ComponentIdInventory, 1, restoreInventory)
	registerComponentRestore(ComponentIdInventory, 2, restoreInventory)
}
func restoreInventory(saved SavedComponent) (Component, error) {
	if saved.Version == 3 {
		var s savedInventoryGrid
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		if len(s.Items) != len(s.Slots) {
			return nil, fmt.Errorf("inventory slot count mismatch")
		}
		c := NewCInventory()
		c.items = s.Items
		for i, item := range s.Items {
			if item != nil {
				c.slots[item.Id] = s.Slots[i]
			}
		}
		if err := c.validateSlots(); err != nil {
			return nil, err
		}
		return c, nil
	}
	var s savedInventory
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	c := NewCInventory()
	c.items = s.Items
	for i, item := range s.Items {
		if item != nil {
			c.slots[item.Id] = i
		}
	}
	return c, nil
}

func (c *CInventory) ValidateSaved(ctx SaveContext) error {
	if err := c.validateSlots(); err != nil {
		return err
	}
	if len(c.items) > InventoryCapacity {
		return fmt.Errorf("inventory exceeds capacity")
	}
	types := map[string]bool{}
	for _, item := range c.items {
		if err := ctx.ClaimItem(item); err != nil {
			return err
		}
		if item.IsStackable() && types[item.DefinitionID] {
			return fmt.Errorf("duplicate inventory stack")
		}
		if item.IsStackable() {
			types[item.DefinitionID] = true
		}
	}
	return nil
}

type savedInventoryGrid struct {
	Items []*model.Item `json:"items"`
	Slots []int         `json:"slots"`
}

func (c *CInventory) validateSlots() error {
	if len(c.slots) != len(c.items) {
		return fmt.Errorf("inventory slot count mismatch")
	}
	seen := map[int]bool{}
	for _, item := range c.items {
		if item == nil {
			return fmt.Errorf("nil inventory item")
		}
		slot, ok := c.slots[item.Id]
		if !ok || slot < 0 || slot >= InventoryCapacity || seen[slot] {
			return fmt.Errorf("invalid inventory slot")
		}
		seen[slot] = true
	}
	return nil
}
