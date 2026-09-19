package component

import (
	"testing"
	"webscape/server/game/model"
)

func TestInventoryAddItemEnforcesCapacity(t *testing.T) {
	inventory := NewCInventory()

	for i := 0; i < InventoryCapacity; i++ {
		if !inventory.AddItem(model.NewItem("bread")) {
			t.Fatalf("AddItem returned false before capacity at item %d", i)
		}
	}

	if inventory.GetItemCount() != InventoryCapacity {
		t.Fatalf("inventory count = %d, want %d", inventory.GetItemCount(), InventoryCapacity)
	}
	if inventory.AvailableSlots() != 0 {
		t.Fatalf("available slots = %d, want 0", inventory.AvailableSlots())
	}
	if !inventory.IsFull() {
		t.Fatal("inventory is not full at capacity")
	}
	if inventory.AddItem(model.NewItem("bread")) {
		t.Fatal("AddItem returned true after inventory reached capacity")
	}
	if inventory.GetItemCount() != InventoryCapacity {
		t.Fatalf("overflow changed inventory count to %d", inventory.GetItemCount())
	}
}

func TestInventoryRemovesFirstItemByType(t *testing.T) {
	inventory := NewCInventory()
	bread := model.NewItem("bread")
	firstArrow := model.NewItem("arrow")
	secondArrow := model.NewItem("arrow")
	inventory.AddItem(bread)
	inventory.AddItem(firstArrow)
	inventory.AddItem(secondArrow)

	if inventory.GetItemCount() != 2 || firstArrow.Quantity != 2 {
		t.Fatal("arrows did not merge into one stack")
	}
	removed := inventory.RemoveFirstItemByDefinition(model.ItemTypeArrow)
	if removed == nil || removed.Quantity != 1 || removed.Type() != model.ItemTypeArrow || removed.Id == firstArrow.Id || !inventory.HasItem(firstArrow.Id) || firstArrow.Quantity != 1 || inventory.HasItem(secondArrow.Id) {
		t.Fatalf("removed = %#v, remaining = %#v", removed, inventory.GetAllItems())
	}
	if !inventory.HasItemDefinition(model.ItemTypeArrow) {
		t.Fatal("remaining arrow was not found")
	}
	inventory.RemoveFirstItemByDefinition(model.ItemTypeArrow)
	if inventory.HasItemDefinition(model.ItemTypeArrow) || inventory.RemoveFirstItemByDefinition(model.ItemTypeArrow) != nil {
		t.Fatal("empty arrow inventory still reports ammunition")
	}
}
