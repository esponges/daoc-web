// Headless checks of the loot window's rows.
//   node web/loot.test.mjs

import { lootRows, lootHeight } from './loot.js';

let failed = 0;
const check = (name, ok, extra = '') => {
  if (!ok) {
    failed++;
    console.log(`FAIL ${name} ${extra}`);
  } else console.log(`ok   ${name}`);
};
const money = (c) => c + 'c';

const rows = lootRows({ id: 7, nm: 'frog', mo: 10, items: [{ slot: 0, nm: 'pelt', d: 'level 1', k: 'misc' }, { slot: 1, nm: 'club', k: 'weapon' }] }, money);
check('money first', rows[0].money && rows[0].slot === -1 && rows[0].name === '10c');
check('items follow with their slots', rows.length === 3 && rows[1].slot === 0 && rows[2].slot === 1);
check('detail kept', rows[1].detail === 'level 1' && rows[2].detail === '');
check('no money no money row', lootRows({ id: 1, items: [{ slot: 0, nm: 'x' }] }, money).length === 1);
check('empty corpse', lootRows({ id: 1, items: [] }, money).length === 0 && lootRows(null, money).length === 0);
check('height grows by a row', lootHeight(3) - lootHeight(2) === 36);
check('height has a minimum', lootHeight(0) === lootHeight(1));

// Take all clicks the highest slot first, money last, so slots do not shift.
const order = rows.slice().sort((a, b) => b.slot - a.slot).map((r) => r.slot);
check('take-all order', JSON.stringify(order) === '[1,0,-1]', JSON.stringify(order));

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
