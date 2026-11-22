// node web/quiverwin.test.mjs
import assert from 'node:assert/strict';
import { quiverRows, quiverMove } from './quiverwin.js';

{
  const inv = { items: [{ slot: 50, nm: 'bodkin arrows', n: 12, k: 'ammo' }, { slot: 52, nm: 'broadhead arrows', n: 30, k: 'ammo' }, { slot: 3, nm: 'Ale' }] };
  const v = quiverRows(inv);
  assert.equal(v.rows.length, 4);
  assert.equal(v.rows[0].item, null);
  assert.equal(v.rows[1].text, 'bodkin arrows  x12');
  assert.equal(v.active, 1, 'the first slot with arrows is the one the server fires from');
  assert.deepEqual(quiverMove(v, 3), { from: 52, to: 50 }, 'picking a row moves its stack to the active slot');
  assert.equal(quiverMove(v, 1), null, 'the active row is already in use');
  assert.equal(quiverMove(v, 0), null, 'an empty row cannot be picked');
  assert.equal(quiverRows(null).active, -1);
  console.log('ok: the quiver rows and the move that picks one');
}
