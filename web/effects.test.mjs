// node web/effects.test.mjs
import assert from 'node:assert/strict';
import { effectRows, effectChangeLines, spellLine, castLine, clock, releaseLine, EFFECT_KINDS } from './effects.js';

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

// Songs and buffs have badges of their own, not the "?" of an unknown kind.
for (const k of ['buff', 'dmgadd', 'speed', 'regen', 'resist', 'stealth']) {
  const r = effectRows({ at: 0, list: [{ nm: 'Song', k, left: 4 }] }, 0)[0];
  assert.notEqual(r.label, '?', k + ' has a badge');
  assert.match(r.title, /^Song \(/);
}
console.log('ok: the song and buff kinds draw');

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

// Resurrection sickness is drawn, and the release timer speaks in minutes when it is long.
assert.equal(EFFECT_KINDS.sick.label, 'X');
assert.match(effectRows({ at: 0, list: [{ nm: 'Resurrection Sickness', k: 'sick', left: 60, of: 60 }] }, 0)[0].title, /weakened by resurrection/);
assert.equal(releaseLine(900), 'You will be released in 15 minutes');
assert.equal(releaseLine(121), 'You will be released in 3 minutes');
assert.equal(releaseLine(90), 'You will be released in 90 seconds');
assert.equal(releaseLine(0), 'Releasing...');
console.log('ok: resurrection sickness and the release timer');

// The pet's fight, in the log of its owner.
{
  const { petLine, petNames } = await import('./effects.js');
  const names = petNames({ id: 50, nm: 'returned commander', mn: [{ id: 51, nm: 'skeletal minion' }] });
  const the = (id) => (id === 9 ? 'the svartalf guard' : id === 8 ? 'the wolf' : 'the creature');
  assert.equal(petLine({ k: 'hit', from: 50, to: 9, n: 23 }, names, the, 7).text, 'Your returned commander hits the svartalf guard for 23 damage!');
  assert.equal(petLine({ k: 'hit', from: 50, to: 9, n: 40, c: 1 }, names, the, 7).text, 'Your returned commander hits the svartalf guard for 40 damage! Critical!');
  assert.equal(petLine({ k: 'hit', from: 9, to: 50, n: 12 }, names, the, 7).text, 'The svartalf guard hits your returned commander for 12 damage!');
  assert.equal(petLine({ k: 'miss', from: 51, to: 9 }, names, the, 7).text, 'Your skeletal minion misses the svartalf guard!', 'a minion is logged too');
  assert.equal(petLine({ k: 'miss', from: 9, to: 50 }, names, the, 7).text, 'The svartalf guard misses your returned commander!');
  assert.equal(petLine({ k: 'parry', from: 50, to: 9 }, names, the, 7).text, "Your returned commander parries the svartalf guard's attack!");
  assert.equal(petLine({ k: 'spell', from: 50, to: 9, x: 'Bone Bolt', n: 14 }, names, the, 7).text, "Your returned commander's Bone Bolt hits the svartalf guard for 14 damage!");
  assert.equal(petLine({ k: 'spell', from: 9, to: 50, x: 'Dart', n: 5 }, names, the, 7).text, "The svartalf guard's Dart hits your returned commander for 5 damage!");
  assert.equal(petLine({ k: 'hit', from: 8, to: 9, n: 5 }, names, the, 7), null, "someone else's fight is not mine");
  assert.equal(petLine({ k: 'hit', from: 9, to: 7, n: 5 }, names, the, 7), null, 'a blow on me has its own line');
  assert.equal(petLine({ k: 'hit', from: 50, to: 9, n: 5 }, petNames(null), the, 7), null, 'no pet, no lines');
  assert.equal(petNames({ cl: true, id: 50 }).size, 0, 'a released pet is nobody');
  console.log('ok: the pet\'s fight is logged in the game\'s words');
}
