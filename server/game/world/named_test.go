package world

import "testing"

func TestNamedMetadata(t *testing.T) {
	for _, tc := range []struct {
		name     string
		metadata map[string]any
		valid    bool
	}{
		{"generic", map[string]any{"name": "Tree"}, true},
		{"named", map[string]any{"name": "Rowan", "named": true}, true},
		{"disabled", map[string]any{"named": false}, true},
		{"wrong type", map[string]any{"named": "true"}, false},
		{"missing name", map[string]any{"named": true}, false},
		{"blank name", map[string]any{"named": true, "name": "  "}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNamedMetadata("npc", map[string]any{"metadata": tc.metadata})
			if (err == nil) != tc.valid {
				t.Fatalf("validation error = %v, valid = %v", err, tc.valid)
			}
		})
	}
}
