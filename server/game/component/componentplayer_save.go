package component

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type savedPlayer struct {
	Name string `json:"name"`
}

func (c *CPlayer) Save() (SavedComponent, error) {
	return marshalSaved(savedPlayer{Name: c.name})
}
func init() { registerComponentRestore(ComponentIdPlayer, 1, restorePlayer) }
func restorePlayer(saved SavedComponent) (Component, error) {
	var s savedPlayer
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CPlayer{name: s.Name}, nil
}

func (c *CPlayer) ValidateSaved(ctx SaveContext) error {
	if strings.TrimSpace(c.name) == "" || utf8.RuneCountInString(c.name) > 24 {
		return fmt.Errorf("invalid player name")
	}
	for _, id := range []ComponentId{ComponentIdMetadata, ComponentIdRenderable, ComponentIdAppearance, ComponentIdHealth, ComponentIdInventory, ComponentIdEquipped, ComponentIdBaseStats, ComponentIdCombatStats, ComponentIdCombatLog, ComponentIdQuestLog, ComponentIdLocomotion} {
		if ctx.Component(id) == nil {
			return fmt.Errorf("player missing component %s", id)
		}
	}

	return nil
}
