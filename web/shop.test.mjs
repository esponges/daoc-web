// Headless checks of the merchant window's rows.
//   node web/shop.test.mjs

import { shopRows, canAfford, pageCount, pageRows, stepQty, priceOf, coinFields, createBuyQueue, MAX_QTY } from './shop.js';

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

// Pages, quantity, price and the buy queue.
check('pages round up and there is always one', pageCount(0, 10) === 1 && pageCount(10, 10) === 1 && pageCount(11, 10) === 2 && pageCount(126, 18) === 7 && pageCount(127, 18) === 8);
const many = Array.from({ length: 25 }, (_, i) => ({ id: 'i' + i }));
check('a page is its slice of the rows', pageRows(many, 1, 10).map((r) => r.id).join() === 'i10,i11,i12,i13,i14,i15,i16,i17,i18,i19' && pageRows(many, 2, 10).length === 5 && pageRows(many, 9, 10).length === 0);
check('the quantity stays from 1 to the most', stepQty(1, -1) === 1 && stepQty(5, 1) === 6 && stepQty(MAX_QTY, 1) === MAX_QTY && stepQty(3, -9) === 1);
check('the price is the unit price times the quantity', priceOf(rows[0], 3) === 3750 && priceOf(undefined, 3) === 0);
check('the coin fields split copper into mithril, platinum, gold, silver, copper', coinFields(1234567).join() === '0,1,23,45,67' && coinFields(0).join() === '0,0,0,0,0', coinFields(1234567).join());

// A quantity is asked for one unit at a time, and a refusal stops the rest.
{
  const sent = [];
  const q = createBuyQueue((id) => sent.push(id));
  q.start('arrows', 3);
  check('the first is asked for at once, the rest wait', sent.length === 1 && q.pending === 2);
  q.ack();
  check('each ack asks for the next', sent.length === 2 && q.pending === 1);
  q.fail();
  q.ack();
  check('a refusal drops the rest', sent.length === 2 && q.pending === 0);
  q.start('hammer', 1);
  q.ack();
  check('one unit sends once; a stray ack sends nothing', sent.length === 3 && sent[2] === 'hammer');
  q.start('x', 0);
  check('none asked for, none sent', sent.length === 3);
}

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
