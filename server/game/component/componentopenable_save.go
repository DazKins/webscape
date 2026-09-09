package component

type savedOpenable struct {
	IsOpen bool `json:"isOpen"`
}

func (c *COpenable) Save() (SavedComponent, error) {
	return marshalSaved(savedOpenable{IsOpen: c.isOpen})
}
func init() { registerComponentRestore(ComponentIdOpenable, 1, restoreOpenable) }
func restoreOpenable(saved SavedComponent) (Component, error) {
	var s savedOpenable
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &COpenable{isOpen: s.IsOpen}, nil
}
