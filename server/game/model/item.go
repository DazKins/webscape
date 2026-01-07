package model

import "github.com/google/uuid"

type ItemId uuid.UUID

func NewItemId() ItemId {
	return ItemId(uuid.New())
}

func (i ItemId) String() string {
	return uuid.UUID(i).String()
}

type EquipmentSlot string

const (
	SlotHead    EquipmentSlot = "head"
	SlotChest   EquipmentSlot = "chest"
	SlotLegs    EquipmentSlot = "legs"
	SlotFeet    EquipmentSlot = "feet"
	SlotWeapon  EquipmentSlot = "weapon"
	SlotOffhand EquipmentSlot = "offhand"
)

type Item struct {
	Id            ItemId
	Name          string
	Type          string
	EquipmentSlot *EquipmentSlot // nil if item is not equipable
}

func NewItem(name string, itemType string) *Item {
	return &Item{
		Id:            NewItemId(),
		Name:          name,
		Type:          itemType,
		EquipmentSlot: nil,
	}
}

func NewEquipableItem(name string, itemType string, slot EquipmentSlot) *Item {
	return &Item{
		Id:            NewItemId(),
		Name:          name,
		Type:          itemType,
		EquipmentSlot: &slot,
	}
}

func (i *Item) IsEquipable() bool {
	return i.EquipmentSlot != nil
}

func (i *Item) GetEquipmentSlot() *EquipmentSlot {
	return i.EquipmentSlot
}
