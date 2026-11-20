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
  mezdur: { label: 'z', colour: '#6a4a8f', title: 'mesmerizes are shorter' },
};

// What starting to suffer an effect of a kind says.
const START = { stun: 'You are stunned!', mez: 'You are mesmerized!', slow: 'You are slowed!', disease: 'You are diseased!', nearsight: 'You are nearsighted!', debuff: 'You feel weaker!' };

// effectRows turns the last effects message into rows to draw, with the time
// left counted down from when it arrived.
export function effectRows(effects, now) {
  const gone = Math.max(0, now - (effects.at || 0));
  return (effects.list || []).map((e) => {
    const k = EFFECT_KINDS[e.k] || { label: '?', colour: '#666', title: e.k };
    const left = Math.max(0, (e.left || 0) - gone);
    return { name: e.nm, kind: e.k, label: k.label, colour: k.colour, title: e.nm + ' (' + k.title + ')', left, text: clock(left) };
  });
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
