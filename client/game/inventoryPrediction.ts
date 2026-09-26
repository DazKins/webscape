export type InventoryLayoutItem = {
  id: string;
  slot?: number;
  definitionId: string;
  stackable: boolean;
};

type PendingMove = { itemId: string; slot: number; sequence: number };

// This queue changes presentation only. Confirmed entities and quantities stay
// untouched; every tick rebases unacknowledged moves onto the current inventory.
export default class InventoryPrediction {
  private sequence = 0;
  private acknowledged = 0;
  private pending: PendingMove[] = [];

  enqueue(itemId: string, slot: number): PendingMove {
    const move = { itemId, slot, sequence: ++this.sequence };
    this.pending.push(move);
    return move;
  }

  acknowledge(sequence: number): boolean {
    if (!Number.isSafeInteger(sequence) || sequence <= this.acknowledged || sequence > this.sequence) return false;
    this.acknowledged = sequence;
    this.pending = this.pending.filter(move => move.sequence > sequence);
    return true;
  }

  project<T extends InventoryLayoutItem>(confirmed: readonly T[], capacity: number): T[] {
    const items = confirmed.map(item => ({ ...item }));
    for (const move of this.pending) {
      const item = items.find(item => item.id === move.itemId);
      if (!item || item.slot === undefined || move.slot < 0 || move.slot >= capacity) continue;
      const target = items.find(item => item.slot === move.slot);
      // Stack merging changes quantities, so wait for the server for that case.
      if (target && target.id !== item.id && item.stackable && target.stackable && item.definitionId === target.definitionId) continue;
      if (target) target.slot = item.slot;
      item.slot = move.slot;
    }
    return items;
  }

  reset() {
    this.sequence = 0;
    this.acknowledged = 0;
    this.pending = [];
  }
}
