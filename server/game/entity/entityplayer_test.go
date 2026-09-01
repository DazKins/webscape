package entity

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestNewPlayerInventoryIncludesStarterTools(t *testing.T) {
	components := CreatePlayerEntity(model.NewEntityId(), "Player", math.Vec2{}, 0)
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
	hasFishingRod := false
	hasMagicStaff := false
	for _, item := range inventory.GetAllItems() {
		hasSword = hasSword || item.Name == "Iron Sword"
		hasAxe = hasAxe || item.Name == "Woodcutting Axe" && item.Type == "axe"
		hasFishingRod = hasFishingRod || item.Name == "Fishing Rod" && item.Type == "fishingRod" && item.RenderModel == "fishingRod"
		hasMagicStaff = hasMagicStaff || item.Name == "Magic Staff" && item.RenderModel == "magicStaff"
	}
	if !hasSword || !hasAxe || !hasFishingRod || !hasMagicStaff {
		t.Fatalf("starter inventory has sword=%v axe=%v fishingRod=%v magicStaff=%v", hasSword, hasAxe, hasFishingRod, hasMagicStaff)
	}
}

func TestPlayerReceivesValidAppearance(t *testing.T) {
	components := CreatePlayerEntity(model.NewEntityId(), "Player", math.Vec2{}, 0)
	appearance := findAppearance(components)
	if appearance == nil {
		t.Fatal("player has no appearance")
	}
	if err := component.ValidateAppearance(appearance.GetAppearance()); err != nil {
		t.Fatalf("player appearance is invalid: %v", err)
	}
}
