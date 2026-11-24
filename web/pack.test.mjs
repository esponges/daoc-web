// Headless checks of the inventory window's geometry and money.
//   node web/pack.test.mjs

import { existsSync, readFileSync } from 'node:fs';
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
check('first row', eq(bagCell(0), { page: 0, x: 12, y: 278, w: 152, h: 16 }), JSON.stringify(bagCell(0)));
check('the pack is one column: the fifth slot is the fifth row', bagCell(4).x === 12 && bagCell(4).y === 278 + 4 * 17, JSON.stringify(bagCell(4)));
check('the eighth row sits on the art\'s last box', bagCell(7).y === 397);
check('ninth slot starts the second page', bagCell(8).page === 1 && bagCell(8).x === 12 && bagCell(8).y === 278);
check('last slot', bagCell(39).page === 4);
const cellsOnPage = [...Array(PACK_SLOTS).keys()].map(bagCell).filter((c) => c.page === 2);
check('a page holds eight distinct cells', cellsOnPage.length === 8 && new Set(cellsOnPage.map((c) => c.x + ',' + c.y)).size === 8);
check('every row fits the doll, left of the sack tabs', [...Array(PACK_SLOTS).keys()].every((i) => { const c = bagCell(i); return c.x + c.w <= 168 && c.y + c.h <= 416; }));

check('worn slots are the servers', eq(Object.keys(WORN_SLOTS).map(Number).sort(), [40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52]));
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

// With the item table's own numbers from the server, each kind has its own picture.
const pic = (o) => itemIconNumber(o);
const weapons = { sword: pic({ k: 'weapon', ot: 11, ty: 10 }), hammer: pic({ k: 'weapon', ot: 12, ty: 10 }), axe: pic({ k: 'weapon', ot: 13, ty: 10 }),
  spear: pic({ k: 'weapon', ot: 14, ty: 12 }), staff: pic({ k: 'weapon', ot: 8, ty: 12 }), bow: pic({ k: 'misc', ot: 15, ty: 13 }),
  thrown: pic({ k: 'misc', ot: 16, ty: 13 }), leftAxe: pic({ k: 'weapon', ot: 17, ty: 11 }), claws: pic({ k: 'weapon', ot: 25, ty: 11 }) };
check('every weapon family has its own picture', new Set(Object.values(weapons)).size === Object.keys(weapons).length, JSON.stringify(weapons));
check('a bow is not the hammer', weapons.bow !== 634 && weapons.bow !== weapons.hammer);
check('two-handed weapons differ from one-handed', pic({ k: 'weapon', ot: 11, ty: 12 }) !== weapons.sword && pic({ k: 'weapon', ot: 12, ty: 12 }) !== weapons.hammer && pic({ k: 'weapon', ot: 13, ty: 12 }) !== weapons.axe);
const arrow = pic({ k: 'misc', ot: 43, ty: 40 });
check('arrows are not the generic item or any weapon', arrow !== 1236 && !Object.values(weapons).includes(arrow));
check('poison, tincture and gem differ', new Set([pic({ k: 'poison', ot: 46 }), pic({ k: 'misc', ot: 47 }), pic({ k: 'misc', ot: 48 }), arrow]).size === 4);
check('shield sizes differ', new Set([1, 2, 3].map((sz) => pic({ k: 'shield', ot: 42, sz }))).size === 3 && pic({ k: 'shield', ot: 42 }) !== 705);
const torso = [32, 33, 34, 35, 36].map((ot) => pic({ k: 'armour', ot, w: 42 }));
check('armour differs by material', new Set(torso).size === 5, torso.join(','));
const slots = [42, 43, 44, 45, 46, 47].map((w) => pic({ k: 'armour', ot: 35, w }));
check('armour differs by slot', new Set(slots).size === 6, slots.join(','));
check('armour slot from the item type when the worn slot is missing', pic({ k: 'armour', ot: 35, ty: 23 }) === pic({ k: 'armour', ot: 35, w: 45 }));
check('magical items by type', new Set([26, 29, 32, 33, 35, 40].map((ty) => pic({ k: 'misc', ot: 41, ty }))).size === 6);
check('an unknown object type still draws the generic item', pic({ k: 'misc', ot: 99 }) === 1236);

// Every picture the table can give must be on a sheet the converter made.
{
  const file = new URL('./data/ui/atlantis/layout.json', import.meta.url);
  if (existsSync(file)) {
    const firsts = new Set(JSON.parse(readFileSync(file, 'utf8')).icons.map((s) => s.first));
    const all = new Set([...Object.values(weapons), arrow, 1236, 634, 705, 1204,
      ...[1, 2, 3].map((sz) => pic({ k: 'shield', ot: 42, sz })), ...torso, ...slots,
      pic({ k: 'weapon', ot: 11, ty: 12 }), pic({ k: 'weapon', ot: 12, ty: 12 }), pic({ k: 'weapon', ot: 13, ty: 12 }),
      pic({ k: 'poison', ot: 46 }), pic({ k: 'misc', ot: 47 }), pic({ k: 'misc', ot: 48 }),
      ...[26, 29, 32, 33, 35, 40].map((ty) => pic({ k: 'misc', ot: 41, ty })),
      ...[32, 33, 34, 35, 36].flatMap((ot) => [42, 43, 44, 45, 46, 47].map((w) => pic({ k: 'armour', ot, w })))]);
    check('every picture is on a converted sheet', [...all].every((n) => firsts.has(n - (n % 100))), [...all].filter((n) => !firsts.has(n - (n % 100))).join(','));
  }
}

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
