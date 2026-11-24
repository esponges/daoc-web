// Headless checks of the crafting window's pure parts.
//   node web/craft.test.mjs
import { craftView, matLine } from './craft.js';

let failed = 0;
const ok = (cond, name) => {
  if (!cond) { failed++; console.log(`FAIL ${name}`); } else console.log(`ok   ${name}`);
};

const m = {
  t: 'craft', pr: 1, pn: 'Weaponcrafting',
  skills: [{ id: 1, nm: 'Weaponcrafting', v: 12, cap: 1100 }, { id: 14, nm: 'Woodworking', v: 3, cap: 12 }],
  recipes: [
    { id: 'dagger', nm: 'bronze dagger', sk: 1, lv: 15, ok: 85, s: 3, m: [{ id: 'bar', nm: 'bronze bar', n: 3, have: 5 }] },
    { id: 'sword', nm: 'bronze sword', sk: 1, lv: 40, ok: 50, s: 4, m: [{ id: 'bar', nm: 'bronze bar', n: 10, have: 5 }] },
    { id: 'board', nm: 'rowan board', sk: 14, lv: 0, ok: 100, s: 3, m: [{ id: 'log', nm: 'rowan log', n: 1, have: 1 }] },
  ],
};
const v = craftView(m);
ok(v.prof === 'Weaponcrafting' && v.skills.length === 2 && v.rows.length === 3, 'a crafter sees skills and recipes');
ok(v.rows[0].can && !v.rows[1].can && v.rows[2].can, 'a recipe can be made when every material is held');
ok(matLine(v.rows[1]) === 'bronze bar 5/10', 'the materials show held against needed');
ok(craftView(m, { skill: 14 }).rows.length === 1, 'the list filters by skill');
ok(v.skills[1].text === 'Woodworking 3 / 12', 'a skill shows its value and cap');
const o = craftView({ t: 'craft', offer: 1, on: 'Weaponcrafting', skills: [], recipes: [] });
ok(o.offer === 'Weaponcrafting' && o.offerSkill === 1 && o.prof === '', 'a master offers a craft to one who has none');
ok(craftView({ ...m, recipes: [{ ...m.recipes[0], ok: 0 }] }).rows[0].can === false, 'a recipe with no chance cannot be made');
ok(craftView(null) === null, 'nothing to draw');
if (failed) process.exit(1);
