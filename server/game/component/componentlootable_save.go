package component

type savedLootable struct {
	Once   bool       `json:"once"`
	Looted bool       `json:"looted"`
	Items  []LootItem `json:"items"`
}

func (c *CLootable) Save() (SavedComponent, error) {
	return marshalSaved(savedLootable{Once: c.once, Looted: c.looted, Items: c.items}, 2)
}
func init() {
	registerComponentRestore(ComponentIdLootable, 1, restoreLootable)
	registerComponentRestore(ComponentIdLootable, 2, restoreLootable)
}
func restoreLootable(saved SavedComponent) (Component, error) {
	var s savedLootable
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CLootable{once: s.Once, looted: s.Looted, items: s.Items}, nil
}

func (c *CLootable) ValidateSaved(ctx SaveContext) error {
	for _, item := range c.items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}
