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
};

// What starting to suffer an effect of a kind says.
const START = { stun: 'You are stunned!', mez: 'You are mesmerized!', slow: 'You are slowed!' };

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

// effectChangeLines are the lines a new effects message brings: a hold, a
// slow beginning, and being able to move again.
export function effectChangeLines(prev, next) {
  const had = new Set((prev.list || []).map((e) => e.k));
  const has = new Set((next.list || []).map((e) => e.k));
  const out = [];
  for (const k of ['stun', 'mez', 'slow']) if (has.has(k) && !had.has(k)) out.push({ text: START[k], cls: 'spell-in' });
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
      if (ev.n < 0) return toMe ? { text: (mine ? 'Your ' + ev.x + ' heals you' : cap(the(ev.from)) + ' heals you') + ' for ' + -ev.n + ' hit points.', cls: 'heal' } : null;
      if (toMe) return { text: (ev.from ? cap(the(ev.from)) + "'s " : '') + ev.x + ' hits you for ' + ev.n + ' damage!', cls: 'spell-in' };
      return { text: 'Your ' + ev.x + ' hits ' + the(ev.to) + ' for ' + ev.n + ' damage!', cls: 'spell' };
    case 'resist':
      return toMe ? { text: 'You resist the effect of ' + the(ev.from) + "'s " + ev.x + '!', cls: 'miss' }
        : { text: cap(the(ev.to)) + ' resists the effect of your ' + ev.x + '!', cls: 'miss' };
    case 'effect':
      return mine && !toMe ? { text: 'Your ' + ev.x + ' takes hold of ' + the(ev.to) + '.', cls: 'spell' } : null;
  }
  return null;
}
