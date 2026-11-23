// Headless checks of the trade, friends and duel pure parts.
//   node web/social.test.mjs
import { coin, tradeView, friendsView, MAX_OFFER } from './social.js';

let failed = 0;
const ok = (cond, name, extra = '') => {
  if (!cond) { failed++; console.log(`FAIL ${name} ${extra}`); } else console.log(`ok   ${name}`);
};

ok(coin(0) === '0c' && coin(5) === '5c' && coin(100) === '1s' && coin(10250) === '1g 2s 50c', 'money is written in gold, silver and copper', coin(10250));

const m = {
  with: 'Bjorn',
  mine: { items: [{ slot: 3, nm: 'bronze axe', n: 1 }], mo: 0, ok: true },
  theirs: { items: [{ slot: 1, nm: 'wolf pelt', n: 3 }], mo: 500 },
};
const v = tradeView(m);
ok(v.title === 'Trade with Bjorn', 'the trade is titled with the other player');
ok(v.mine.items[0].text === 'bronze axe' && v.theirs.items[0].text === 'wolf pelt x3', 'items show their counts');
ok(v.theirs.moneyText === '5s', 'their money is written out');
ok(/Waiting for Bjorn/.test(v.status), 'after your accept it says it is waiting');
ok(/has accepted/.test(tradeView({ ...m, mine: { ...m.mine, ok: false }, theirs: { ...m.theirs, ok: true } }).status), 'their accept is shown');
ok(tradeView(null) === null, 'no trade, no view');
const many = { with: 'B', mine: { items: Array.from({ length: MAX_OFFER }, (_, i) => ({ slot: i, nm: 'x', n: 1 })) }, theirs: { items: [] } };
ok(tradeView(many).full, 'a full offer is marked');

const f = friendsView({ list: [{ nm: 'Sigrid' }, { nm: 'Bjorn', on: true }, { nm: 'Astrid', on: true }] });
ok(f.rows.map((r) => r.name).join() === 'Astrid,Bjorn,Sigrid' && f.online === 2, 'friends in the game come first, then by name');
ok(f.rows[0].text.includes('(in game)'), 'an online friend is marked');
ok(friendsView(null).rows.length === 0, 'no list, no rows');
if (failed) process.exit(1);
