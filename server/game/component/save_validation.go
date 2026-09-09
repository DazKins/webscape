package component

import (
	"fmt"
	"strings"
	"unicode/utf8"
	"webscape/server/game/model"
)

func IdleComponents(components []Component, tick uint64) []Component {
	result := make([]Component, 0, len(components))
	for _, c := range components {
		if transient(c.GetId()) {
			continue
		}
		if c.GetId() == ComponentIdLocomotion {
			c = NewCLocomotion(LocomotionPhaseIdle, tick)
		}
		result = append(result, c)
	}
	return result
}

// ValidateSavedEntity checks the invariants needed by systems before restored
// components are admitted into the ECS. A broken save must not become a fresh game.
func ValidateSavedEntity(components []Component) error {
	byID := map[ComponentId]Component{}
	itemIDs := map[model.ItemId]bool{}
	checkItem := func(item *model.Item) error {
		if item == nil || item.Id == (model.ItemId{}) || !item.ValidQuantity() || item.Type == "" {
			return fmt.Errorf("invalid saved item")
		}
		if itemIDs[item.Id] {
			return fmt.Errorf("duplicate saved item")
		}
		itemIDs[item.Id] = true
		return nil
	}
	for _, c := range components {
		byID[c.GetId()] = c
		switch c := c.(type) {
		case *CAppearance:
			if err := ValidateAppearance(c.appearance); err != nil {
				return err
			}
		case *CPlayer:
			if strings.TrimSpace(c.name) == "" || utf8.RuneCountInString(c.name) > 24 {
				return fmt.Errorf("invalid player name")
			}
		case *CHealth:
			if c.maxHealth <= 0 || c.currentHealth > c.maxHealth {
				return fmt.Errorf("invalid health")
			}
		case *CInventory:
			if len(c.items) > InventoryCapacity {
				return fmt.Errorf("inventory exceeds capacity")
			}
			types := map[string]bool{}
			for _, item := range c.items {
				if err := checkItem(item); err != nil {
					return err
				}
				if item.IsStackable() && types[item.Type] {
					return fmt.Errorf("duplicate inventory stack")
				}
				types[item.Type] = true
			}
		case *CEquipped:
			if c.slots == nil {
				return fmt.Errorf("missing equipment slots")
			}
			for slot, item := range c.slots {
				if err := checkItem(item); err != nil {
					return err
				}
				if _, ok := model.ParseEquipmentSlot(string(slot)); !ok || item.EquipmentSlot == nil || *item.EquipmentSlot != slot {
					return fmt.Errorf("invalid equipped item")
				}
			}
		case *CShop:
			if len(c.Offers) < 1 || len(c.Offers) > 100 {
				return fmt.Errorf("invalid shop offers")
			}
			offers := map[string]bool{}
			for _, offer := range c.Offers {
				if offer.Item == nil || model.CreateShopItem(offer.ItemId) == nil || offers[offer.ItemId] || offer.BuyPrice < 1 || offer.BuyPrice > MaxShopPrice || offer.SellPrice < 1 || offer.SellPrice >= offer.BuyPrice {
					return fmt.Errorf("invalid shop offer")
				}
				offers[offer.ItemId] = true
			}
		case *CDroppedItem:
			if err := checkItem(c.Item); err != nil {
				return err
			}
		case *CSpawn:
			if c.respawnTicks < 0 || c.remainingRespawnTicks < 0 || c.childTemplateComponents == nil {
				return fmt.Errorf("invalid spawn state")
			}
		case *CWoodcuttable:
			if c.maxDurability < 1 || c.currentDurability < 0 || c.currentDurability > c.maxDurability || c.remainingRespawnTicks < 0 {
				return fmt.Errorf("invalid resource state")
			}
		case *CFishable:
			if c.catchChancePercent < 1 || c.catchChancePercent > 100 {
				return fmt.Errorf("invalid fishing chance")
			}
		}
	}
	if byID[ComponentIdPosition] == nil {
		return fmt.Errorf("entity has no position")
	}
	if byID[ComponentIdPlayer] != nil {
		for _, id := range []ComponentId{ComponentIdMetadata, ComponentIdRenderable, ComponentIdAppearance, ComponentIdHealth, ComponentIdInventory, ComponentIdEquipped, ComponentIdBaseStats, ComponentIdCombatStats, ComponentIdCombatLog, ComponentIdQuestLog, ComponentIdLocomotion} {
			if byID[id] == nil {
				return fmt.Errorf("player missing component %s", id)
			}
		}
	}
	return nil
}
