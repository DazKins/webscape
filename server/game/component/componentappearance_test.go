package component

import (
	"testing"
	"webscape/server/util"
)

var testAppearance = Appearance{
	SkinTone:      "tan",
	HairStyle:     "swept",
	HairColor:     "chestnut",
	TunicColor:    "forest",
	TrousersColor: "charcoal",
	ShoeColor:     "darkBrown",
}

func TestAppearanceSerializesExactWireShape(t *testing.T) {
	appearance, err := NewCAppearance(testAppearance)
	if err != nil {
		t.Fatal(err)
	}
	if appearance.GetId() != ComponentIdAppearance {
		t.Fatalf("component id = %q, want %q", appearance.GetId(), ComponentIdAppearance)
	}
	want := util.JObject{
		"skinTone": util.JString("tan"), "hairStyle": util.JString("swept"),
		"hairColor": util.JString("chestnut"), "tunicColor": util.JString("forest"),
		"trousersColor": util.JString("charcoal"), "shoeColor": util.JString("darkBrown"),
	}
	if !util.JsonEqual(appearance.Serialize(), want) {
		t.Fatalf("serialized appearance = %#v, want %#v", appearance.Serialize(), want)
	}
}

func TestParseAppearanceRejectsIncompleteUnknownAndInvalidValues(t *testing.T) {
	valid := map[string]any{
		"skinTone": "tan", "hairStyle": "swept", "hairColor": "chestnut",
		"tunicColor": "forest", "trousersColor": "charcoal", "shoeColor": "darkBrown",
	}
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing", mutate: func(value map[string]any) { delete(value, "shoeColor") }},
		{name: "unknown value", mutate: func(value map[string]any) { value["hairStyle"] = "mohawk" }},
		{name: "unknown field", mutate: func(value map[string]any) { value["hatColor"] = "red" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := make(map[string]any, len(valid))
			for key, item := range valid {
				value[key] = item
			}
			test.mutate(value)
			if _, err := ParseAppearance(value); err == nil {
				t.Fatal("ParseAppearance accepted invalid value")
			}
		})
	}
}

func TestGeneratedAppearancesUseCategoryPalettesAndStableAuthoredFallback(t *testing.T) {
	for range 100 {
		if err := ValidateAppearance(RandomAppearance()); err != nil {
			t.Fatalf("random appearance was invalid: %v", err)
		}
	}
	first := DeterministicAppearance("npc_001")
	for range 10 {
		if next := DeterministicAppearance("npc_001"); next != first {
			t.Fatalf("deterministic appearance changed: %#v then %#v", first, next)
		}
	}
	if err := ValidateAppearance(first); err != nil {
		t.Fatalf("deterministic appearance was invalid: %v", err)
	}
}
