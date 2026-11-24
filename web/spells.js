// The spellbook on the client: what the SPELLS window lists, the spell bar's
// slots and the cast bar. The pure parts are tested by spells.test.mjs; the
// DOM parts are the built-in panels used when the converted interface is
// absent, and the bar and cast bar in both cases.

import { buildWindow } from './uiwin.js';
// spellsView is the spells window's content: for each spec line a header and
// its spells, each with its level, power and cast time. A spell the server
// cannot cast yet is greyed (on: false) and cannot be dragged.
export function spellsView(book, song = 0) {
  const rows = [];
  for (const line of (book && book.lines) || []) {
    rows.push({ header: line.nm });
    for (const s of line.sp) {
      const cast = s.ct ? s.ct + 's' : 'instant';
      rows.push({
        id: s.id,
        text: s.nm + '  ' + s.lv + ', ' + s.pw + (s.hp ? ' hp, ' : ' pw, ') + cast + (s.ok ? '' : ' (not yet)') + (song && s.id === song ? '  (playing)' : ''),
        on: !!s.ok,
        playing: !!song && s.id === song,
        draggable: !!s.ok,
        drag: 'spell:' + s.id,
        spell: s,
      });
    }
  }
  return rows;
}

// findSpell returns a spell of the book by id, or null.
export function findSpell(book, id) {
  for (const line of (book && book.lines) || []) {
    for (const s of line.sp) if (s.id === id) return s;
  }
  return null;
}

// spellDragId reads the spell id out of what a spell row carries when it is
// dragged ("spell:<id>"), or null for anything else.
export function spellDragId(data) {
  const m = /^spell:(\d+)$/.exec(data || '');
  return m ? Number(m[1]) : null;
}

export const SPELL_SLOTS = 8;

// slotState is what stops a spell on the bar being cast now: poor when the
// character cannot pay for it (power, or for a Savage spell hit points, which
// may not go to 0), and wait, the whole seconds left of its recast (0 when it is
// ready). recasts maps a spell id to when its recast has passed; nowS is the
// same clock, in seconds.
export function slotState(s, me, recasts, nowS) {
  if (!s || !me) return { poor: false, wait: 0 };
  const poor = s.hp ? me.hp <= s.pw : me.power < s.pw;
  const until = recasts && recasts.get ? recasts.get(s.id) : undefined;
  const wait = until > nowS ? Math.ceil(until - nowS) : 0;
  return { poor, wait };
}

// slotsFor keeps a bar's slot assignments to spells the character still has:
// a slot holds a spell id, or 0.
export function slotsFor(saved, book) {
  const out = new Array(SPELL_SLOTS).fill(0);
  (saved || []).slice(0, SPELL_SLOTS).forEach((id, i) => { if (findSpell(book, id)) out[i] = id; });
  return out;
}

// assignSlot puts a spell into a slot (the first free one when slot is
// negative), taking it out of any other, and returns the new slots.
export function assignSlot(slots, id, slot = -1) {
  const out = slots.map((x) => (x === id ? 0 : x));
  if (slot < 0) slot = out.findIndex((x) => x === 0);
  if (slot >= 0 && slot < SPELL_SLOTS) out[slot] = id;
  return out;
}

// castBar is the state of the cast bar at a time: null when nothing is being
// cast, else the spell's name and how far along it is, 0 to 1. A cast starts
// on the player's own cast event and ends with its interrupt or failure, or
// when its time has run.
export function castBar(cast, nowS) {
  if (!cast || !(cast.dur > 0)) return null; // an instant spell has no bar
  const f = (nowS - cast.start) / cast.dur;
  if (f >= 1.05) return null;
  return { name: cast.name, f: Math.max(0, Math.min(1, f)) };
}

// DRAW_NAMES are the labels on the cast bar while a bow is drawn.
const DRAW_NAMES = { normal: 'Drawing', critical: 'Critical Shot', rapid: 'Rapid Fire' };

// castEventEffect says what a cast-related event does to the player's own cast:
// 'start' (with the spell's name and time), 'end', or null when it is not theirs.
export function castEventEffect(ev, me) {
  if (ev.from !== me) return null;
  switch (ev.k) {
    case 'cast': return { what: 'start', name: ev.x, dur: ev.dur || 0 };
    case 'draw': return { what: 'start', name: DRAW_NAMES[ev.x] || 'Drawing', dur: ev.dur || 0, draw: true };
    case 'drawstop': return { what: 'end' };
    case 'interrupt': case 'castfail': return { what: 'end', failed: true };
  }
  return null;
}

// castLine is the log line for the player's own cast beginning, failing or
// being interrupted; null for anything else.
export function ownCastLine(ev, me) {
  const e = castEventEffect(ev, me);
  if (!e || ev.k === 'draw' || ev.k === 'drawstop') return null;
  if (e.what === 'start') return { text: 'You begin casting ' + e.name + '.', cls: 'spell' };
  return ev.k === 'interrupt' ? { text: 'Your spell is interrupted.', cls: 'miss' } : { text: 'Your spell fails.', cls: 'miss' };
}

// --- the built-in panels ---

const el = (doc, tag, cls, text) => {
  const e = doc.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
};

// createSpellPanel builds the plain spells panel: the book's lines with each
// spell; a click casts it, a drag (or shift-click) puts it on the spell bar.
//   onCast(id)  onAssign(id)
export function createSpellPanel(doc, { onCast, onAssign } = {}) {
  const panel = el(doc, 'div', 'panel spells');
  panel.append(el(doc, 'h2', '', 'Spells'));
  const list = el(doc, 'div', 'spell-list');
  panel.appendChild(list);
  return {
    el: panel,
    toggle(on = !panel.classList.contains('open')) { panel.classList.toggle('open', on); },
    isOpen: () => panel.classList.contains('open'),
    render(book, song = 0) {
      list.replaceChildren();
      const rows = spellsView(book, song);
      if (!rows.length) list.appendChild(el(doc, 'div', 'empty', 'You have no spells.'));
      for (const r of rows) {
        if (r.header) { list.appendChild(el(doc, 'h3', '', r.header)); continue; }
        const d = el(doc, 'div', 'spell' + (r.on ? '' : ' off') + (r.playing ? ' playing' : ''), r.text);
        if (r.on) {
          d.draggable = true;
          d.title = 'Click to cast at your target · shift-click or drag onto the spell bar';
          d.addEventListener('dragstart', (e) => e.dataTransfer?.setData('text/plain', r.drag));
          d.addEventListener('click', (e) => (e.shiftKey ? onAssign?.(r.id) : onCast?.(r.id)));
        } else d.title = 'This kind of spell does not work yet.';
        list.appendChild(d);
      }
    },
  };
}

// createSpellBar builds the row of spell slots (Shift+1 to Shift+8). set(slots,
// book) refills it; a slot's click casts, and a spell dropped on a slot is
// assigned to it.
//   onCast(id)  onAssign(id, slot)  onClear(slot)
export function createSpellBar(doc, { onCast, onAssign, onClear } = {}) {
  const bar = el(doc, 'div', 'stylebar spellbar');
  return {
    el: bar,
    set(slots, book, song = 0, held = null) {
      bar.replaceChildren();
      const any = slots.some(Boolean);
      bar.style.display = any ? '' : 'none';
      slots.forEach((id, i) => {
        const s = id ? findSpell(book, id) : null;
        const b = el(doc, 'button', 'style' + (s ? '' : ' empty') + (s && ((song && s.id === song) || (held && held.has(s.id))) ? ' on' : ''));
        b.dataset.slot = i;
        b.append(el(doc, 'span', 'key', '⇧' + (i + 1)), el(doc, 'span', 'nm', s ? s.nm : ''), el(doc, 'span', 'cost', s ? s.pw + (s.hp ? ' hp' : ' pw') : ''));
        if (s) {
          b.title = s.nm + ' (level ' + s.lv + ')' + (s.pl ? (song === s.id ? ' · playing: click to stop' : ' · a song: click to start, again to stop') : '') + (held && held.has(s.id) ? ' · you hold this buff' : '') + ' · right-click to remove';
          b.addEventListener('click', () => onCast?.(id));
          b.addEventListener('contextmenu', (e) => { e.preventDefault(); onClear?.(i); });
        }
        b.addEventListener('dragover', (e) => e.preventDefault());
        b.addEventListener('drop', (e) => {
          e.preventDefault();
          const sid = spellDragId(e.dataTransfer?.getData('text/plain'));
          if (sid) onAssign?.(sid, i);
        });
        bar.appendChild(b);
      });
    },
  };
}

// createStockCastBar is the cast bar on the stock Timer window: the spell's name
// over a green bar that fills, shown only while the cast lasts.
function createStockCastBar(doc, data, { mover }) {
  const win = buildWindow(doc, data, 'timer', { captions: true, labels: true });
  Object.assign(win.el.style, { position: 'fixed', left: '50%', top: 'auto', bottom: '190px', marginLeft: '-150px', zIndex: 13, display: 'none', pointerEvents: 'none' });
  let shown = false;
  return {
    el: win.el,
    get shown() { return shown; },
    update(cast, nowS) {
      const s = castBar(cast, nowS);
      shown = !!s;
      win.el.style.display = s ? 'block' : 'none';
      if (!s) return;
      win.setText('timer_text', s.name);
      win.setBar('timer_time', s.f);
    },
  };
}

// createCastBar builds the bar that fills while the player casts. update(cast,
// nowS) takes the cast state castBar reads.
export function createCastBar(doc, data = null, { mover } = {}) {
  if (data && data.window && data.window('timer')) return createStockCastBar(doc, data, { mover });
  const box = el(doc, 'div', 'castbar');
  const name = el(doc, 'span', 'nm');
  const fill = el(doc, 'div', 'fill');
  box.append(fill, name);
  box.style.display = 'none';
  return {
    el: box,
    update(cast, nowS) {
      const s = castBar(cast, nowS);
      box.style.display = s ? '' : 'none';
      if (!s) return;
      name.textContent = s.name;
      fill.style.width = (100 * s.f).toFixed(1) + '%';
    },
  };
}

// castGuard stops one press from becoming two casts: a double-click on a
// spell row arrives as two clicks, and the second would be refused by the
// server with "already casting". It lets a request through unless the same
// spell was asked for a moment ago. ask(id, nowMs) says whether to send.
export function castGuard(gapMs = 300) {
  let lastId = 0, lastAt = -Infinity;
  return {
    ask(id, nowMs) {
      if (id === lastId && nowMs - lastAt < gapMs) return false;
      lastId = id;
      lastAt = nowMs;
      return true;
    },
  };
}
