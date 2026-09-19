package component

import (
	"fmt"
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdShop = ComponentId("shop")
const ComponentIdTrading = ComponentId("trading")
const MaxShopPrice = 1000000

type ShopOffer struct {
	DefinitionID string `json:"definitionId"`
	BuyPrice     int    `json:"buyPrice"`
	SellPrice    int    `json:"sellPrice"`
}

type CShop struct{ Offers []ShopOffer }

func (c *CShop) GetId() ComponentId { return ComponentIdShop }
func (c *CShop) Serialize() util.Json {
	offers := util.JArray{}
	for _, offer := range c.Offers {
		item := SerializeItemDefinition(offer.DefinitionID)
		offers = append(offers, util.JObject{"definitionId": util.JString(offer.DefinitionID), "buyPrice": util.JNumber(offer.BuyPrice), "sellPrice": util.JNumber(offer.SellPrice), "item": item})
	}
	return util.JObject{"offers": offers}
}

// ParseShop is shared by world validation and authored entity conversion.
func ParseShop(raw any) (*CShop, error) {
	value, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("shop must be an object")
	}
	if len(value) != 1 {
		return nil, fmt.Errorf("shop must contain only offers")
	}
	offers, ok := value["offers"].([]any)
	if !ok || len(offers) < 1 || len(offers) > 100 {
		return nil, fmt.Errorf("shop.offers must contain 1 to 100 offers")
	}
	shop := &CShop{}
	seen := map[string]bool{}
	for _, rawOffer := range offers {
		offer, ok := rawOffer.(map[string]any)
		if !ok || len(offer) != 3 {
			return nil, fmt.Errorf("shop offer must contain definitionId, buyPrice and sellPrice")
		}
		id, _ := offer["definitionId"].(string)
		if _, canonical := offer["definitionId"]; !canonical {
			id, _ = offer["itemId"].(string)
		}
		_, known := model.GetItemDefinition(id)
		if !known || id == "gold" || seen[id] {
			return nil, fmt.Errorf("unknown or duplicate shop definitionId %q", id)
		}
		buy, buyOK := shopPrice(offer["buyPrice"])
		sell, sellOK := shopPrice(offer["sellPrice"])
		if !buyOK || !sellOK || sell >= buy {
			return nil, fmt.Errorf("shop prices must be integers from 1 to %d with sellPrice below buyPrice", MaxShopPrice)
		}
		seen[id] = true
		shop.Offers = append(shop.Offers, ShopOffer{DefinitionID: id, BuyPrice: buy, SellPrice: sell})
	}
	return shop, nil
}

func shopPrice(raw any) (int, bool) {
	value, ok := raw.(float64)
	if !ok || value < 1 || value > MaxShopPrice {
		return 0, false
	}
	price := int(value)
	return price, float64(price) == value
}

type CTrading struct{ TargetEntityId model.EntityId }

func (c *CTrading) GetId() ComponentId                { return ComponentIdTrading }
func (c *CTrading) GetTargetEntityId() model.EntityId { return c.TargetEntityId }
func (c *CTrading) Serialize() util.Json {
	return util.JObject{"targetEntityId": util.JString(c.TargetEntityId.String())}
}

// transientSave excludes connection activity from durable saves.
func (*CTrading) transientSave() {}
