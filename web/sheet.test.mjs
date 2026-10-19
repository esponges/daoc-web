// node web/sheet.test.mjs — the sheet, trainer and healer panels' content.
import assert from 'node:assert/strict';
import { sheetView, trainerView, healerText, eventLine, conColour, CON_COLOURS } from './sheet.js';

const stats = {
  t: 'stats', cls: 'Warrior', lv: 44, half: true,
  base: [110, 80, 45, 35, 60, 60, 60, 60], stats: [152, 99, 60, 35, 60, 60, 60, 60],
  mhp: 1004, mend: 100, ws: 312, af: 140, conlost: 2, pts: 44,
  specs: [{ key: 'Sword', nm: 'Sword', lv: 30, bonus: 3 }, { key: 'Parry', nm: 'Parry', lv: 1 }],
  ab: [{ nm: 'Evade', lv: 1, on: true }, { nm: 'Protect', lv: 3 }],
};
const v = sheetView(stats);
assert.equal(v.title, 'Warrior, level 44.5');
assert.deepEqual(v.stats[0], { name: 'Strength', value: 152, base: 110, diff: 42 });
assert.equal(v.stats.length, 8);
assert.ok(v.lines.includes('Constitution lost 2 (a healer restores it)'));
assert.deepEqual(v.specs.map((s) => s.text), ['Sword 30 (+3)', 'Parry 1']);
assert.deepEqual(v.abilities, [{ text: 'Evade', on: true }, { text: 'Protect 3', on: false }]);
assert.equal(v.points, 44);
assert.equal(sheetView(null), null);
assert.ok(!sheetView({ ...stats, conlost: 0 }).lines.some((l) => l.startsWith('Constitution lost')));

const t = trainerView({ t: 'trainer', npc: 4, nm: 'Asmund', pts: 12,
  specs: [{ key: 'Hammer', nm: 'Hammer', lv: 3, cost: 4, can: true }, { key: 'Axe', nm: 'Axe', lv: 10, cost: 11 }] });
assert.equal(t.title, 'Asmund');
assert.equal(t.rows[0].can, true);
assert.match(t.rows[0].hint, /Train to 4 for 4 points/);
assert.equal(t.rows[1].can, false);

assert.equal(healerText({ nm: 'Hilda', lost: 3, price: 352, total: 1056 }, (c) => c + 'c'),
  'Hilda can restore your 3 lost constitution for 1056c (352c a point).');

assert.equal(eventLine({ k: 'xploss', n: 3328 }).text, 'You lose 3328 experience.');
assert.equal(eventLine({ k: 'conloss', n: 1, x: 'You lose 1 constitution.' }).text, 'You lose 1 constitution.');
assert.equal(eventLine({ k: 'pray', x: 'done', n: 3328 }).text, 'You regain 3328 experience from your grave.');
assert.equal(eventLine({ k: 'train', x: 'Hammer', n: 4 }).text, 'You train Hammer to 4.');
assert.equal(eventLine({ k: 'hit' }), null);

assert.equal(conColour(undefined), undefined);
assert.equal(conColour(-3), CON_COLOURS['-3']);
assert.equal(conColour(1), '#ffa24a');
assert.equal(conColour(9), CON_COLOURS[3]);

console.log('sheet: ok');
