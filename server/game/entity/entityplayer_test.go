package entity

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestNewPlayerInventoryIncludesSwordAndWoodcuttingAxe(t *testing.T) {
	components := CreatePlayerEntity(model.NewEntityId(), "Player", math.Vec2{})
	var inventory *component.CInventory
	for _, value := range components {
		if candidate, ok := value.(*component.CInventory); ok {
			inventory = candidate
			break
		}
	}
	if inventory == nil {
		t.Fatal("player has no inventory")
	}
	hasSword := false
	hasAxe := false
	for _, item := range inventory.GetAllItems() {
		hasSword = hasSword || item.Name == "Iron Sword"
		hasAxe = hasAxe || item.Name == "Woodcutting Axe" && item.Type == "axe"
	}
	if !hasSword || !hasAxe {
		t.Fatalf("starter inventory has sword=%v axe=%v", hasSword, hasAxe)
	}
}
