package component

type savedBaseStats struct {
	Strength  int `json:"strength"`
	Dexterity int `json:"dexterity"`
	Vitality  int `json:"vitality"`
}

func (c *CBaseStats) Save() (SavedComponent, error) {
	return marshalSaved(savedBaseStats{Strength: c.strength, Dexterity: c.dexterity, Vitality: c.vitality})
}
func init() { registerComponentRestore(ComponentIdBaseStats, 1, restoreBaseStats) }
func restoreBaseStats(saved SavedComponent) (Component, error) {
	var s savedBaseStats
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CBaseStats{strength: s.Strength, dexterity: s.Dexterity, vitality: s.Vitality}, nil
}
