package component

import (
	"fmt"
)

type savedLogEntry struct {
	Text string `json:"text"`
	Kind string `json:"kind"`
}

type savedCombatLog struct {
	Entries    []savedLogEntry `json:"entries"`
	MaxEntries int             `json:"maxEntries"`
}

func (c *CCombatLog) Save() (SavedComponent, error) {
	s := savedCombatLog{MaxEntries: c.maxEntries, Entries: []savedLogEntry{}}
	for _, e := range c.entries {
		s.Entries = append(s.Entries, savedLogEntry{e.text, e.kind})
	}
	return marshalSaved(s)
}
func init() { registerComponentRestore(ComponentIdCombatLog, 1, restoreCombatLog) }
func restoreCombatLog(saved SavedComponent) (Component, error) {
	var s savedCombatLog
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	if s.MaxEntries < 0 || len(s.Entries) > s.MaxEntries {
		return nil, fmt.Errorf("invalid combat log capacity")
	}
	c := NewCCombatLog(s.MaxEntries)
	for _, e := range s.Entries {
		c.AddEntry(NewCombatLogEntry(e.Text, e.Kind))
	}
	return c, nil
}
