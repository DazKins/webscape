package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"reflect"
)

func decodeItemJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing item data")
	}
	return nil
}

// Old snapshots stored an entire definition on each item. Resolve that identity
// once on load, preserving real stat differences as instance properties.
func (i *Item) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	if _, legacy := keys["Name"]; legacy {
		var old struct {
			Id                      ItemId
			Quantity                int
			Name, Type, RenderModel string
			EquipmentSlot           *EquipmentSlot
			CombatStats             *ItemCombatStats
		}
		if err := decodeItemJSON(data, &old); err != nil {
			return err
		}
		id, err := legacyDefinitionID(old.Name, old.Type)
		if err != nil {
			return err
		}
		definition := itemDefinitions[id]
		slot := EquipmentSlot("")
		if old.EquipmentSlot != nil {
			slot = *old.EquipmentSlot
		}
		if slot != definition.EquipmentSlot || (old.RenderModel != "" && old.RenderModel != definition.RenderModel) {
			return fmt.Errorf("legacy item %q conflicts with its definition", old.Name)
		}
		*i = Item{Id: old.Id, DefinitionID: id, Quantity: old.Quantity}
		if !reflect.DeepEqual(old.CombatStats, definition.CombatStats) {
			if old.CombatStats == nil {
				return fmt.Errorf("legacy equipment %q has no stats", old.Name)
			}
			i.Properties = &ItemProperties{CombatStats: old.CombatStats}
		}
	} else {
		type plainItem Item
		var decoded plainItem
		if err := decodeItemJSON(data, &decoded); err != nil {
			return err
		}
		*i = Item(decoded)
	}
	return i.ValidateSaved()
}

// This compatibility path only maps known legacy names; it cannot invent new
// definitions. Saving or editing the reference writes the canonical ID form.
func (r *ItemReference) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, canonical := fields["definitionId"]; canonical {
		type reference ItemReference
		var value reference
		if err := decodeItemJSON(data, &value); err != nil {
			return err
		}
		*r = ItemReference(value)
	} else {
		var old struct {
			Name  string
			Type  string
			Count int
		}
		if err := decodeItemJSON(data, &old); err != nil {
			return err
		}
		id, err := legacyDefinitionID(old.Name, old.Type)
		if err != nil {
			return err
		}
		*r = ItemReference{DefinitionID: id, Count: old.Count}
	}
	return r.Validate()
}

func (id ItemId) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }
func (id *ItemId) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		var legacy []byte
		if err := json.Unmarshal(data, &legacy); err != nil {
			return err
		}
		if len(legacy) != 16 {
			return fmt.Errorf("invalid legacy item UUID")
		}
		copy(id[:], legacy)
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return err
	}
	*id = ItemId(parsed)
	return nil
}
