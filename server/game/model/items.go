package model

// Test items - a mix of equipable and non-equipable items

// Equipable Items - Weapons
func CreateIronSword() *Item {
	return NewEquipableItem("Iron Sword", "weapon", SlotWeapon, &ItemCombatStats{
		MinDamage:        6,
		MaxDamage:        10,
		AccuracyBonus:    5,
		ArmorBonus:       0,
		CritBonus:        0.05,
		Range:            1,
		AttackSpeedTicks: 2,
	})
}

func CreateWoodenBow() *Item {
	return NewEquipableItem("Wooden Bow", "weapon", SlotWeapon, &ItemCombatStats{
		MinDamage:        4,
		MaxDamage:        7,
		AccuracyBonus:    6,
		ArmorBonus:       0,
		CritBonus:        0.03,
		Range:            3,
		AttackSpeedTicks: 3,
	})
}

func CreateMagicStaff() *Item {
	return NewEquipableItem("Magic Staff", "weapon", SlotWeapon, &ItemCombatStats{
		MinDamage:        5,
		MaxDamage:        9,
		AccuracyBonus:    4,
		ArmorBonus:       0,
		CritBonus:        0.08,
		Range:            2,
		AttackSpeedTicks: 3,
	})
}

// Equipable Items - Armor
func CreateLeatherHelmet() *Item {
	return NewEquipableItem("Leather Helmet", "armor", SlotHead, &ItemCombatStats{
		MinDamage:        0,
		MaxDamage:        0,
		AccuracyBonus:    0,
		ArmorBonus:       2,
		CritBonus:        0,
		Range:            1,
		AttackSpeedTicks: 0,
	})
}

func CreateChainmailChestplate() *Item {
	return NewEquipableItem("Chainmail Chestplate", "armor", SlotChest, &ItemCombatStats{
		MinDamage:        0,
		MaxDamage:        0,
		AccuracyBonus:    0,
		ArmorBonus:       4,
		CritBonus:        0,
		Range:            1,
		AttackSpeedTicks: 0,
	})
}

func CreateIronLeggings() *Item {
	return NewEquipableItem("Iron Leggings", "armor", SlotLegs, &ItemCombatStats{
		MinDamage:        0,
		MaxDamage:        0,
		AccuracyBonus:    0,
		ArmorBonus:       3,
		CritBonus:        0,
		Range:            1,
		AttackSpeedTicks: 0,
	})
}

func CreateLeatherBoots() *Item {
	return NewEquipableItem("Leather Boots", "armor", SlotFeet, &ItemCombatStats{
		MinDamage:        0,
		MaxDamage:        0,
		AccuracyBonus:    0,
		ArmorBonus:       1,
		CritBonus:        0,
		Range:            1,
		AttackSpeedTicks: 0,
	})
}

// Equipable Items - Offhand
func CreateWoodenShield() *Item {
	return NewEquipableItem("Wooden Shield", "shield", SlotOffhand, &ItemCombatStats{
		MinDamage:        0,
		MaxDamage:        0,
		AccuracyBonus:    0,
		ArmorBonus:       3,
		CritBonus:        0,
		Range:            1,
		AttackSpeedTicks: 0,
	})
}

// Non-Equipable Items - Consumables
func CreateHealthPotion() *Item {
	return NewItem("Health Potion", "consumable")
}

func CreateBread() *Item {
	return NewItem("Bread", "consumable")
}

func CreateApple() *Item {
	return NewItem("Apple", "consumable")
}

// Non-Equipable Items - Materials
func CreateIronOre() *Item {
	return NewItem("Iron Ore", "material")
}

func CreateWood() *Item {
	return NewItem("Wood", "material")
}

func CreateStone() *Item {
	return NewItem("Stone", "material")
}

// Non-Equipable Items - Quest Items
func CreateMysteriousKey() *Item {
	return NewItem("Mysterious Key", "quest")
}

func CreateAncientScroll() *Item {
	return NewItem("Ancient Scroll", "quest")
}
