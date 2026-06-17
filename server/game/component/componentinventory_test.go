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
