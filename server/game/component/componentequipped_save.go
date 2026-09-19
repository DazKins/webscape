package component

import (
	"fmt"
	"webscape/server/game/model"
)

type savedEquipped struct {
	Slots map[model.EquipmentSlot]*model.Item `json:"slots"`
}

func (c *CEquipped) Save() (SavedComponent, error) {
	return marshalSaved(savedEquipped{Slots: c.slots}, 2)
}
func init() {
	registerComponentRestore(ComponentIdEquipped, 1, restoreEquipped)
	registerComponentRestore(ComponentIdEquipped, 2, restoreEquipped)
}
func restoreEquipped(saved SavedComponent) (Component, error) {
	var s savedEquipped
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CEquipped{slots: s.Slots}, nil
}

func (c *CEquipped) ValidateSaved(ctx SaveContext) error {
	if c.slots == nil {
		return fmt.Errorf("missing equipment slots")
	}
	for slot, item := range c.slots {
		if err := ctx.ClaimItem(item); err != nil {
			return err
		}
		if _, ok := model.ParseEquipmentSlot(string(slot)); !ok || item.GetEquipmentSlot() == nil || *item.GetEquipmentSlot() != slot {
			return fmt.Errorf("invalid equipped item")
		}
	}
	return nil
}
