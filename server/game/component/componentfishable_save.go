package component

import (
	"fmt"
)

type savedFishable struct {
	CatchChancePercent int      `json:"catchChancePercent"`
	Yield              LootItem `json:"yield"`
}

func (c *CFishable) Save() (SavedComponent, error) {
	return marshalSaved(savedFishable{CatchChancePercent: c.catchChancePercent, Yield: c.yield})
}
func init() { registerComponentRestore(ComponentIdFishable, 1, restoreFishable) }
func restoreFishable(saved SavedComponent) (Component, error) {
	var s savedFishable
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CFishable{catchChancePercent: s.CatchChancePercent, yield: s.Yield}, nil
}

func (c *CFishable) ValidateSaved(ctx SaveContext) error {
	if c.catchChancePercent < 1 || c.catchChancePercent > 100 {
		return fmt.Errorf("invalid fishing chance")
	}
	return nil
}
