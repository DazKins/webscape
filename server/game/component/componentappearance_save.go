package component

import ()

type savedAppearance struct {
	Appearance Appearance `json:"appearance"`
}

func (c *CAppearance) Save() (SavedComponent, error) {
	return marshalSaved(savedAppearance{Appearance: c.appearance})
}
func init() { registerComponentRestore(ComponentIdAppearance, 1, restoreAppearance) }
func restoreAppearance(saved SavedComponent) (Component, error) {
	var s savedAppearance
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CAppearance{appearance: s.Appearance}, nil
}

func (c *CAppearance) ValidateSaved(ctx SaveContext) error {
	if err := ValidateAppearance(c.appearance); err != nil {
		return err
	}
	return nil
}
