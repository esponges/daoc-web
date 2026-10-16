// Headless checks of the merchant window's rows.
//   node web/shop.test.mjs

import { shopRows, shopHeight, canAfford, ROW_H, MAX_ROWS } from './shop.js';

let failed = 0;
const check = (name, ok, extra = '') => {
  if (!ok) {
    failed++;
    console.log(`FAIL ${name} ${extra}`);
  } else console.log(`ok   ${name}`);
};
const money = (c) => c + 'c';

const shop = { npc: 9, nm: 'Smith', items: [
  { slot: 0, id: 'hammer', nm: 'bronze hammer', d: '1.8 dps', pr: 1250, k: 'weapon', line: 'Hammer' },
  { slot: 1, id: 'shield', nm: 'round shield', k: 'shield' },
] };
const rows = shopRows(shop, money);
check('one row per item, in order', rows.length === 2 && rows[0].id === 'hammer' && rows[1].id === 'shield');
check('price and its text', rows[0].price === 1250 && rows[0].priceText === '1250c');
check('a row with no price is free, with no detail', rows[1].price === 0 && rows[1].detail === '');
check('buying names the item id, not the row', rows[1].index === 1 && rows[1].id === 'shield');
check('no shop, no rows', shopRows(null, money).length === 0 && shopRows({ items: [] }, money).length === 0);
check('affordable', canAfford(rows[0], 1250) && !canAfford(rows[0], 1249) && canAfford(rows[1], 0));
check('height grows by a row', shopHeight(3) - shopHeight(2) === ROW_H);
check('height has a minimum', shopHeight(0) === shopHeight(1));
check('height stops at the visible rows, the rest scroll', shopHeight(126) === shopHeight(MAX_ROWS));

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
