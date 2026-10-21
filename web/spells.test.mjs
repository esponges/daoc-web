// Headless checks of the spellbook's pure parts.
//   node web/spells.test.mjs

import { spellsView, findSpell, spellDragId, slotsFor, assignSlot, castBar, castEventEffect, ownCastLine, SPELL_SLOTS } from './spells.js';

let failed = 0;
const ok = (cond, name, extra = '') => {
  if (!cond) { failed++; console.log(`FAIL ${name} ${extra}`); } else console.log(`ok   ${name}`);
};

const book = { lines: [
  { spec: 'Mending', nm: 'Mending', sp: [
    { id: 3001, nm: 'Minor Mending', lv: 1, pw: 1, ct: 2.2, ok: true },
    { id: 4001, nm: 'Summon Thing', lv: 1, pw: 5, ct: 0, ok: false },
  ] },
  { spec: 'Smite', nm: 'Smite', sp: [{ id: 5001, nm: 'Smite Bolt', lv: 3, pw: 4, ct: 3, ok: true }] },
] };

const rows = spellsView(book);
ok(rows.length === 5 && rows[0].header === 'Mending' && rows[3].header === 'Smite', 'a header for each line, then its spells');
ok(rows[1].text === 'Minor Mending  1, 1 pw, 2.2s' && rows[1].draggable && rows[1].drag === 'spell:3001', 'a spell row shows level, power and cast time');
ok(rows[2].on === false && !rows[2].draggable && rows[2].text.endsWith('(not yet)') && rows[2].text.includes('instant'), 'a spell the server cannot cast is greyed and cannot be dragged');
ok(spellsView(null).length === 0 && spellsView({ lines: [] }).length === 0, 'no book, no rows');
ok(findSpell(book, 5001).nm === 'Smite Bolt' && findSpell(book, 1) === null, 'a spell is found by id');
ok(spellDragId('spell:3001') === 3001 && spellDragId('style:12') === null && spellDragId('') === null, 'a dragged spell id is read, a style is not');

let slots = slotsFor([3001, 99999, 5001], book);
ok(slots.length === SPELL_SLOTS && slots[0] === 3001 && slots[1] === 0 && slots[2] === 5001, 'saved slots keep only the spells the character still has');
slots = assignSlot(slots, 5001, 5);
ok(slots[5] === 5001 && slots[2] === 0, 'a spell moves between slots');
slots = assignSlot(slots, 4001, -1);
ok(slots[1] === 4001, 'a spell with no slot named takes the first free one');

ok(castBar(null, 10) === null, 'no cast, no bar');
const c = { name: 'Minor Mending', start: 10, dur: 2 };
ok(castBar(c, 11).f === 0.5 && castBar(c, 10).f === 0 && castBar(c, 12).f === 1, 'the bar fills over the cast time');
ok(castBar(c, 12.2) === null, 'the bar goes when the time has run');

ok(castEventEffect({ k: 'cast', from: 7, x: 'Minor Mending', dur: 2.2 }, 7).what === 'start', 'my cast starts the bar');
ok(castEventEffect({ k: 'cast', from: 8 }, 7) === null, 'another player\'s cast is not mine');
ok(castEventEffect({ k: 'interrupt', from: 7 }, 7).what === 'end' && castEventEffect({ k: 'castfail', from: 7 }, 7).what === 'end', 'an interrupt or failure ends it');
ok(ownCastLine({ k: 'cast', from: 7, x: 'Bolt' }, 7).text === 'You begin casting Bolt.', 'the log says when I begin');
ok(ownCastLine({ k: 'interrupt', from: 7 }, 7).text === 'Your spell is interrupted.', 'and when I am interrupted');

if (failed) { console.log(`\n${failed} failed`); process.exit(1); }
