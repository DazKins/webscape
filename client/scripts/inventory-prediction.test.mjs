import assert from 'node:assert/strict';
import test from 'node:test';
import InventoryPrediction from '../game/inventoryPrediction.ts';
const item = (id, slot, extra = {}) => ({ id, slot, definitionId: id, stackable: false, quantity: 1, ...extra });

test('moves and swaps immediately without mutating authoritative items', () => {
 const prediction = new InventoryPrediction();
 const confirmed = [item('a', 0), item('b', 1)];
 prediction.enqueue('a', 1);
 assert.deepEqual(prediction.project(confirmed, 20).map(x => x.slot), [1, 0]);
 prediction.enqueue('b', 19);
 assert.deepEqual(prediction.project(confirmed, 20).map(x => x.slot), [1, 19]);
 assert.deepEqual(confirmed.map(x => x.slot), [0, 1]);
});

test('older snapshots and partial acknowledgements preserve rapid later moves', () => {
 const prediction = new InventoryPrediction();
 const confirmed = [item('a', 0), item('b', 1)];
 const first = prediction.enqueue('a', 1);
 const second = prediction.enqueue('b', 19);
 prediction.acknowledge(0);
 assert.deepEqual(prediction.project(confirmed, 20).map(x => x.slot), [1, 19]);
 prediction.acknowledge(first.sequence);
 assert.deepEqual(prediction.project([item('a', 1), item('b', 0)], 20).map(x => x.slot), [1, 19]);
 assert.equal(prediction.acknowledge(0), false);
 prediction.acknowledge(second.sequence);
 assert.deepEqual(prediction.project([item('a', 1), item('b', 19)], 20).map(x => x.slot), [1, 19]);
});

test('rejection restores confirmed state and item removal never resurrects an item', () => {
 const prediction = new InventoryPrediction();
 const move = prediction.enqueue('a', 19);
 assert.deepEqual(prediction.project([], 20), []);
 prediction.acknowledge(move.sequence);
 assert.equal(prediction.project([item('a', 0)], 20)[0].slot, 0);
});

test('pending moves rebase onto inventory changes without predicting quantities or merges', () => {
 const prediction = new InventoryPrediction();
 prediction.enqueue('a', 1);
 const confirmed = [item('a', 0, {stackable:true, definitionId:'arrow', quantity:4}), item('b', 1, {stackable:true, definitionId:'arrow', quantity:3})];
 assert.deepEqual(prediction.project(confirmed, 20), confirmed);
 prediction.enqueue('a', 19);
 const rebased = prediction.project([item('a', 0, {quantity:2}), item('new', 19)], 20);
 assert.deepEqual(rebased.map(x => [x.id, x.slot, x.quantity]), [['a', 19, 2], ['new', 1, 1]]);
});

test('disconnect clears pending work and restarts the connection sequence', () => {
 const prediction = new InventoryPrediction();
 prediction.enqueue('a', 19);
 prediction.reset();
 assert.equal(prediction.project([item('a', 0)], 20)[0].slot, 0);
 assert.equal(prediction.enqueue('a', 1).sequence, 1);
});
