package component

import (
	"encoding/json"
	"fmt"
	"webscape/server/game/model"
)

func ParseItemReference(raw any) (model.ItemReference, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return model.ItemReference{}, err
	}
	var reference model.ItemReference
	if err := json.Unmarshal(data, &reference); err != nil {
		return reference, err
	}
	return reference, reference.Validate()
}

func ValidateLootable(raw any) error {
	value, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("lootable must be an object")
	}
	items, ok := value["items"].([]any)
	if !ok {
		return fmt.Errorf("lootable.items must be an array")
	}
	for index, item := range items {
		if _, err := ParseItemReference(item); err != nil {
			return fmt.Errorf("lootable item %d: %w", index, err)
		}
	}
	return nil
}
