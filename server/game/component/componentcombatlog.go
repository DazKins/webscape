package component

import (
	"fmt"
	"webscape/server/util"
)

const ComponentIdCombatLog = ComponentId("combatlog")

type CombatLogEntry struct {
	id   string
	text string
	kind string
}

func NewCombatLogEntry(id string, text string, kind string) CombatLogEntry {
	return CombatLogEntry{
		id:   id,
		text: text,
		kind: kind,
	}
}

func (e CombatLogEntry) GetId() string {
	return e.id
}

func (e CombatLogEntry) GetText() string {
	return e.text
}

func (e CombatLogEntry) GetKind() string {
	return e.kind
}

type CCombatLog struct {
	entries    []CombatLogEntry
	maxEntries int
}

func NewCCombatLog(maxEntries int) *CCombatLog {
	return &CCombatLog{
		entries:    []CombatLogEntry{},
		maxEntries: maxEntries,
	}
}

func (c *CCombatLog) GetId() ComponentId {
	return ComponentIdCombatLog
}

func (c *CCombatLog) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"entries": util.JArrayFrom(c.entries, func(entry CombatLogEntry) util.Json {
			return util.JObject(map[string]util.Json{
				"id":   util.JString(entry.id),
				"text": util.JString(entry.text),
				"kind": util.JString(entry.kind),
			})
		}),
	})
}

func (c *CCombatLog) AddEntry(entry CombatLogEntry) {
	c.entries = append(c.entries, entry)
	if len(c.entries) > c.maxEntries {
		c.entries = c.entries[len(c.entries)-c.maxEntries:]
	}
}

func (c *CCombatLog) AddText(text string, kind string, entryIndex int) {
	c.AddEntry(NewCombatLogEntry(fmt.Sprintf("evt-%d", entryIndex), text, kind))
}

func (c *CCombatLog) GetEntries() []CombatLogEntry {
	result := make([]CombatLogEntry, len(c.entries))
	copy(result, c.entries)
	return result
}

func (c *CCombatLog) Clear() {
	c.entries = []CombatLogEntry{}
}
