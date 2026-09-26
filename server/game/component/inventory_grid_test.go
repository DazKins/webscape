package component

import (
	"encoding/json"
	"reflect"
	"testing"
	"webscape/server/game/model"
)

func TestInventoryGridMovesAndKeepsHoles(t *testing.T) {
	c := NewCInventory()
	a, b := model.NewItem("bread"), model.NewItem("ironSword")
	c.AddItem(a)
	c.AddItem(b)
	if !c.MoveItem(a.Id, 19) || c.itemAtSlot(0) != nil || c.itemAtSlot(19) != a {
		t.Fatal("move failed")
	}
	if !c.MoveItem(b.Id, 19) || c.itemAtSlot(1) != a || c.itemAtSlot(19) != b {
		t.Fatal("swap failed")
	}
	for _, slot := range []int{-1, InventoryCapacity} {
		if c.MoveItem(a.Id, slot) {
			t.Fatal("accepted invalid slot")
		}
	}
	if c.MoveItem(model.NewItemId(), 0) {
		t.Fatal("accepted unknown item")
	}
	c.RemoveItem(a.Id)
	c.AddItem(model.NewItem("bread"))
	if c.itemAtSlot(1) != nil || c.itemAtSlot(19) != b || c.itemAtSlot(0) == nil {
		t.Fatal("removal or addition shifted items")
	}
	clone := c.Clone()
	clone.MoveItem(b.Id, 18)
	if c.itemAtSlot(19) != b {
		t.Fatal("clone shares layout")
	}
}

func TestInventoryGridStacksAndTransfers(t *testing.T) {
	c := NewCInventory()
	arrows := model.NewItem("arrow")
	c.AddItem(arrows)
	c.MoveItem(arrows.Id, 19)
	c.AddItem(model.NewItem("arrow"))
	if arrows.Quantity != 2 || c.itemAtSlot(19) != arrows {
		t.Fatal("stacking moved item")
	}
	bank := NewCBank()
	if !bank.Transfer(c, true, arrows.Id, 1) || c.itemAtSlot(19).Quantity != 1 {
		t.Fatal("partial deposit moved stack")
	}
	bankItem := bank.GetAllItems()[0]
	if !bank.Transfer(c, false, bankItem.Id, 1) || c.itemAtSlot(19).Quantity != 2 {
		t.Fatal("withdraw moved stack")
	}
	gold := model.CreateGold(10)
	c.AddItem(gold)
	c.MoveItem(gold.Id, 18)
	if !c.Exchange(gold.Id, 1, model.NewItem("bread")) || c.itemAtSlot(18).Quantity != 9 || c.itemAtSlot(19) == nil {
		t.Fatal("trade lost layout")
	}
	for !c.IsFull() {
		c.AddItem(model.NewItem("bread"))
	}
	if !c.MoveItem(arrows.Id, 0) || c.itemAtSlot(0).Id != arrows.Id {
		t.Fatal("full inventory cannot swap")
	}
	if c.AddItem(model.NewItem("bread")) {
		t.Fatal("overfilled inventory")
	}
	if !c.AddItem(model.NewItem("arrow")) || c.itemAtSlot(0).Quantity != 3 {
		t.Fatal("full inventory cannot stack")
	}
}

func TestInventoryGridSaveAndLegacyMigration(t *testing.T) {
	c := NewCInventory()
	item := model.NewItem("bread")
	c.AddItem(item)
	c.MoveItem(item.Id, 19)
	saved, err := c.Save()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreComponent(ComponentIdInventory, saved)
	if err != nil || !reflect.DeepEqual(c, restored) {
		t.Fatalf("layout round trip: %v", err)
	}
	for _, version := range []int{1, 2} {
		legacy, _ := marshalSaved(savedInventory{Items: c.items}, version)
		restored, err := RestoreComponent(ComponentIdInventory, legacy)
		if err != nil || restored.(*CInventory).itemAtSlot(0).Id != item.Id {
			t.Fatalf("legacy migration: %v", err)
		}
	}
	for _, slots := range [][]int{nil, {-1}, {20}, {0, 0}} {
		items := c.items
		if len(slots) == 2 {
			items = append(append([]*model.Item{}, items...), model.NewItem("bread"))
		}
		data, _ := json.Marshal(savedInventoryGrid{Items: items, Slots: slots})
		if _, err := RestoreComponent(ComponentIdInventory, SavedComponent{Version: 3, Data: data}); err == nil {
			t.Fatalf("accepted invalid slots %v", slots)
		}
	}
}

func TestInventoryGridMerge(t *testing.T) {
	// Normal collection already merges stacks. Exercise movement defensively for
	// two compatible stacks should another inventory operation introduce them.
	c := NewCInventory()
	a, b := model.NewItem("arrow"), model.NewItem("arrow")
	c.items = []*model.Item{a, b}
	c.slots[a.Id], c.slots[b.Id] = 0, 1
	b.Quantity = model.MaxStackQuantity
	if c.MoveItem(a.Id, 1) || c.GetItemCount() != 2 {
		t.Fatal("overflow changed stacks")
	}
	b.Quantity = 2
	if !c.MoveItem(a.Id, 1) || c.HasItem(a.Id) || c.itemAtSlot(1).Quantity != 3 {
		t.Fatal("merge failed")
	}
}
