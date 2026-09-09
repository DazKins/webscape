package component

type savedRewardDrop struct {
}

func (c *CRewardDrop) Save() (SavedComponent, error) {
	return marshalSaved(savedRewardDrop{})
}
func init() { registerComponentRestore(ComponentIdRewardDrop, 1, restoreRewardDrop) }
func restoreRewardDrop(saved SavedComponent) (Component, error) {
	var s savedRewardDrop
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CRewardDrop{}, nil
}
