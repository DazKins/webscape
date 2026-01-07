package model

// Test items - a mix of equipable and non-equipable items

// Equipable Items - Weapons
func CreateIronSword() *Item {
	return NewEquipableItem("Iron Sword", "weapon", SlotWeapon)
}

func CreateWoodenBow() *Item {
	return NewEquipableItem("Wooden Bow", "weapon", SlotWeapon)
}

func CreateMagicStaff() *Item {
	return NewEquipableItem("Magic Staff", "weapon", SlotWeapon)
}

// Equipable Items - Armor
func CreateLeatherHelmet() *Item {
	return NewEquipableItem("Leather Helmet", "armor", SlotHead)
}

func CreateChainmailChestplate() *Item {
	return NewEquipableItem("Chainmail Chestplate", "armor", SlotChest)
}

func CreateIronLeggings() *Item {
	return NewEquipableItem("Iron Leggings", "armor", SlotLegs)
}

func CreateLeatherBoots() *Item {
	return NewEquipableItem("Leather Boots", "armor", SlotFeet)
}

// Equipable Items - Offhand
func CreateWoodenShield() *Item {
	return NewEquipableItem("Wooden Shield", "shield", SlotOffhand)
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

