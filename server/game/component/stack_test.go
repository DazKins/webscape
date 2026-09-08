package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestGoldStacksInFullInventoryAndRejectsOverflow(t *testing.T) {
	inventory := NewCInventory()
	gold := model.CreateGold(100)
	inventory.AddItem(gold)
	for !inventory.IsFull() {
		inventory.AddItem(model.CreateBread())
	}
	if !inventory.AddItem(model.CreateGold(50)) || inventory.GetItemCount() != InventoryCapacity || gold.Quantity != 150 {
		t.Fatal("gold did not merge in a full backpack")
	}
	before := inventory.Serialize()
	for _, invalid := range []*model.Item{gold, model.CreateGold(0), model.CreateGold(-1), model.CreateGold(model.MaxStackQuantity), nil} {
		if inventory.AddItem(invalid) {
			t.Fatal("accepted invalid, duplicate or overflowing stack")
		}
		if !util.JsonEqual(before, inventory.Serialize()) {
			t.Fatal("rejected addition mutated inventory")
		}
	}
	wire := SerializeItem(gold).(util.JObject)
	if wire["quantity"] != util.JNumber(150) || wire["stackable"] != util.JBool(true) {
		t.Fatal("gold quantity is not replicated")
	}
}

func TestExchangeChecksFinalInventoryAndRollsBack(t *testing.T) {
	inventory := NewCInventory()
	inventory.AddItem(model.CreateGold(10))
	for !inventory.IsFull() {
		inventory.AddItem(model.CreateBread())
	}
	gold := inventory.FindByType(model.ItemTypeGold)
	before := inventory.Serialize()
	if inventory.Exchange(gold.Id, 9, model.CreateArrow()) || !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("full backpack purchase spent gold")
	}
	if inventory.Exchange(gold.Id, 11, model.CreateArrow()) || !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("insufficient gold purchase mutated inventory")
	}
	arrow := model.CreateArrow()
	if !inventory.Exchange(gold.Id, 10, arrow) || inventory.FindByType(model.ItemTypeGold) != nil || !inventory.IsFull() {
		t.Fatal("spending last gold did not free its slot")
	}
	if !inventory.Exchange(arrow.Id, 1, model.CreateGold(5)) || inventory.FindByType(model.ItemTypeGold).Quantity != 5 || !inventory.IsFull() {
		t.Fatal("sale did not replace sold item with gold")
	}
	inventory.FindByType(model.ItemTypeGold).Quantity = model.MaxStackQuantity
	before = inventory.Serialize()
	bread := inventory.FindByType("consumable")
	if inventory.Exchange(bread.Id, 1, model.CreateGold(1)) || !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("overflow sale lost payment")
	}
}
