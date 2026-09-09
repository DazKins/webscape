package component

func (c *CLocomotion) Save() (SavedComponent, error) {
	return marshalSaved(struct{}{})
}
func init() { registerComponentRestore(ComponentIdLocomotion, 1, restoreLocomotion) }
func restoreLocomotion(saved SavedComponent) (Component, error) {
	var s struct{}
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return NewCLocomotion(LocomotionPhaseIdle, 0), nil
}

func (c *CLocomotion) idleForRestore(tick uint64) Component {
	return NewCLocomotion(LocomotionPhaseIdle, tick)
}
