// Headless checks of the inventory window's geometry and money.
//   node web/pack.test.mjs

import { bagCell, coins, WORN_SLOTS, PACK_SLOTS, PAGE_SLOTS } from './pack.js';
import { itemIconNumber } from './uiwin.js';

let failed = 0;
const check = (name, ok, extra = '') => {
  if (!ok) {
    failed++;
    console.log(`FAIL ${name} ${extra}`);
  } else console.log(`ok   ${name}`);
};
const eq = (a, b) => JSON.stringify(a) === JSON.stringify(b);

check('forty pack slots in five pages', PACK_SLOTS === 40 && PACK_SLOTS / PAGE_SLOTS === 5);
check('first cell', eq(bagCell(0), { page: 0, x: 14, y: 282 }));
check('fifth cell wraps to the second row', eq(bagCell(4), { page: 0, x: 14, y: 320 }), JSON.stringify(bagCell(4)));
check('ninth slot starts the second page', eq(bagCell(8), { page: 1, x: 14, y: 282 }));
check('last slot', bagCell(39).page === 4);
const cellsOnPage = [...Array(PACK_SLOTS).keys()].map(bagCell).filter((c) => c.page === 2);
check('a page holds eight distinct cells', cellsOnPage.length === 8 && new Set(cellsOnPage.map((c) => c.x + ',' + c.y)).size === 8);
check('every cell fits the doll', [...Array(PACK_SLOTS).keys()].every((i) => { const c = bagCell(i); return c.x + 32 <= 172 && c.y + 32 <= 416; }));

check('worn slots are the servers', eq(Object.keys(WORN_SLOTS).map(Number).sort(), [40, 41, 42, 43, 44, 45, 46, 47]));
const rects = Object.values(WORN_SLOTS);
const overlap = (a, b) => a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h;
check('worn slots do not overlap', rects.every((a, i) => rects.every((b, j) => i === j || !overlap(a, b))));

check('no money', eq(coins(0), [0, 0, 0, 0, 0]));
check('copper and silver', eq(coins(1234), [0, 0, 0, 12, 34]));
check('gold', eq(coins(10000 * 5 + 250 + 7), [0, 0, 5, 2, 57]));
check('platinum', eq(coins(1000000 * 3), [0, 3, 0, 0, 0]));
check('negative is zero', eq(coins(-5), [0, 0, 0, 0, 0]));

check('hammer picture', itemIconNumber({ k: 'weapon', line: 'Hammer' }) === 634);
check('unknown weapon line still draws', itemIconNumber({ k: 'weapon' }) === 634);
check('armour by worn slot', itemIconNumber({ k: 'armour', w: 45 }) === 1225);
check('shield and misc', itemIconNumber({ k: 'shield' }) === 705 && itemIconNumber({ k: 'misc' }) === 1236);

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
