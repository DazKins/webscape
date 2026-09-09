package component

import (
	"fmt"
	"webscape/server/game/model"
)

type savedShop struct {
	Offers []ShopOffer `json:"offers"`
}

func (c *CShop) Save() (SavedComponent, error) {
	return marshalSaved(savedShop{Offers: c.Offers})
}
func init() { registerComponentRestore(ComponentIdShop, 1, restoreShop) }
func restoreShop(saved SavedComponent) (Component, error) {
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
		if offer.Item == nil || model.CreateShopItem(offer.ItemId) == nil || offers[offer.ItemId] || offer.BuyPrice < 1 || offer.BuyPrice > MaxShopPrice || offer.SellPrice < 1 || offer.SellPrice >= offer.BuyPrice {
			return fmt.Errorf("invalid shop offer")
		}
		offers[offer.ItemId] = true
	}
	return nil
}
