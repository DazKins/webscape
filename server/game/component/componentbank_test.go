package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestBankTransfersPreserveItemsAndQuantities(t *testing.T) {
	inventory, bank := NewCInventory(), NewCBank()
	sword := model.NewItem("ironSword")
	sword.Properties = &model.ItemProperties{CustomName: "Heirloom"}
	inventory.AddItem(sword)
	original := SerializeItem(sword)
	if !bank.Transfer(inventory, true, sword.Id, 1) || inventory.HasItem(sword.Id) || !util.JsonEqual(original, SerializeItem(bank.GetAllItems()[0])) {
		t.Fatal("deposit lost identity or properties")
	}
	if !bank.Transfer(inventory, false, sword.Id, 1) || len(bank.GetAllItems()) != 0 || !util.JsonEqual(original, SerializeItem(inventory.GetItem(sword.Id))) {
		t.Fatal("withdrawal lost identity or properties")
	}
	gold := model.CreateGold(100)
	inventory.AddItem(gold)
	if !bank.Transfer(inventory, true, gold.Id, 40) {
		t.Fatal("partial deposit failed")
	}
	stored := bank.GetAllItems()[0]
	if stored.Id == gold.Id || stored.Quantity != 40 || inventory.GetItem(gold.Id).Quantity != 60 {
		t.Fatal("split did not conserve identity and quantity")
	}
	if !bank.Transfer(inventory, true, gold.Id, 10) || len(bank.GetAllItems()) != 1 || bank.GetAllItems()[0].Quantity != 50 {
		t.Fatal("deposit did not merge")
	}
	if !bank.Transfer(inventory, false, stored.Id, 20) || inventory.GetItem(gold.Id).Quantity != 70 || bank.GetAllItems()[0].Quantity != 30 {
		t.Fatal("withdrawal did not merge")
	}
}

func TestBankTransfersAreAtomicAndRespectStackLimits(t *testing.T) {
	inventory, bank := NewCInventory(), NewCBank()
	sword := model.NewItem("ironSword")
	inventory.AddItem(sword)
	bank.Transfer(inventory, true, sword.Id, 1)
	gold := model.CreateGold(10)
	inventory.AddItem(gold)
	bank.Transfer(inventory, true, gold.Id, 5)
	for !inventory.IsFull() {
		inventory.AddItem(model.NewItem("bread"))
	}
	beforeInventory, beforeBank := inventory.Serialize(), bank.Serialize()
	for _, quantity := range []int{-1, 0, 1, 2, model.MaxStackQuantity} {
		if bank.Transfer(inventory, false, sword.Id, quantity) {
			t.Fatal("invalid or full transfer accepted")
		}
		if !util.JsonEqual(beforeInventory, inventory.Serialize()) || !util.JsonEqual(beforeBank, bank.Serialize()) {
			t.Fatal("failed transfer changed containers")
		}
	}
	storedGold := bank.GetAllItems()[1]
	if !bank.Transfer(inventory, false, storedGold.Id, 5) || inventory.GetItem(gold.Id).Quantity != 10 {
		t.Fatal("merge should succeed into full backpack")
	}
	inventory.GetItem(gold.Id).Quantity = model.MaxStackQuantity
	bank.storage.AddItem(model.CreateGold(1))
	beforeInventory, beforeBank = inventory.Serialize(), bank.Serialize()
	if bank.Transfer(inventory, false, bank.GetAllItems()[1].Id, 1) || !util.JsonEqual(beforeInventory, inventory.Serialize()) || !util.JsonEqual(beforeBank, bank.Serialize()) {
		t.Fatal("stack overflow was not atomic")
	}
}

func TestBankHasNoBackpackSlotLimit(t *testing.T) {
	inventory, bank := NewCInventory(), NewCBank()
	for range InventoryCapacity + 5 {
		item := model.NewItem("bread")
		inventory.AddItem(item)
		if !bank.Transfer(inventory, true, item.Id, 1) {
			t.Fatal("bank incorrectly limited to backpack capacity")
		}
	}
	if len(bank.GetAllItems()) != InventoryCapacity+5 {
		t.Fatal("items lost")
	}
}
