// node web/charwin.test.mjs — what the attributes, skills and styles windows show.
import { attributesView, skillsView, stylesView, STAT_ADAPTERS } from './charwin.js';
import { styleDragId } from './hud.js';
import { MENU_WHY } from './chrome.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

const st = {
  cls: 'Warrior', race: 'Troll', lv: 44, half: true, mhp: 1004, mend: 100, ws: 312, af: 140, conlost: 2, pts: 12,
  stats: [152, 99, 60, 35, 60, 60, 60, 60], base: [110, 80, 45, 35, 60, 60, 60, 60],
  specs: [{ nm: 'Sword', lv: 30, bonus: 3 }, { nm: 'Shields', lv: 5 }],
  ab: [{ nm: 'Guard', lv: 1, on: true }, { nm: 'Protect', lv: 3, on: true }, { nm: 'Sprint', lv: 1, on: false }],
};
const a = attributesView({ name: 'Browsertroll' }, st);
ok(a.name === 'Browsertroll' && a.level === '44.5' && a.weaponSkill === '312' && a.hitPoints === '1004' && a.armourFactor === '140', 'the attributes window shows name, level, weapon skill, hit points and armour factor');
ok(a.stats.join(',') === '152,99,60,35,60,60,60,60' && STAT_ADAPTERS.length === 8, 'the eight stats are in order');
ok(a.info[0] === 'Troll Warrior' && a.info.includes('Endurance 100') && a.info.includes('Constitution lost 2'), 'race, class and endurance are in the detail');
const none = attributesView(null, null);
ok(none.name === '' && none.level === '' && none.stats.every((s) => s === '') && none.info.length === 0, 'with nothing sent every field is empty');
ok(attributesView({}, { ...st, race: undefined }).info[0] === 'Warrior', 'a missing race is left out, not invented');

const s = skillsView(st);
ok(s.points === 'pts: 12' && s.specs.length === 2 && s.specs[0].text === 'Sword  30 (+3)', 'skills list the specs with levels and the points to spend');
ok(s.abilities.map((x) => x.text).join('|') === 'Guard|Protect 3|Sprint (not yet)', 'skills list abilities, marking those the server does not act on');
ok(skillsView(null).specs.length === 0 && skillsView(null).points === '', 'no sheet, no skills');

const slots = [{ id: 1, nm: 'Slam', spec: 'Hammer', opens: 'anytime', cost: 10 }, { id: 2, nm: 'Bash', spec: 'Shield', opens: 'after block', cost: 8 }, { id: 3, nm: 'Smash', spec: 'Hammer', opens: 'after parry', cost: 12 }];
const v = stylesView(slots, [{ id: 9, nm: 'Slice', spec: 'Sword' }]);
ok(v.usable.map((r) => r.header || r.text).join('|') === 'Hammer|Slam  anytime, 10 end|Smash  after parry, 12 end|Shield|Bash  after block, 8 end', 'styles are grouped by line with opening and cost');
ok(v.usable.filter((r) => r.draggable).every((r) => Number.isInteger(r.id)), 'each usable style row carries its id to drag');
ok(v.locked.length === 1 && v.locked[0].draggable === false && v.locked[0].text.includes('Sword'), 'a style the character cannot use now is listed but cannot be dragged');
ok(styleDragId('style:42') === 42 && styleDragId('42') === null && styleDragId(undefined) === null, 'only a style drag is read as one');
ok(Object.keys(MENU_WHY).length === 5 && MENU_WHY.ShowWarMap && Object.values(MENU_WHY).every((t) => t.length > 10), 'each greyed menu button has a tooltip saying why');
process.exit(failed ? 1 : 0);
