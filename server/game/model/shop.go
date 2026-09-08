package model

import "reflect"

// Catalog ids are stable authored-content contracts. Factories preserve equipment stats.
var shopItemFactories = map[string]func() *Item{
	"ironSword": CreateIronSword, "woodcuttingAxe": CreateWoodcuttingAxe,
	"fishingRod": CreateFishingRod, "magicStaff": CreateMagicStaff, "woodenBow": CreateWoodenBow,
	"arrow": CreateArrow, "leatherHelmet": CreateLeatherHelmet, "chainmailChestplate": CreateChainmailChestplate,
	"ironLeggings": CreateIronLeggings, "leatherBoots": CreateLeatherBoots, "woodenShield": CreateWoodenShield,
	"healthPotion": CreateHealthPotion, "bread": CreateBread, "apple": CreateApple,
	"ironOre": CreateIronOre, "wood": CreateWood, "logs": CreateLogs, "stone": CreateStone,
	"rawFish": func() *Item { return NewItem("Raw Fish", "fish") },
}

func CreateShopItem(id string) *Item {
	factory := shopItemFactories[id]
	if factory == nil {
		return nil
	}
	return factory()
}

// Only equivalent server-owned items can be sold at an offer's price.
func SameItemKind(a, b *Item) bool {
	if a == nil || b == nil {
		return false
	}
	left, right := *a, *b
	left.Id, right.Id = ItemId{}, ItemId{}
	left.Quantity, right.Quantity = 0, 0
	return reflect.DeepEqual(left, right)
}
