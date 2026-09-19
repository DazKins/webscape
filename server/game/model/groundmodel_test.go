package model

import "testing"

func TestExistingItemsHaveIndividualGroundModels(t *testing.T) {
	items := []*Item{
		NewItem("ironSword"), NewItem("woodcuttingAxe"), NewItem("fishingRod"), NewItem("magicStaff"), NewItem("woodenBow"),
		NewItem("leatherHelmet"), NewItem("chainmailChestplate"), NewItem("ironLeggings"), NewItem("leatherBoots"), NewItem("woodenShield"),
		NewItem("healthPotion"), NewItem("bread"), NewItem("apple"), NewItem("ironOre"), NewItem("wood"), NewItem("logs"), NewItem("stone"),
		CreateGold(100), NewItem("arrow"), NewItem("mysteriousKey"), NewItem("ancientScroll"), NewItem("rawFish"),
	}
	seen := map[string]bool{}
	for _, item := range items {
		name := item.GroundRenderModel()
		if name == "" || name == "unknownItem" || seen[name] {
			t.Errorf("%s has missing or shared ground model %q", item.Name(), name)
		}
		seen[name] = true
	}
}
