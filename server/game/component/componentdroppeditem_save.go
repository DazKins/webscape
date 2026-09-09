package component

import (
	"webscape/server/game/model"
)

type savedDroppedItem struct {
	Item *model.Item `json:"item"`
}

func (c *CDroppedItem) Save() (SavedComponent, error) {
	return marshalSaved(savedDroppedItem{Item: c.Item})
}
func init() { registerComponentRestore(ComponentIdDroppedItem, 1, restoreDroppedItem) }
func restoreDroppedItem(saved SavedComponent) (Component, error) {
	var s savedDroppedItem
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CDroppedItem{Item: s.Item}, nil
}

func (c *CDroppedItem) ValidateSaved(ctx SaveContext) error {
	if err := ctx.ClaimItem(c.Item); err != nil {
		return err
	}
	return nil
}
