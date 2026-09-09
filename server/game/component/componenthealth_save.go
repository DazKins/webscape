package component

import (
	"fmt"
)

type savedHealth struct {
	MaxHealth     int `json:"maxHealth"`
	CurrentHealth int `json:"currentHealth"`
}

func (c *CHealth) Save() (SavedComponent, error) {
	return marshalSaved(savedHealth{MaxHealth: c.maxHealth, CurrentHealth: c.currentHealth})
}
func init() { registerComponentRestore(ComponentIdHealth, 1, restoreHealth) }
func restoreHealth(saved SavedComponent) (Component, error) {
	var s savedHealth
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CHealth{maxHealth: s.MaxHealth, currentHealth: s.CurrentHealth}, nil
}

func (c *CHealth) ValidateSaved(ctx SaveContext) error {
	if c.maxHealth <= 0 || c.currentHealth > c.maxHealth {
		return fmt.Errorf("invalid health")
	}
	return nil
}
