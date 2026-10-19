// node web/effects.test.mjs
import assert from 'node:assert/strict';
import { effectRows, effectChangeLines, spellLine, castLine, clock, EFFECT_KINDS } from './effects.js';

// Rows: drawn by kind, counting down from when the message arrived.
const fx = { at: 10, list: [{ nm: 'Shroud of Doubt', k: 'slow', ic: 10082, left: 5, of: 5 }, { nm: 'Mind Agony', k: 'dot', ic: 4427, left: 70, of: 70 }] };
const rows = effectRows(fx, 12);
assert.equal(rows.length, 2);
assert.equal(rows[0].label, EFFECT_KINDS.slow.label);
assert.equal(rows[0].text, '3s');
assert.equal(rows[1].text, '1:08');
assert.match(rows[0].title, /Shroud of Doubt \(slowed\)/);
assert.equal(effectRows(fx, 100)[0].left, 0, 'never below zero');
assert.equal(effectRows({ list: [{ nm: 'X', k: 'odd', left: 1 }], at: 0 }, 0)[0].label, '?', 'an unknown kind still draws');
assert.equal(clock(59.2), '1:00');
console.log('ok: effect rows count down and draw by kind');

// Change lines.
const stun = { list: [{ nm: 'Stunning Yell', k: 'stun', left: 3 }] };
assert.deepEqual(effectChangeLines({ list: [] }, stun).map((l) => l.text), ['You are stunned!']);
assert.deepEqual(effectChangeLines(stun, { list: [] }).map((l) => l.text), ['You can move again.']);
assert.deepEqual(effectChangeLines({ list: [] }, fx).map((l) => l.text), ['You are slowed!']);
assert.deepEqual(effectChangeLines(fx, { list: [] }).map((l) => l.text), ['You can move at full speed again.']);
assert.deepEqual(effectChangeLines(fx, fx), [], 'nothing new, nothing said');
console.log('ok: holds and slows are announced, and their end');

// Spell lines.
const the = (id) => (id === 5 ? 'the shaman' : 'the wolf');
assert.equal(spellLine({ k: 'spell', from: 5, to: 7, x: 'Spirit Bolt', n: 24 }, 7, the).text, "The shaman's Spirit Bolt hits you for 24 damage!");
assert.equal(spellLine({ k: 'spell', from: 7, to: 6, x: 'Blightstrike', n: 11 }, 7, the).text, 'Your Blightstrike hits the wolf for 11 damage!');
assert.equal(spellLine({ k: 'spell', from: 0, to: 7, x: 'Mind Agony', n: 10 }, 7, the).text, 'Mind Agony hits you for 10 damage!', 'a tick whose caster has gone');
assert.equal(spellLine({ k: 'resist', from: 5, to: 7, x: 'Shroud of Doubt' }, 7, the).text, "You resist the effect of the shaman's Shroud of Doubt!");
assert.equal(spellLine({ k: 'spell', from: 5, to: 5, x: 'Mend', n: -36 }, 7, the), null, 'a creature healing itself is not the player\'s business');
assert.equal(spellLine({ k: 'spell', from: 5, to: 6, x: 'Spirit Bolt', n: 24 }, 7, the), null);
console.log('ok: spell, resist and effect lines');

// Cast lines.
assert.equal(castLine({ k: 'cast', from: 5, to: 7, x: 'Spirit Bolt', dur: 3 }, the).text, 'The shaman begins casting a spell!');
assert.equal(castLine({ k: 'interrupt', from: 5, x: 'Spirit Bolt' }, the).text, "The shaman's spell is interrupted!");
assert.equal(castLine({ k: 'castfail', from: 5 }, the), null);
console.log('ok: cast lines, with no bar');
