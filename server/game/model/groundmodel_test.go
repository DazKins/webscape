package model

import "testing"

func TestExistingItemsHaveIndividualGroundModels(t *testing.T) {
	items := []*Item{
		CreateIronSword(), CreateWoodcuttingAxe(), CreateFishingRod(), CreateMagicStaff(), CreateWoodenBow(),
		CreateLeatherHelmet(), CreateChainmailChestplate(), CreateIronLeggings(), CreateLeatherBoots(), CreateWoodenShield(),
		CreateHealthPotion(), CreateBread(), CreateApple(), CreateIronOre(), CreateWood(), CreateLogs(), CreateStone(),
		CreateGold(100), CreateArrow(), CreateMysteriousKey(), CreateAncientScroll(), NewItem("Raw Fish", "fish"),
	}
	seen := map[string]bool{}
	for _, item := range items {
		name := item.GroundRenderModel()
		if name == "" || name == "unknownItem" || seen[name] {
			t.Errorf("%s has missing or shared ground model %q", item.Name, name)
		}
		seen[name] = true
	}
}
