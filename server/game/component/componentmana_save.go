package component

import "fmt"

type savedMana struct {
	MaxMana     int `json:"maxMana"`
	CurrentMana int `json:"currentMana"`
}

func (c *CMana) Save() (SavedComponent, error) {
	return marshalSaved(savedMana{MaxMana: c.maxMana, CurrentMana: c.currentMana})
}

func init() { registerComponentRestore(ComponentIdMana, 1, restoreMana) }

func restoreMana(saved SavedComponent) (Component, error) {
	var s savedMana
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	c := &CMana{maxMana: s.MaxMana, currentMana: s.CurrentMana}
	if err := c.ValidateSaved(nil); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *CMana) ValidateSaved(_ SaveContext) error {
	if c.maxMana <= 0 || c.currentMana < 0 || c.currentMana > c.maxMana {
		return fmt.Errorf("invalid mana")
	}
	return nil
}
