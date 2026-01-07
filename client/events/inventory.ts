export const InventoryUpdateEventName = "inventoryUpdate";

export class InventoryUpdateEvent extends Event {
    constructor() {
        super(InventoryUpdateEventName);
    }
}

