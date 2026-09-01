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

type AttackMethod string

const (
	AttackMethodMelee AttackMethod = "melee"
	AttackMethodMagic AttackMethod = "magic"
)

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
	RenderModel   string
	EquipmentSlot *EquipmentSlot // nil if item is not equipable
	CombatStats   *ItemCombatStats
}

func NewItem(name string, itemType string) *Item {
	return &Item{
		Id:            NewItemId(),
		Name:          name,
		Type:          itemType,
		RenderModel:   "",
		EquipmentSlot: nil,
		CombatStats:   nil,
	}
}

type ItemCombatStats struct {
	MinDamage        int
	MaxDamage        int
	AccuracyBonus    int
	ArmorBonus       int
	CritBonus        float64
	Range            int
	AttackSpeedTicks int
	AttackMethod     AttackMethod
	WindUpTicks      int
	TravelTicks      int
	ProjectileType   string
}

func NewEquipableItem(name string, itemType string, renderModel string, slot EquipmentSlot, combatStats *ItemCombatStats) *Item {
	return &Item{
		Id:            NewItemId(),
		Name:          name,
		Type:          itemType,
		RenderModel:   renderModel,
		EquipmentSlot: &slot,
		CombatStats:   combatStats,
	}
}

func (i *Item) IsEquipable() bool {
	return i.EquipmentSlot != nil
}

func (i *Item) GetEquipmentSlot() *EquipmentSlot {
	return i.EquipmentSlot
}

func ParseEquipmentSlot(value string) (EquipmentSlot, bool) {
	switch EquipmentSlot(value) {
	case SlotHead, SlotChest, SlotLegs, SlotFeet, SlotWeapon, SlotOffhand:
		return EquipmentSlot(value), true
	default:
		return "", false
	}
}
