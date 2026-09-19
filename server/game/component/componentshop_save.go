package component

import (
	"fmt"
	"webscape/server/game/model"
)

type savedShop struct {
	Offers []ShopOffer `json:"offers"`
}

func (c *CShop) Save() (SavedComponent, error) {
	return marshalSaved(savedShop{Offers: c.Offers}, 2)
}
func init() {
	registerComponentRestore(ComponentIdShop, 1, restoreShop)
	registerComponentRestore(ComponentIdShop, 2, restoreShop)
}
func restoreShop(saved SavedComponent) (Component, error) {
	if saved.Version == 1 {
		var legacy struct {
			Offers []struct {
				ItemId              string
				BuyPrice, SellPrice int
				Item                *model.Item
			} `json:"offers"`
		}
		if err := decodeSaved(saved.Data, &legacy); err != nil {
			return nil, err
		}
		shop := &CShop{}
		for _, offer := range legacy.Offers {
			if offer.Item == nil || offer.Item.DefinitionID != offer.ItemId || offer.Item.HasProperties() {
				return nil, fmt.Errorf("invalid legacy shop offer")
			}
			shop.Offers = append(shop.Offers, ShopOffer{DefinitionID: offer.ItemId, BuyPrice: offer.BuyPrice, SellPrice: offer.SellPrice})
		}
		return shop, nil
	}
	var s savedShop
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CShop{Offers: s.Offers}, nil
}

func (c *CShop) ValidateSaved(ctx SaveContext) error {
	if len(c.Offers) < 1 || len(c.Offers) > 100 {
		return fmt.Errorf("invalid shop offers")
	}
	offers := map[string]bool{}
	for _, offer := range c.Offers {
		if !validShopDefinition(offer.DefinitionID) || offers[offer.DefinitionID] || offer.BuyPrice < 1 || offer.BuyPrice > MaxShopPrice || offer.SellPrice < 1 || offer.SellPrice >= offer.BuyPrice {
			return fmt.Errorf("invalid shop offer")
		}
		offers[offer.DefinitionID] = true
	}
	return nil
}

func validShopDefinition(id string) bool {
	_, ok := model.GetItemDefinition(id)
	return ok && id != "gold"
}
