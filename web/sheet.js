// What the character sheet, trainer and healer panels show, worked out from
// the server's messages. Pure functions, so node can test them; the panels in
// ui.js only lay them out.

export const STAT_NAMES = ['Strength', 'Constitution', 'Dexterity', 'Quickness', 'Intelligence', 'Piety', 'Empathy', 'Charisma'];

// CON_COLOURS are the game's con colours, from grey (-3) to purple (3).
export const CON_COLOURS = { '-3': '#9aa3ad', '-2': '#7bd47b', '-1': '#7fb4ff', 0: '#ffe066', 1: '#ffa24a', 2: '#ff5a5a', 3: '#c77dff' };

// conColour is the colour for a con number from the server, or undefined.
export function conColour(cc) {
  if (cc === undefined || cc === null) return undefined;
  return CON_COLOURS[Math.max(-3, Math.min(3, cc))];
}

// sheetView is the character sheet's content: a title, stat rows (with what
// items and levels changed from the base) and the spec and ability lists.
// USABLE are the abilities a player uses on a target, by name (the server's /guard and so on).
const USABLE = new Set(['Guard', 'Protect', 'Intercept', 'Engage']);

export function sheetView(s) {
  if (!s) return null;
  const level = s.lv + (s.half ? '.5' : '');
  return {
    title: s.cls + ', level ' + level,
    stats: STAT_NAMES.map((name, i) => ({ name, value: s.stats[i], base: s.base[i], diff: s.stats[i] - s.base[i] })),
    lines: [
      'Hit points ' + s.mhp, 'Endurance ' + s.mend, 'Weapon skill ' + s.ws, 'Armour factor ' + s.af,
      ...(s.conlost ? ['Constitution lost ' + s.conlost + ' (a healer restores it)'] : []),
    ],
    specs: (s.specs || []).map((r) => ({ name: r.nm, level: r.lv, text: r.nm + ' ' + r.lv + (r.bonus ? ' (+' + r.bonus + ')' : '') })),
    points: s.pts,
    abilities: (s.ab || []).map((a) => ({ text: a.nm + (a.lv > 1 ? ' ' + a.lv : ''), on: !!a.on, ...(USABLE.has(a.nm) && a.on ? { use: a.nm.toLowerCase() } : {}) })),
  };
}

// trainerView is a trainer's offer as rows with a train button each.
export function trainerView(t) {
  if (!t) return null;
  return {
    title: t.nm,
    points: t.pts,
    rows: (t.specs || []).map((r) => ({
      key: r.key, text: r.nm + ' ' + r.lv, cost: r.cost, can: !!r.can,
      hint: r.can ? 'Train to ' + (r.lv + 1) + ' for ' + r.cost + ' points' : 'Needs ' + r.cost + ' points, and a higher level than ' + r.lv,
    })),
  };
}

// healerText is the healer's offer, in words.
export function healerText(h, formatMoney = String) {
  if (!h) return '';
  return h.nm + ' can restore your ' + h.lost + ' lost constitution for ' + formatMoney(h.total) + ' (' + formatMoney(h.price) + ' a point).';
}

// eventLine is the log line for a rule-book event sent to you, or null.
export function eventLine(ev) {
  switch (ev.k) {
    case 'xploss': return { text: ev.x || 'You lose ' + ev.n + ' experience.', cls: 'death' };
    case 'conloss': return { text: ev.x || 'You lose ' + ev.n + ' constitution.', cls: 'death' };
    case 'train': return { text: 'You train ' + ev.x + ' to ' + ev.n + '.', cls: 'level' };
    case 'pray':
      if (ev.x === 'started') return { text: 'You kneel and pray at your grave.', cls: 'sys' };
      if (ev.x === 'cancelled') return { text: 'Your prayer is interrupted.', cls: 'sys' };
      if (ev.x === 'done') return { text: 'You regain ' + ev.n + ' experience from your grave.', cls: 'xp' };
      return null;
  }
  return null;
}
