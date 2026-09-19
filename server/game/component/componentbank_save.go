package component

import "fmt"

func (c *CBank) Save() (SavedComponent, error) {
	return marshalSaved(savedInventory{Items: c.storage.items})
}
func init() {
	registerComponentRestore(ComponentIdBank, 1, func(saved SavedComponent) (Component, error) {
		var s savedInventory
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CBank{storage: CInventory{items: s.Items}}, nil
	})
}
func (c *CBank) ValidateSaved(ctx SaveContext) error {
	if ctx.Component(ComponentIdPlayer) == nil {
		return fmt.Errorf("bank requires a player")
	}
	seen := map[string]bool{}
	for _, item := range c.storage.items {
		if err := ctx.ClaimItem(item); err != nil {
			return err
		}
		if item.IsStackable() {
			if seen[item.DefinitionID] {
				return fmt.Errorf("duplicate bank stack")
			}
			seen[item.DefinitionID] = true
		}
	}
	return nil
}
