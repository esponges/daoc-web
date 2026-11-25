// node web/charwin.test.mjs — what the attributes, skills and styles windows show.
import { attributesView, skillsView, stylesView, STAT_ADAPTERS, spellKind, spellGroups, styleGroups, SPELL_ICON } from './charwin.js';
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
ok(Object.keys(MENU_WHY).length === 3 && !MENU_WHY.ToggleJournal && MENU_WHY.ShowWarMap && Object.values(MENU_WHY).every((t) => t.length > 10), 'each greyed menu button has a tooltip saying why');
// The spells tree: a group for each spec line, rows drawn by kind.
ok(spellKind('DirectDamage') === 'damage' && spellKind('Bolt') === 'damage' && spellKind('HealthRegenBuff') === 'heal', 'spells are sorted into kinds by their type');
ok(spellKind('StrengthConstitutionBuff', 'Pet') === 'pet' && spellKind('Summon', 'Self') === 'pet', 'a spell aimed at the pet, or a summon, is a pet spell');
ok(spellKind('Mesmerize') === 'mez' && spellKind('Stun') === 'stun' && spellKind('Disease') === 'debuff' && spellKind('Odd') === 'other', 'holds, debuffs and the unknown have their kinds');
const book = { lines: [
  { nm: 'Bone Army', sp: [{ id: 1, nm: 'Summon Bone Boss', lv: 1, pw: 27, ct: 6, ok: true, tp: 'Summon' }, { id: 2, nm: 'Soften Bones', lv: 1, pw: 2, ct: 3, ok: true, tp: 'Debuff' }] },
  { nm: 'Darkness', sp: [{ id: 3, nm: 'Minor Dusk Dart', lv: 1, pw: 2, ct: 2.6, ok: true, tp: 'DirectDamage' }, { id: 4, nm: 'Odd Spell', lv: 5, pw: 3, ct: 0, ok: false, tp: 'Odd' }] },
] };
const sg = spellGroups(book, 3, { Darkness: false });
ok(sg.length === 2 && sg[0].title === 'Bone Army' && sg[0].open === true && sg[1].open === false, 'a group for each line, open unless the player folded it');
ok(sg[0].rows[0].text === 'Summon Bone Boss  1, 27 pw, 6s' && sg[0].rows[0].icon === SPELL_ICON.pet && sg[0].rows[0].drag === 'spell:1', 'a spell row has its cost and time, its kind icon and a drag payload');
ok(sg[1].rows[0].text.endsWith('(playing)') && sg[1].rows[1].text.includes('instant') && sg[1].rows[1].text.includes('(not yet)'), 'the song that is on is marked, an instant spell says so, one the server cannot cast says not yet');
ok(sg[1].rows[1].on === false && sg[1].rows[1].drag === undefined, 'a spell that cannot be cast is greyed and cannot be dragged');
ok(spellGroups(null).length === 0 && spellGroups({ lines: [] }).length === 0, 'no book, no groups');

// The styles tree.
const stg = styleGroups(slots, [{ id: 9, nm: 'Slice', spec: 'Sword' }], (spec) => (spec === 'Hammer' ? 634 : 705), { Shield: false });
ok(stg.map((g) => g.title).join('|') === 'Hammer|Shield|Not usable now', 'styles are grouped by line, then the ones that cannot be used');
ok(stg[0].rows.length === 2 && stg[0].rows[0].drag === 'style:1' && stg[0].rows[0].icon === 634 && stg[1].open === false, 'a usable style carries its id and the line icon; a folded line stays folded');
ok(stg[2].rows[0].on === false && stg[2].rows[0].drag === 'style:9' && stg[2].rows[0].text.includes('Sword'), 'a style that cannot be used now is greyed, and can still go on a quickbar');
ok(styleGroups([], []).length === 0, 'no styles, no groups');
process.exit(failed ? 1 : 0);
