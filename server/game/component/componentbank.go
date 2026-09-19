package component

import (
	"fmt"
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdBank = ComponentId("bank")
const ComponentIdBanker = ComponentId("banker")
const ComponentIdBanking = ComponentId("banking")

// CBank belongs to the character; bankers only grant access to it.
type CBank struct{ storage CInventory }

func NewCBank() *CBank                               { return &CBank{storage: *NewCInventory()} }
func (*CBank) GetId() ComponentId                    { return ComponentIdBank }
func (c *CBank) Serialize() util.Json                { return c.storage.Serialize() }
func (c *CBank) GetItem(id model.ItemId) *model.Item { return c.storage.GetItem(id) }
func (c *CBank) GetAllItems() []*model.Item          { return c.storage.GetAllItems() }

// Transfer prepares both containers before committing either. Partial stacks
// receive a new identity; complete items retain their identity and properties.
func (c *CBank) Transfer(inventory *CInventory, deposit bool, id model.ItemId, quantity int) bool {
	backpack, bank := inventory.Clone(), c.storage.Clone()
	source, destination, capacity := bank, backpack, InventoryCapacity
	if deposit {
		source, destination, capacity = backpack, bank, 0
	}
	item := source.GetItem(id)
	if item == nil || quantity <= 0 || quantity > item.Quantity {
		return false
	}
	moved := item.Clone()
	moved.Quantity = quantity
	if quantity == item.Quantity {
		source.RemoveItem(id)
	} else {
		item.Quantity -= quantity
		moved.Id = model.NewItemId()
	}
	if !destination.addItem(moved, capacity) {
		return false
	}
	inventory.items, c.storage.items = backpack.items, bank.items
	return true
}

type CBanker struct{}

func (*CBanker) GetId() ComponentId   { return ComponentIdBanker }
func (*CBanker) Serialize() util.Json { return util.JObject{} }
func ParseBanker(raw any) (*CBanker, error) {
	value, ok := raw.(map[string]any)
	if !ok || len(value) != 0 {
		return nil, fmt.Errorf("banker must be an empty object")
	}
	return &CBanker{}, nil
}

type CBanking struct{ TargetEntityId model.EntityId }

func (*CBanking) GetId() ComponentId                  { return ComponentIdBanking }
func (c *CBanking) GetTargetEntityId() model.EntityId { return c.TargetEntityId }
func (c *CBanking) Serialize() util.Json {
	return util.JObject{"targetEntityId": util.JString(c.TargetEntityId.String())}
}
func (*CBanking) transientSave() {}
