// node web/trainwin.test.mjs
import assert from 'node:assert/strict';
import { trainRows, pointsText } from './trainwin.js';
import { trainerView } from './sheet.js';

{
  const v = trainerView({ nm: 'Trainer', pts: 4, specs: [{ key: 'a', nm: 'Sword', lv: 2, cost: 3, can: true }, { key: 'b', nm: 'Shield', lv: 5, cost: 6, can: false }] });
  const rows = trainRows(v);
  assert.equal(rows.length, 2);
  assert.equal(rows[0].text, 'Sword 2   3 pts');
  assert.equal(rows[0].on, true);
  assert.equal(rows[1].on, false, 'a spec that cannot be trained is dimmed');
  assert.ok(rows[1].title.includes('Needs 6 points'));
  assert.equal(rows[0].key, 'a');
  assert.equal(pointsText(4), '4 points to spend');
  console.log('ok: a trainer offer is rows with their cost, and the points left');
}
