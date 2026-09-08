package model

const ItemTypeGold = "gold"

// Bound quantities to a portable, exactly representable integer on both ends.
const MaxStackQuantity = 2147483647

func CreateGold(quantity int) *Item {
	item := NewItem("Gold", ItemTypeGold)
	item.Quantity = quantity
	return item
}

func (i *Item) IsStackable() bool { return i.Type == ItemTypeGold }

func (i *Item) ValidQuantity() bool {
	return i.Quantity > 0 && i.Quantity <= MaxStackQuantity && (i.IsStackable() || i.Quantity == 1)
}
