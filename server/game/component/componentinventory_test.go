package component

import (
	"fmt"
	"testing"
	"webscape/server/game/model"
)

func TestInventoryAddItemEnforcesCapacity(t *testing.T) {
	inventory := NewCInventory()

	for i := 0; i < InventoryCapacity; i++ {
		if !inventory.AddItem(model.NewItem(fmt.Sprintf("Item %d", i), "test")) {
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
	if inventory.AddItem(model.NewItem("Overflow", "test")) {
		t.Fatal("AddItem returned true after inventory reached capacity")
	}
	if inventory.GetItemCount() != InventoryCapacity {
		t.Fatalf("overflow changed inventory count to %d", inventory.GetItemCount())
	}
}

func TestInventoryRemovesFirstItemByType(t *testing.T) {
	inventory := NewCInventory()
	bread := model.CreateBread()
	firstArrow := model.CreateArrow()
	secondArrow := model.CreateArrow()
	inventory.AddItem(bread)
	inventory.AddItem(firstArrow)
	inventory.AddItem(secondArrow)

	removed := inventory.RemoveFirstItemByType(model.ItemTypeArrow)
	if removed != firstArrow || inventory.HasItem(firstArrow.Id) || !inventory.HasItem(secondArrow.Id) {
		t.Fatalf("removed = %#v, remaining = %#v", removed, inventory.GetAllItems())
	}
	if !inventory.HasItemType(model.ItemTypeArrow) {
		t.Fatal("second arrow was not found")
	}
	inventory.RemoveFirstItemByType(model.ItemTypeArrow)
	if inventory.HasItemType(model.ItemTypeArrow) || inventory.RemoveFirstItemByType(model.ItemTypeArrow) != nil {
		t.Fatal("empty arrow inventory still reports ammunition")
	}
}
