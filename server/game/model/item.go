package model

import (
	"fmt"
	"github.com/google/uuid"
	"math"
	"strings"
	"unicode/utf8"
)

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
	AttackMethodMelee  AttackMethod = "melee"
	AttackMethodMagic  AttackMethod = "magic"
	AttackMethodRanged AttackMethod = "ranged"
)

const ItemTypeArrow = "arrow"

const (
	SlotHead    EquipmentSlot = "head"
	SlotChest   EquipmentSlot = "chest"
	SlotLegs    EquipmentSlot = "legs"
	SlotFeet    EquipmentSlot = "feet"
	SlotWeapon  EquipmentSlot = "weapon"
	SlotOffhand EquipmentSlot = "offhand"
)

// Item contains only identity, quantity and properties unique to this instance.
type Item struct {
	Id           ItemId          `json:"id"`
	DefinitionID string          `json:"definitionId"`
	Quantity     int             `json:"quantity"`
	Properties   *ItemProperties `json:"properties,omitempty"`
}

// Extend this typed structure when gameplay introduces other individual traits.
// CombatStats, when present, replaces this instance's base combat profile.
type ItemProperties struct {
	CustomName  string           `json:"customName,omitempty"`
	CombatStats *ItemCombatStats `json:"combatStats,omitempty"`
}

func NewItem(definitionID string) *Item {
	if _, ok := itemDefinitions[definitionID]; !ok {
		return nil
	}
	return &Item{Id: NewItemId(), DefinitionID: definitionID, Quantity: 1}
}

func (i *Item) Definition() ItemDefinition {
	definition, _ := GetItemDefinition(i.DefinitionID)
	return definition
}
func (i *Item) Name() string {
	if i.Properties != nil && i.Properties.CustomName != "" {
		return i.Properties.CustomName
	}
	return itemDefinitions[i.DefinitionID].Name
}
func (i *Item) Type() string        { return itemDefinitions[i.DefinitionID].Type }
func (i *Item) RenderModel() string { return itemDefinitions[i.DefinitionID].RenderModel }
func (i *Item) CombatStats() *ItemCombatStats {
	if i.Properties != nil && i.Properties.CombatStats != nil {
		return cloneCombatStats(i.Properties.CombatStats)
	}
	return cloneCombatStats(itemDefinitions[i.DefinitionID].CombatStats)
}
func (i *Item) HasProperties() bool {
	return i.Properties != nil && (i.Properties.CustomName != "" || i.Properties.CombatStats != nil)
}
func (i *Item) Clone() *Item {
	copy := *i
	if i.Properties != nil {
		properties := *i.Properties
		properties.CombatStats = cloneCombatStats(properties.CombatStats)
		copy.Properties = &properties
	}
	return &copy
}
func cloneCombatStats(stats *ItemCombatStats) *ItemCombatStats {
	if stats == nil {
		return nil
	}
	copy := *stats
	return &copy
}

type ItemCombatStats struct {
	MinDamage        int          `json:"minDamage"`
	MaxDamage        int          `json:"maxDamage"`
	AccuracyBonus    int          `json:"accuracyBonus"`
	ArmorBonus       int          `json:"armorBonus"`
	CritBonus        float64      `json:"critBonus"`
	Range            int          `json:"range"`
	AttackSpeedTicks int          `json:"attackSpeedTicks"`
	AttackMethod     AttackMethod `json:"attackMethod"`
	WindUpTicks      int          `json:"windUpTicks"`
	TravelTicks      int          `json:"travelTicks"`
	ProjectileType   string       `json:"projectileType"`
}

func (i *Item) IsEquipable() bool { return itemDefinitions[i.DefinitionID].EquipmentSlot != "" }
func (i *Item) GetEquipmentSlot() *EquipmentSlot {
	slot := itemDefinitions[i.DefinitionID].EquipmentSlot
	if slot == "" {
		return nil
	}
	return &slot
}

func ParseEquipmentSlot(value string) (EquipmentSlot, bool) {
	switch EquipmentSlot(value) {
	case SlotHead, SlotChest, SlotLegs, SlotFeet, SlotWeapon, SlotOffhand:
		return EquipmentSlot(value), true
	default:
		return "", false
	}
}

// ValidateSaved checks item invariants independently of inventories or storage.
func (i *Item) ValidateSaved() error {
	if i == nil || i.Id == (ItemId{}) || !i.ValidQuantity() || i.DefinitionID == "" {
		return fmt.Errorf("invalid saved item")
	}
	if _, ok := itemDefinitions[i.DefinitionID]; !ok {
		return fmt.Errorf("unknown item definition %q", i.DefinitionID)
	}
	if i.Properties != nil {
		name := i.Properties.CustomName
		if name != "" && (strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 64) {
			return fmt.Errorf("invalid item custom name")
		}
		if stats := i.Properties.CombatStats; stats != nil {
			if !i.IsEquipable() {
				return fmt.Errorf("combat properties require equipment")
			}
			if err := stats.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s ItemCombatStats) Validate() error {
	if s.MinDamage < 0 || s.MaxDamage < s.MinDamage || s.Range < 0 || s.AttackSpeedTicks < 0 || s.WindUpTicks < 0 || s.TravelTicks < 0 || math.IsNaN(s.CritBonus) || math.IsInf(s.CritBonus, 0) {
		return fmt.Errorf("invalid item combat stats")
	}
	switch s.AttackMethod {
	case "", AttackMethodMelee, AttackMethodMagic, AttackMethodRanged:
	default:
		return fmt.Errorf("invalid item attack method")
	}
	return nil
}
