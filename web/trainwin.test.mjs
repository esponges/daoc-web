// node web/trainwin.test.mjs
import assert from 'node:assert/strict';
import { trainRows, pointsText, trainAction } from './trainwin.js';
import { trainerView, infoLines } from './sheet.js';

{
  const v = trainerView({ nm: 'Trainer', pts: 4, specs: [{ key: 'a', nm: 'Sword', lv: 2, cost: 3, can: true }, { key: 'b', nm: 'Shield', lv: 5, cost: 6, can: false }] });
  const rows = trainRows(v);
  assert.equal(rows.length, 2);
  assert.equal(rows[0].text, 'Sword 2');
  assert.equal(rows[0].right, '3', 'the cost is its own column');
  assert.equal(rows[0].on, true);
  assert.equal(rows[1].on, false, 'a spec that cannot be trained is dimmed');
  assert.ok(rows[1].title.includes('Needs 6 points; you have 4.'));
  assert.equal(rows[0].key, 'a');
  assert.equal(pointsText(4), '4 points');
  assert.equal(pointsText(1), '1 point');
  console.log('ok: a trainer offer is rows with their cost, and the points left');

  assert.deepEqual(trainAction(rows[0]), { train: 'a' });
  assert.deepEqual(trainAction(undefined), { say: 'Choose a specialization to train first.' });
  assert.deepEqual(trainAction(rows[1]), { say: 'Needs 6 points; you have 4.' });
  console.log('ok: TRAIN trains a trainable row and otherwise says why not');
}
{
  const v = trainerView({ nm: 'Ronja', pts: 12, say: 'Welcome, Hrafn.', specs: [
    { key: 'Hand to Hand', nm: 'Hand to Hand', lv: 4, cost: 5, can: true,
      gives: [{ lv: 5, nm: 'Savage Rage', k: 'style' }, { lv: 7, nm: 'Fist Wraps', k: 'ability' }] },
    { key: 'Parry', nm: 'Parry', lv: 4, cost: 5, can: true }] });
  assert.equal(v.say, 'Welcome, Hrafn.');
  const rows = trainRows(v);
  assert.match(rows[0].title, /Next: 5 Savage Rage \(style\)\./);
  assert.deepEqual(infoLines(rows[0].view), ['Hand to Hand gives: 5 Savage Rage (style), 7 Fist Wraps (ability).']);
  assert.deepEqual(infoLines(rows[1].view), ['Parry: nothing more to gain from training.']);
  assert.deepEqual(infoLines(undefined), ['Choose a specialization first.']);
  console.log('ok: INFO and the tooltip say what the next levels give');
}
