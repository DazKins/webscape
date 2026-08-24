package entity

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/util"
)

func TestCreateAuthoredEntityParsesOpenableComponent(t *testing.T) {
	components := CreateAuthoredEntity(world.WorldEntity{
		Id: "door_001",
		Components: map[string]any{
			"position": map[string]any{"x": 1, "y": 2},
			"openable": map[string]any{
				"isOpen": true,
			},
		},
	})

	openable := findOpenable(components)
	if openable == nil {
		t.Fatal("openable component was not created")
	}
	if !openable.IsOpen() {
		t.Fatal("openable isOpen = false, want true")
	}

	serialized := openable.Serialize()
	object, ok := serialized.(util.JObject)
	if !ok {
		t.Fatalf("serialized openable = %#v, want object", serialized)
	}
	if object["isOpen"] != util.JBool(true) {
		t.Fatalf("serialized isOpen = %#v, want true", object["isOpen"])
	}
}

func TestCreateAuthoredEntityPreservesRenderableOrientation(t *testing.T) {
	components := CreateAuthoredEntity(world.WorldEntity{
		Id: "door_001",
		Components: map[string]any{
			"position": map[string]any{"x": 1, "y": 2},
			"renderable": map[string]any{
				"type":        "door",
				"orientation": "east",
			},
		},
	})

	renderable := findRenderable(components)
	if renderable == nil {
		t.Fatal("renderable component was not created")
	}
	if renderable.GetOrientation() != "east" {
		t.Fatalf("orientation = %q, want east", renderable.GetOrientation())
	}

	serialized := renderable.Serialize()
	object, ok := serialized.(util.JObject)
	if !ok {
		t.Fatalf("serialized renderable = %#v, want object", serialized)
	}
	if object["orientation"] != util.JString("east") {
		t.Fatalf("serialized orientation = %#v, want east", object["orientation"])
	}
}

func TestCreateAuthoredEntityDefaultsOpenableClosed(t *testing.T) {
	components := CreateAuthoredEntity(world.WorldEntity{
		Id: "door_001",
		Components: map[string]any{
			"position": map[string]any{"x": 1, "y": 2},
			"openable": map[string]any{},
		},
	})

	openable := findOpenable(components)
	if openable == nil {
		t.Fatal("openable component was not created")
	}
	if openable.IsOpen() {
		t.Fatal("openable isOpen = true, want false")
	}
}

func TestCreateAuthoredEntityParsesRandomWalkMaxDistanceAndOrigin(t *testing.T) {
	components := CreateAuthoredEntity(world.WorldEntity{
		Id: "npc_001",
		Components: map[string]any{
			"position":   map[string]any{"x": 3, "y": 4},
			"randomwalk": map[string]any{"walkTimer": 10, "maxDistance": 6},
		},
	})

	randomWalk := findRandomWalk(components)
	if randomWalk == nil {
		t.Fatal("randomwalk component was not created")
	}
	if randomWalk.GetMaxDistance() != 6 {
		t.Fatalf("maxDistance = %d, want 6", randomWalk.GetMaxDistance())
	}
	if !randomWalk.HasOrigin() {
		t.Fatal("randomwalk origin was not set")
	}
	if randomWalk.GetOrigin().X != 3 || randomWalk.GetOrigin().Y != 4 {
		t.Fatalf("origin = %#v, want {X:3 Y:4}", randomWalk.GetOrigin())
	}
}

func TestCreateAuthoredEntityParsesWoodcuttableComponent(t *testing.T) {
	components := CreateAuthoredEntity(world.WorldEntity{
		Id: "tree_001",
		Components: map[string]any{
			"position": map[string]any{"x": 1, "y": 2},
			"woodcuttable": map[string]any{
				"maxDurability": 5,
				"respawnTicks":  60,
				"yield": map[string]any{
					"name": "Logs", "type": "material", "count": 1,
				},
			},
		},
	})

	woodcuttable := findWoodcuttable(components)
	if woodcuttable == nil {
		t.Fatal("woodcuttable component was not created")
	}
	if woodcuttable.GetMaxDurability() != 5 || woodcuttable.GetCurrentDurability() != 5 {
		t.Fatalf("durability = %d/%d, want 5/5", woodcuttable.GetCurrentDurability(), woodcuttable.GetMaxDurability())
	}
	if woodcuttable.GetRespawnTicks() != 60 {
		t.Fatalf("respawnTicks = %d, want 60", woodcuttable.GetRespawnTicks())
	}
	materialYield := woodcuttable.GetYield()
	if materialYield.Name != "Logs" || materialYield.Type != "material" || materialYield.Count != 1 {
		t.Fatalf("yield = %#v, want one Logs material", materialYield)
	}

	serialized := woodcuttable.Serialize().(util.JObject)
	if serialized["currentDurability"] != util.JNumber(5) ||
		serialized["depleted"] != util.JBool(false) ||
		serialized["remainingRespawnTicks"] != util.JNumber(0) {
		t.Fatalf("serialized woodcuttable = %#v", serialized)
	}
}

func findOpenable(components []component.Component) *component.COpenable {
	for _, comp := range components {
		if openable, ok := comp.(*component.COpenable); ok {
			return openable
		}
	}
	return nil
}

func findRenderable(components []component.Component) *component.CRenderable {
	for _, comp := range components {
		if renderable, ok := comp.(*component.CRenderable); ok {
			return renderable
		}
	}
	return nil
}

func findRandomWalk(components []component.Component) *component.CRandomWalk {
	for _, comp := range components {
		if randomWalk, ok := comp.(*component.CRandomWalk); ok {
			return randomWalk
		}
	}
	return nil
}

func findWoodcuttable(components []component.Component) *component.CWoodcuttable {
	for _, comp := range components {
		if woodcuttable, ok := comp.(*component.CWoodcuttable); ok {
			return woodcuttable
		}
	}
	return nil
}
