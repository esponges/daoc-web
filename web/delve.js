// Delve: right-click on an item, spell or style asks the server what it is, and
// the answer is written to the chat log, as the stock client does. The server's
// `delveinfo` message carries the name, a level, whether the name is con
// coloured (gear is) and the lines to show.

// DELVE_HINT is what a slot's tooltip says about right-clicking; REMOVE_HINT the
// same for the slots that can be cleared (a quickbar's).
export const DELVE_HINT = 'Right-click for more information.';
export const REMOVE_HINT = 'Shift+Right-click to remove.';

// delveEntries turns a `delveinfo` message into the log's lines: the name (in
// its con colour for gear, against the player's own level) and then the details.
// colorFor(level, myLevel) is the game's con colour, passed in so this stays pure.
export function delveEntries(m, myLevel, colorFor) {
  const out = [];
  const head = m.nm + (m.lv > 0 ? ' (level ' + m.lv + ')' : '');
  out.push({ text: head, kind: 'info', color: m.con && m.lv > 0 && colorFor ? colorFor(m.lv, myLevel) : undefined });
  for (const l of m.ln || []) out.push({ text: '  ' + l, kind: 'info' });
  return out;
}

// rightClick says what a right-click on a slot should do: clear it with Shift
// held and the slot can be cleared, delve otherwise.
export function rightClick(ev, { clearable = false } = {}) {
  return ev.shiftKey && clearable ? 'clear' : 'delve';
}
