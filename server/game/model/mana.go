package model

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ManaSettings are authored in game.json. Regeneration follows server ticks,
// including during combat; offline players do not participate in simulation.
type ManaSettings struct {
	MaxMana            int `json:"maxMana"`
	StaffCastCost      int `json:"staffCastCost"`
	RegenAmount        int `json:"regenAmount"`
	RegenIntervalTicks int `json:"regenIntervalTicks"`
}

func DefaultManaSettings() ManaSettings {
	return ManaSettings{MaxMana: 100, StaffCastCost: 10, RegenAmount: 1, RegenIntervalTicks: 2}
}

func (s *ManaSettings) UnmarshalJSON(data []byte) error {
	type settings ManaSettings
	var value settings
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	for _, n := range []int{value.MaxMana, value.StaffCastCost, value.RegenAmount, value.RegenIntervalTicks} {
		if n < 1 || n > 2147483647 {
			return fmt.Errorf("mana settings must contain positive integers up to 2147483647")
		}
	}
	if value.StaffCastCost > value.MaxMana || value.RegenAmount > value.MaxMana {
		return fmt.Errorf("mana staffCastCost and regenAmount must not exceed maxMana")
	}
	*s = ManaSettings(value)
	return nil
}
