package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// ItemDefinition is shared catalogue data. Instances refer to its stable ID.
// Lookup returns a copy so callers cannot mutate the catalogue.
type ItemDefinition struct {
	ID            string           `json:"-"`
	Name          string           `json:"name"`
	Type          string           `json:"type"`
	RenderModel   string           `json:"renderModel"`
	EquipmentSlot EquipmentSlot    `json:"equipmentSlot,omitempty"`
	CombatStats   *ItemCombatStats `json:"combatStats,omitempty"`
	Stackable     bool             `json:"stackable"`
}

func validateItemDefinitions(definitions map[string]ItemDefinition) map[string]ItemDefinition {
	for id, definition := range definitions {
		if id == "" || definition.Name == "" || definition.Type == "" || definition.RenderModel == "" {
			panic("invalid item definition: " + id)
		}
		if definition.EquipmentSlot != "" {
			if _, ok := ParseEquipmentSlot(string(definition.EquipmentSlot)); !ok || definition.Stackable {
				panic("invalid equipment definition: " + id)
			}
		}
		if definition.CombatStats != nil {
			if err := definition.CombatStats.Validate(); err != nil {
				panic(err)
			}
		}
		definition.ID = id
		definitions[id] = definition
	}
	return definitions
}

func GetItemDefinition(id string) (ItemDefinition, bool) {
	definition, ok := itemDefinitions[id]
	definition.CombatStats = cloneCombatStats(definition.CombatStats)
	return definition, ok
}

func ItemDefinitionIDs() []string {
	ids := make([]string, 0, len(itemDefinitions))
	for id := range itemDefinitions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ItemReference is authored content, not an inventory instance.
type ItemReference struct {
	DefinitionID string `json:"definitionId"`
	Count        int    `json:"count"`
}

func (r ItemReference) Validate() error {
	if _, ok := itemDefinitions[r.DefinitionID]; !ok {
		return fmt.Errorf("unknown item definition %q", r.DefinitionID)
	}
	if r.Count < 1 || r.Count > MaxStackQuantity {
		return fmt.Errorf("item count must be between 1 and %d", MaxStackQuantity)
	}
	return nil
}
func (r ItemReference) Name() string      { return itemDefinitions[r.DefinitionID].Name }
func (r ItemReference) Type() string      { return itemDefinitions[r.DefinitionID].Type }
func (r ItemReference) CreateItem() *Item { return NewItem(r.DefinitionID) }

// Only used when migrating old saves; gameplay never identifies items by name.
func legacyDefinitionID(name, category string) (string, error) {
	for id, definition := range itemDefinitions {
		if definition.Name == name && definition.Type == category {
			return id, nil
		}
	}
	return "", fmt.Errorf("legacy item %q (%s) has no catalogue definition", name, category)
}

// JSON map encoding sorts keys, so the fingerprint depends on definition values,
// not Go source formatting or map iteration order.
func ItemDefinitionsHash() string {
	data, err := json.Marshal(itemDefinitions)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
