package component

type InteractionOption string

const (
	InteractionOptionTalk   = "talk"
	InteractionOptionTrade  = "trade"
	InteractionOptionAttack = "attack"
)

const ComponentIdInteractable ComponentId = "interactable"

type CInteractable struct {
	interactionOptions []InteractionOption
}

func NewCInteractable(interactionOptions []InteractionOption) *CInteractable {
	return &CInteractable{interactionOptions: interactionOptions}
}

func (c *CInteractable) GetId() ComponentId {
	return ComponentIdInteractable
}

func (c *CInteractable) Update() bool {
	return false
}

func (c *CInteractable) Serialize() map[string]any {
	return map[string]any{
		"interactionOptions": c.interactionOptions,
	}
}
