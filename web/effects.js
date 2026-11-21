// Lasting spells on the player, and the lines spells put in the log.
//
// The server sends the player's own effects (`effects`: name, kind, seconds
// left) whenever they change, and spell events as they happen. Effects are
// drawn by kind, a letter on a coloured badge: the imported icon numbers do
// not match this install's sheets (see CLAUDE.md), so they are not looked up.

export const EFFECT_KINDS = {
  dot: { label: 'D', colour: '#b0412e', title: 'damage over time' },
  slow: { label: 'S', colour: '#3a6fb0', title: 'slowed' },
  stun: { label: '!', colour: '#c9a227', title: 'stunned' },
  mez: { label: 'Z', colour: '#8a4fbf', title: 'mesmerized' },
  proc: { label: '+', colour: '#3f8f4a', title: 'weapon charged' },
  buff: { label: 'B', colour: '#2f8f8f', title: 'strengthened' },
  dmgadd: { label: 'A', colour: '#c07a2a', title: 'extra damage' },
  speed: { label: '>', colour: '#2a9d8f', title: 'quickened' },
  regen: { label: 'H', colour: '#4a9a5c', title: 'recovering faster' },
  resist: { label: 'R', colour: '#6a7fb5', title: 'resistant' },
  sick: { label: 'X', colour: '#7a6a3a', title: 'weakened by resurrection' },
  haste: { label: '»', colour: '#c07a2a', title: 'swinging faster' },
  shield: { label: '^', colour: '#a0522d', title: 'damage shield' },
  endregen: { label: 'E', colour: '#4a8f3a', title: 'recovering endurance faster' },
  powregen: { label: 'P', colour: '#4a6fbf', title: 'recovering power faster' },
  debuff: { label: 'v', colour: '#a04040', title: 'weakened' },
  resdebuff: { label: 'r', colour: '#a05a40', title: 'resists lowered' },
  combatslow: { label: '<', colour: '#8a4a4a', title: 'swinging slower' },
  disease: { label: 'd', colour: '#6a7a2a', title: 'diseased' },
  nearsight: { label: 'n', colour: '#5a5a8a', title: 'nearsighted' },
  fatigue: { label: 'f', colour: '#8a6a3a', title: 'tiring faster' },
  bladeturn: { label: '|', colour: '#9aa0a8', title: 'the next blow is turned' },
  ablative: { label: 'a', colour: '#b0a070', title: 'absorbing damage' },
  absorb: { label: 'o', colour: '#7a8a5a', title: 'absorbing more' },
  camouflage: { label: 'c', colour: '#4a6a4a', title: 'camouflaged' },
  stealth: { label: 'h', colour: '#5a6a7a', title: 'hidden' },
  mezdur: { label: 'z', colour: '#6a4a8f', title: 'mesmerizes are shorter' },
};

// The kinds that harm whoever holds them: their badge is framed in red, the
// others in gold.
export const HARMFUL_KINDS = new Set(['dot', 'slow', 'stun', 'mez', 'debuff', 'resdebuff', 'combatslow', 'disease', 'nearsight', 'fatigue', 'sick']);

// What starting to suffer an effect of a kind says.
const START = { stun: 'You are stunned!', mez: 'You are mesmerized!', slow: 'You are slowed!', disease: 'You are diseased!', nearsight: 'You are nearsighted!', debuff: 'You feel weaker!' };

// effectRows turns the last effects message into rows to draw, with the time
// left counted down from when it arrived.
export function effectRows(effects, now) {
  const gone = Math.max(0, now - (effects.at || 0));
  const rows = (effects.list || []).map((e) => {
    const k = EFFECT_KINDS[e.k] || { label: '?', colour: '#666', title: e.k };
    const left = Math.max(0, (e.left || 0) - gone);
    return { name: e.nm, kind: e.k, label: k.label, colour: k.colour, title: e.nm + ' (' + k.title + ')', left, text: e.of === 0 ? '' : clock(left), of: e.of || 0, harmful: HARMFUL_KINDS.has(e.k) };
  });
  // What helps first, what harms after it; each in the order the server holds them.
  return rows.filter((r) => !r.harmful).concat(rows.filter((r) => r.harmful));
}

// clock shows seconds as "12s" or "1:05".
export function clock(s) {
  s = Math.ceil(s);
  return s >= 60 ? Math.floor(s / 60) + ':' + String(s % 60).padStart(2, '0') : s + 's';
}

// releaseLine is what the Release window says of the time left before the
// server releases a dead character: seconds under two minutes, minutes above.
export function releaseLine(left) {
  if (left <= 0) return 'Releasing...';
  if (left > 120) return `You will be released in ${Math.ceil(left / 60)} minutes`;
  return `You will be released in ${Math.ceil(left)} seconds`;
}

// effectChangeLines are the lines a new effects message brings: a hold, a
// slow beginning, and being able to move again.
export function effectChangeLines(prev, next) {
  const had = new Set((prev.list || []).map((e) => e.k));
  const has = new Set((next.list || []).map((e) => e.k));
  const out = [];
  for (const k of ['stun', 'mez', 'slow', 'disease', 'nearsight', 'debuff']) if (has.has(k) && !had.has(k)) out.push({ text: START[k], cls: 'spell-in' });
  const heldBefore = had.has('stun') || had.has('mez'), heldNow = has.has('stun') || has.has('mez');
  if (heldBefore && !heldNow) out.push({ text: 'You can move again.', cls: 'sys' });
  else if (had.has('slow') && !has.has('slow') && !heldNow) out.push({ text: 'You can move at full speed again.', cls: 'sys' });
  return out;
}

// spellLine is the log line for a spell, resist or effect event, given who
// the player is and a name for any other entity ("the wolf"); null when the
// event is not the player's business.
export function spellLine(ev, me, the) {
  const toMe = ev.to === me, mine = ev.from === me;
  if (!toMe && !mine) return null;
  const cap = (t) => t.charAt(0).toUpperCase() + t.slice(1);
  switch (ev.k) {
    case 'spell':
      if (ev.n < 0 && !toMe) return { text: 'Your ' + ev.x + ' heals ' + the(ev.to) + ' for ' + -ev.n + ' hit points.', cls: 'heal' };
      if (ev.n < 0) return toMe ? { text: (mine ? 'Your ' + ev.x + ' heals you' : cap(the(ev.from)) + ' heals you') + ' for ' + -ev.n + ' hit points.', cls: 'heal' } : null;
      if (toMe) return { text: (ev.from ? cap(the(ev.from)) + "'s " : '') + ev.x + ' hits you for ' + (ev.n || 0) + ' damage!', cls: 'spell-in' };
      return { text: 'Your ' + ev.x + ' hits ' + the(ev.to) + ' for ' + (ev.n || 0) + ' damage!', cls: 'spell' };
    case 'resist':
      return toMe ? { text: 'You resist the effect of ' + the(ev.from) + "'s " + ev.x + '!', cls: 'miss' }
        : { text: cap(the(ev.to)) + ' resists the effect of your ' + ev.x + '!', cls: 'miss' };
    case 'effect':
      return mine && !toMe ? { text: 'Your ' + ev.x + ' takes hold of ' + the(ev.to) + '.', cls: 'spell' } : null;
  }
  return null;
}

// petNames maps the ids of the player's pet, and of the minions it commands,
// to their names, from the last `pet` message; empty when there is none.
export function petNames(pet) {
  const m = new Map();
  if (!pet || pet.cl) return m;
  if (pet.id) m.set(pet.id, pet.nm || 'pet');
  for (const x of pet.mn || []) m.set(x.id, x.nm || 'minion');
  return m;
}

// petLine is the log line for a blow or spell that the player's pet (or one of
// its minions) gave or took, in the game's words; null for anything else,
// including events that concern the player, which have lines of their own.
//   names  petNames(pet)      the(id)  a name for any other entity ("the wolf")
const VERB = { block: 'blocks', parry: 'parries', evade: 'evades' };

export function petLine(ev, names, the, me) {
  if (!names.size || ev.from === me || ev.to === me) return null;
  const cap = (t) => t.charAt(0).toUpperCase() + t.slice(1);
  const mine = names.get(ev.from), hit = names.get(ev.to);
  if (!mine && !hit) return null;
  const n = ev.n || 0;
  switch (ev.k) {
    case 'hit':
      if (mine) return { text: 'Your ' + mine + ' hits ' + the(ev.to) + ' for ' + n + ' damage!' + (ev.c ? ' Critical!' : ''), cls: 'hit' };
      return { text: cap(the(ev.from)) + ' hits your ' + hit + ' for ' + n + ' damage!', cls: 'hit-in' };
    case 'miss':
      if (mine) return { text: 'Your ' + mine + ' misses ' + the(ev.to) + '!', cls: 'miss' };
      return { text: cap(the(ev.from)) + ' misses your ' + hit + '!', cls: 'miss-in' };
    case 'block': case 'parry': case 'evade':
      // The event comes from the one who defended.
      if (mine) return { text: 'Your ' + mine + ' ' + VERB[ev.k] + ' ' + the(ev.to) + "'s attack!", cls: ev.k };
      return { text: cap(the(ev.from)) + ' ' + VERB[ev.k] + ' your ' + hit + "'s attack.", cls: 'miss' };
    case 'spell':
      if (n < 0) return null; // a heal has its own line
      if (mine) return { text: 'Your ' + mine + "'s " + ev.x + ' hits ' + the(ev.to) + ' for ' + n + ' damage!', cls: 'spell' };
      return { text: (ev.from ? cap(the(ev.from)) + "'s " : '') + ev.x + ' hits your ' + hit + ' for ' + n + ' damage!', cls: 'spell-in' };
  }
  return null;
}

// castLine is the log line for a creature beginning or losing a cast near
// the player; as on 1.65 there is no cast bar for another's cast.
export function castLine(ev, the) {
  const cap = (t) => t.charAt(0).toUpperCase() + t.slice(1);
  switch (ev.k) {
    case 'cast': return { text: cap(the(ev.from)) + ' begins casting a spell!', cls: 'spell-in' };
    case 'interrupt': return { text: cap(the(ev.from)) + "'s spell is interrupted!", cls: 'spell' };
  }
  return null;
}
