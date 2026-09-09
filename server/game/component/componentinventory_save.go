package component

import (
	"fmt"
	"webscape/server/game/model"
)

type savedInventory struct {
	Items []*model.Item `json:"items"`
}

func (c *CInventory) Save() (SavedComponent, error) {
	return marshalSaved(savedInventory{Items: c.items})
}
func init() { registerComponentRestore(ComponentIdInventory, 1, restoreInventory) }
func restoreInventory(saved SavedComponent) (Component, error) {
	var s savedInventory
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CInventory{items: s.Items}, nil
}

func (c *CInventory) ValidateSaved(ctx SaveContext) error {
	if len(c.items) > InventoryCapacity {
		return fmt.Errorf("inventory exceeds capacity")
	}
	types := map[string]bool{}
	for _, item := range c.items {
		if err := ctx.ClaimItem(item); err != nil {
			return err
		}
		if item.IsStackable() && types[item.Type] {
			return fmt.Errorf("duplicate inventory stack")
		}
		types[item.Type] = true
	}
	return nil
}
