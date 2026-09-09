package component

type savedRenderable struct {
	RenderType  string `json:"renderType"`
	Orientation string `json:"orientation"`
}

func (c *CRenderable) Save() (SavedComponent, error) {
	return marshalSaved(savedRenderable{RenderType: c.renderType, Orientation: c.orientation})
}
func init() { registerComponentRestore(ComponentIdRenderable, 1, restoreRenderable) }
func restoreRenderable(saved SavedComponent) (Component, error) {
	var s savedRenderable
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CRenderable{renderType: s.RenderType, orientation: s.Orientation}, nil
}
