package component

import (
	"fmt"
	"webscape/server/game/model"
)

type savedWoodcuttable struct {
	MaxDurability         int            `json:"maxDurability"`
	CurrentDurability     int            `json:"currentDurability"`
	RespawnTicks          int            `json:"respawnTicks"`
	Yield                 LootItem       `json:"yield"`
	Depleted              bool           `json:"depleted"`
	RemainingRespawnTicks int            `json:"remainingRespawnTicks"`
	LastFellerEntityId    model.EntityId `json:"lastFellerEntityId"`
}

func (c *CWoodcuttable) Save() (SavedComponent, error) {
	return marshalSaved(savedWoodcuttable{MaxDurability: c.maxDurability, CurrentDurability: c.currentDurability, RespawnTicks: c.respawnTicks, Yield: c.yield, Depleted: c.depleted, RemainingRespawnTicks: c.remainingRespawnTicks, LastFellerEntityId: c.lastFellerEntityId})
}
func init() { registerComponentRestore(ComponentIdWoodcuttable, 1, restoreWoodcuttable) }
func restoreWoodcuttable(saved SavedComponent) (Component, error) {
	var s savedWoodcuttable
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CWoodcuttable{maxDurability: s.MaxDurability, currentDurability: s.CurrentDurability, respawnTicks: s.RespawnTicks, yield: s.Yield, depleted: s.Depleted, remainingRespawnTicks: s.RemainingRespawnTicks, lastFellerEntityId: s.LastFellerEntityId}, nil
}

func (c *CWoodcuttable) ValidateSaved(ctx SaveContext) error {
	if c.maxDurability < 1 || c.currentDurability < 0 || c.currentDurability > c.maxDurability || c.remainingRespawnTicks < 0 {
		return fmt.Errorf("invalid resource state")
	}
	return nil
}
