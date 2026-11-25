// The quickbars: three bars of ten slots, each with ten pages, as the stock
// client has them (default.ini places Quickbar, Quickbar2 and Quickbar3, each
// GroupSize=10). Any slot holds a style, a spell or an ability; the layout is
// the player's own and is kept in the browser per character, as the stock
// client keeps it in the character's ini.
//
// Keys: the first bar is 1 to 0, the second Shift+1 to Shift+0, the third
// Alt+1 to Alt+0. 1 on the first bar's empty first slot is still the attack
// key. The arrows beside a bar (or the mouse wheel over it) turn its page.
//
// The model is pure functions, so node can test it; createQuickbars draws it,
// from the skin's art when the interface is converted and as plain buttons when not.

import { icon, text, button, styleIconNumber, MISC_ICON } from './uiwin.js';

// ABILITY_ICON pictures an ability by what it is used with, from the item sheets
// (abilities have no icon of their own in the imported data): a shield for the
// ones used on a group mate or a foe's attention, a bow for the shots. Each also
// carries a short name, as several share a picture.
const ABILITY_ICON = { guard: 1231, protect: 1231, intercept: 1231, engage: 1231, 'critical shot': 924, 'rapid fire': 924 };
export const ABILITY_SHORT = { guard: 'GRD', protect: 'PRT', intercept: 'INT', engage: 'ENG', 'critical shot': 'CRIT', 'rapid fire': 'RAPID', berserk: 'BRSK', charge: 'CHRG', quickcast: 'QC', distraction: 'DIST' };
import { SPELL_ICON, spellKind } from './charwin.js';
import { slotState, findSpell } from './spells.js';
import { slotFrame, SLOT, quickbarPlace } from './quickbar.js';
import { rightClick, DELVE_HINT, REMOVE_HINT } from './delve.js';

export const BARS = 3;
export const PAGES = 10;
export const SLOTS = 10;
// PAGER is the width of the page arrows beside a bar.
export const PAGER = 16;
export const QBAR_W = SLOTS * SLOT + PAGER;

const KINDS = new Set(['style', 'spell', 'ability']);
const same = (a, b) => !!a && !!b && a.k === b.k && a.id === b.id;
const keyOf = (e) => e.k + ':' + e.id;

// --- the layout (pure) ---

// emptyLayout is three bars of ten empty pages, each bar on its first page.
// seen lists what has ever been put on a bar for the player, so that a thing
// they took off is not put back.
export function emptyLayout() {
  return {
    v: 1,
    page: new Array(BARS).fill(0),
    bars: Array.from({ length: BARS }, () => Array.from({ length: PAGES }, () => new Array(SLOTS).fill(null))),
    seen: [],
  };
}

// cleanEntry keeps a slot's entry only when it is well formed.
function cleanEntry(e) {
  if (!e || !KINDS.has(e.k)) return null;
  if (e.k === 'ability') return typeof e.id === 'string' && e.id ? { k: e.k, id: e.id } : null;
  return Number.isInteger(e.id) && e.id > 0 ? { k: e.k, id: e.id } : null;
}

// parseLayout reads a saved layout, or gives an empty one for anything else.
export function parseLayout(raw) {
  const out = emptyLayout();
  let v = raw;
  if (typeof raw === 'string') { try { v = JSON.parse(raw); } catch { return out; } }
  if (!v || v.v !== 1 || !Array.isArray(v.bars)) return out;
  for (let b = 0; b < BARS; b++) {
    const p = v.page && Number.isInteger(v.page[b]) ? v.page[b] : 0;
    out.page[b] = p >= 0 && p < PAGES ? p : 0;
    for (let p2 = 0; p2 < PAGES; p2++) {
      for (let s = 0; s < SLOTS; s++) out.bars[b][p2][s] = cleanEntry(v.bars[b]?.[p2]?.[s]);
    }
  }
  out.seen = Array.isArray(v.seen) ? v.seen.filter((x) => typeof x === 'string') : [];
  return out;
}

const copy = (l) => ({ v: 1, page: l.page.slice(), bars: l.bars.map((b) => b.map((p) => p.slice())), seen: l.seen.slice() });

// placeEntry puts an entry into a slot (a thing may sit on several slots, as in the game).
export function placeEntry(l, bar, page, slot, entry) {
  const out = copy(l);
  out.bars[bar][page][slot] = cleanEntry(entry);
  if (out.bars[bar][page][slot] && !out.seen.includes(keyOf(entry))) out.seen.push(keyOf(entry));
  return out;
}

// moveEntry moves what is in one slot to another, swapping with what was there.
export function moveEntry(l, from, to) {
  const out = copy(l);
  const a = out.bars[from.bar][from.page][from.slot];
  out.bars[from.bar][from.page][from.slot] = out.bars[to.bar][to.page][to.slot];
  out.bars[to.bar][to.page][to.slot] = a;
  return out;
}

export function clearEntry(l, bar, page, slot) {
  const out = copy(l);
  out.bars[bar][page][slot] = null;
  return out;
}

// turnPage moves a bar d pages on, wrapping from the tenth to the first.
export function turnPage(l, bar, d) {
  const out = copy(l);
  out.page[bar] = (((out.page[bar] + d) % PAGES) + PAGES) % PAGES;
  return out;
}

// dragEntry reads what a drag carries: a style, spell or ability from its
// window ("style:<id>", "spell:<id>", "ability:<name>") or another quickbar
// slot ("qb:<bar>:<page>:<slot>").
export function dragEntry(data) {
  const s = data || '';
  let m = /^(style|spell):(\d+)$/.exec(s);
  if (m) return { entry: { k: m[1], id: Number(m[2]) } };
  m = /^ability:(.+)$/.exec(s);
  if (m) return { entry: { k: 'ability', id: m[1] } };
  m = /^qb:(\d):(\d):(\d)$/.exec(s);
  if (m) return { from: { bar: Number(m[1]), page: Number(m[2]), slot: Number(m[3]) } };
  return null;
}

// knownEntries lists what the character has that can go on a bar, by the bar
// it goes to when it is new: styles (usable or held back for the weapon) to the
// first, spells the server casts to the second, usable abilities to the third.
// ctx is { bar, also, book, abilities } as the server sent them; abilities are
// the sheet's usable names ("guard", "critical shot").
export function knownEntries(ctx) {
  const styles = [...(ctx.bar || []), ...(ctx.also || [])].map((s) => ({ k: 'style', id: s.id }));
  const spells = ((ctx.book && ctx.book.lines) || []).flatMap((l) => l.sp.filter((s) => s.ok).map((s) => ({ k: 'spell', id: s.id })));
  const abilities = (ctx.abilities || []).map((a) => ({ k: 'ability', id: a }));
  return [styles, spells, abilities];
}

// firstFree is the first empty slot on a bar, page by page; the first bar's
// first slot on its first page is left for the attack key.
function firstFree(l, bar) {
  for (let p = 0; p < PAGES; p++) {
    for (let s = 0; s < SLOTS; s++) {
      if (bar === 0 && p === 0 && s === 0) continue;
      if (!l.bars[bar][p][s]) return { page: p, slot: s };
    }
  }
  return null;
}

// fillNew puts what the character has newly gained onto its bar, in the first
// empty slot, and remembers it was placed. legacy is the old spell bar (eight
// spell ids by slot) to carry over to the second bar's first page.
export function fillNew(l, ctx, legacy = null) {
  let out = copy(l);
  let changed = false;
  if (legacy && legacy.some(Boolean)) {
    legacy.forEach((id, i) => {
      const e = { k: 'spell', id };
      if (!id || i >= SLOTS || out.seen.includes(keyOf(e)) || !findSpell(ctx.book, id)) return;
      if (!out.bars[1][0][i]) out.bars[1][0][i] = e;
      out.seen.push(keyOf(e));
      changed = true;
    });
  }
  knownEntries(ctx).forEach((list, bar) => {
    for (const e of list) {
      if (out.seen.includes(keyOf(e))) continue;
      out.seen.push(keyOf(e));
      changed = true;
      const at = firstFree(out, bar);
      if (at) out.bars[bar][at.page][at.slot] = e;
    }
  });
  return changed ? out : l;
}

// resolve says what a slot's entry is now: its name, picture, cost and whether
// it can be used, or null when the character no longer has it.
export function resolve(e, ctx) {
  if (!e) return null;
  if (e.k === 'style') {
    const i = (ctx.bar || []).findIndex((s) => s.id === e.id);
    if (i >= 0) {
      const s = ctx.bar[i];
      return { e, name: s.nm, icon: styleIconNumber(s.spec), cost: s.cost, costText: s.cost + ' end', index: i,
        title: s.nm + ' (' + s.opens + ', level ' + s.lv + ', ' + s.cost + ' endurance)' };
    }
    const h = (ctx.also || []).find((s) => s.id === e.id);
    if (h) {
      return { e, name: h.nm, icon: styleIconNumber(h.spec), cost: h.cost, costText: h.cost + ' end', locked: true,
        title: h.nm + ' (needs ' + (h.nd || h.spec || 'another weapon') + ')' };
    }
    return null;
  }
  if (e.k === 'spell') {
    const s = findSpell(ctx.book, e.id);
    if (!s) return null;
    return { e, name: s.nm, icon: SPELL_ICON[spellKind(s.tp, s.tg)], cost: s.pw, costText: s.pw + (s.hp ? ' hp' : ' pw'), spell: s,
      title: s.nm + ' (level ' + s.lv + ')' + (s.pl ? (ctx.song === s.id ? ' · playing: use again to stop' : ' · a song: use to start, again to stop') : '') + (ctx.held && ctx.held.has(s.id) ? ' · you hold this buff' : '') };
  }
  if (!(ctx.abilities || []).includes(e.id)) return null;
  const name = e.id.replace(/\b\w/g, (c) => c.toUpperCase()).replace('Quickcast', 'QuickCast');
  return { e, name, icon: ABILITY_ICON[e.id] ?? MISC_ICON, short: ABILITY_SHORT[e.id] || name.slice(0, 4).toUpperCase(), costText: '', title: name + ' (an ability)' };
}

// slotLook is how a slot is shaded this frame: ready (a style whose opening is
// met), poor (cannot be paid for), wait (seconds of a spell's recast), off (a
// style held back for the weapon), on (a song playing or a buff held).
export function slotLook(r, me, ctx, nowS = 0) {
  const out = { ready: false, poor: false, wait: 0, off: false, on: false };
  if (!r || !me) return out;
  if (r.e.k === 'style') {
    if (r.locked) { out.off = true; return out; }
    out.poor = me.end < r.cost;
    out.ready = (me.avail || []).includes(r.index);
    return out;
  }
  if (r.e.k === 'spell') {
    const st = slotState(r.spell, me, ctx.recasts, nowS);
    out.poor = st.poor;
    out.wait = st.wait;
    out.on = (!!ctx.song && ctx.song === r.spell.id) || !!(ctx.held && ctx.held.has(r.spell.id));
  }
  return out;
}

// keySlot says which bar and slot a key press is for, or null: 1 to 0 for
// the first bar, with Shift for the second, with Alt for the third.
export function keySlot(e) {
  const m = /^Digit(\d)$/.exec(e.code || '');
  if (!m || e.ctrlKey || e.metaKey) return null;
  const n = Number(m[1]);
  const slot = n === 0 ? 9 : n - 1;
  const bar = e.altKey ? 2 : e.shiftKey ? 1 : 0;
  return { bar, slot };
}

// keyName is a slot's key in words, for tooltips.
export const keyName = (bar, slot) => ['', 'Shift+', 'Alt+'][bar] + ((slot + 1) % 10);

// KEY_COLOURS tell the bars' keys apart on the slots: white, the spell bar's
// blue, gold.
export const KEY_COLOURS = ['rgb(255,255,255)', 'rgb(120,220,215)', 'rgb(255,215,106)'];

// barPlaces is where the three bars sit on a screen: the first two where
// default.ini puts Quickbar and Quickbar2, the third beside the second when it
// fits (as on the larger screens' defaults) and above the first when not.
export function barPlaces(vw, vh) {
  const a = quickbarPlace(vw, vh, 0), b = quickbarPlace(vw, vh, 1);
  const left = Math.round((vw - QBAR_W) / 2);
  a.left = left; b.left = left;
  const beside = left + QBAR_W * 2 + 8 <= vw;
  const c = beside ? { left: left + QBAR_W + 8, top: b.top } : { left, top: a.top - 42 };
  return [a, b, c];
}

// --- the bars ---

// createQuickbars builds the three bars. data is the converted interface, or
// null for plain buttons. set(layout, ctx) redraws them; update(me, ctx, nowS)
// shades the slots each frame.
//   onUse(entry, resolved)   a slot was clicked or its key pressed
//   onDelve(entry)           right-click
//   onChange(layout)         the layout changed (a drop, a clear, a page turn)
export function createQuickbars(doc, data, { onUse, onDelve, onChange } = {}) {
  const bars = [];
  let layout = emptyLayout();
  let ctx = {};
  let cells = [[], [], []];

  for (let b = 0; b < BARS; b++) {
    const el = doc.createElement('div');
    el.className = 'qbar' + (data ? ' stock' : '');
    el.dataset.bar = String(b);
    Object.assign(el.style, { position: 'fixed', zIndex: 11, display: 'flex', alignItems: 'center' });
    const row = doc.createElement('div');
    Object.assign(row.style, { display: 'flex' });
    const pager = doc.createElement('div');
    pager.className = 'pager';
    Object.assign(pager.style, { position: 'relative', width: PAGER + 'px', height: SLOT + 'px' });
    // The grip: the page number between the arrows. Dragging it moves the bar
    // (as a window's title strip does); dragging a slot moves what is in it.
    const grip = doc.createElement('div');
    grip.className = 'grip';
    grip.title = 'Drag to move this quickbar';
    Object.assign(grip.style, { position: 'absolute', left: SLOTS * SLOT + 'px', top: '13px', width: PAGER + 'px', height: (SLOT - 26) + 'px', cursor: 'move' });
    el.append(row, pager, grip);
    el.addEventListener('wheel', (e) => {
      if (e.preventDefault) e.preventDefault();
      turn(b, e.deltaY > 0 ? 1 : -1);
    });
    bars.push({ el, row, pager, grip });
  }

  const label = (value, x, y, color) => {
    if (!data) {
      const s = doc.createElement('span');
      s.textContent = value;
      Object.assign(s.style, { position: 'absolute', left: x + 'px', top: y + 'px', color, pointerEvents: 'none' });
      return s;
    }
    const t = text(doc, data, 'arial9', value, color);
    t.style.left = x + 'px';
    t.style.top = y + 'px';
    t.style.pointerEvents = 'none';
    t.style.filter = 'drop-shadow(1px 1px 0 #000)';
    return t;
  };

  function change(next) {
    layout = next;
    draw();
    onChange?.(layout);
  }

  function turn(b, d) { change(turnPage(layout, b, d)); }

  function drawPager(b) {
    const { pager } = bars[b];
    pager.textContent = '';
    const up = data ? button(doc, data, 'scroll_up', 1, 0, () => turn(b, -1)) : plainArrow('▲', 0, () => turn(b, -1));
    const down = data ? button(doc, data, 'scroll_down', 1, SLOT - 13, () => turn(b, 1)) : plainArrow('▼', SLOT - 12, () => turn(b, 1));
    up.title = 'Previous page';
    down.title = 'Next page';
    pager.append(up, down, label(String(layout.page[b] + 1), layout.page[b] === 9 ? 2 : 5, 12, 'rgb(255,215,106)'));
  }

  function plainArrow(ch, y, onClick) {
    const a = doc.createElement('div');
    a.textContent = ch;
    a.className = 'arrow';
    Object.assign(a.style, { position: 'absolute', left: '2px', top: y + 'px', fontSize: '9px', cursor: 'pointer' });
    a.addEventListener('click', onClick);
    return a;
  }

  function drawBar(b) {
    const { row } = bars[b];
    row.textContent = '';
    const page = layout.page[b];
    cells[b] = layout.bars[b][page].map((e, s) => {
      const r = resolve(e, ctx);
      const cell = doc.createElement('div');
      cell.className = 'qslot' + (r ? ' ' + r.e.k : ' empty');
      cell.dataset.slot = String(s);
      Object.assign(cell.style, { position: 'relative', width: SLOT + 'px', height: SLOT + 'px', cursor: r ? 'pointer' : 'default' });
      if (data) cell.appendChild(slotFrame(doc, data));
      if (r && data) {
        const pic = icon(doc, data, r.icon, 2, 2);
        pic.style.transform = 'scale(' + (SLOT - 4) / 32 + ')';
        pic.style.transformOrigin = '0 0';
        cell.appendChild(pic);
      } else if (r) {
        const nm = doc.createElement('span');
        nm.className = 'nm';
        nm.textContent = r.name;
        cell.appendChild(nm);
      }
      const shade = doc.createElement('div');
      Object.assign(shade.style, { position: 'absolute', left: '2px', top: '2px', width: (SLOT - 4) + 'px', height: (SLOT - 4) + 'px', pointerEvents: 'none' });
      cell.appendChild(shade);
      if (r && r.short && data) cell.appendChild(label(r.short, 3, SLOT - 12, 'rgb(255,255,255)'));
      if (r && r.cost !== undefined) cell.appendChild(label(String(r.cost), SLOT - 16, SLOT - 12, 'rgb(255,255,255)'));
      cell.appendChild(label(String((s + 1) % 10), 3, 2, KEY_COLOURS[b]));
      const key = keyName(b, s);
      if (r) {
        cell.title = r.title + ' · ' + key + ' · ' + DELVE_HINT + ' ' + REMOVE_HINT;
        cell.draggable = true;
        cell.addEventListener('dragstart', (ev) => ev.dataTransfer?.setData('text/plain', 'qb:' + b + ':' + page + ':' + s));
        cell.addEventListener('click', () => onUse?.(r.e, r));
      } else {
        cell.title = (b === 0 && s === 0 ? 'Empty: 1 is the attack key until something is put here. ' : 'Empty. ') + 'Drag a style, spell or ability here (' + key + ').';
      }
      cell.addEventListener('contextmenu', (ev) => {
        if (ev.preventDefault) ev.preventDefault();
        if (!e) return;
        if (rightClick(ev, { clearable: true }) === 'clear') change(clearEntry(layout, b, page, s));
        else onDelve?.(e);
      });
      cell.addEventListener('dragover', (ev) => ev.preventDefault?.());
      cell.addEventListener('drop', (ev) => {
        ev.preventDefault?.();
        const d = dragEntry(ev.dataTransfer?.getData('text/plain'));
        if (!d) return;
        if (d.from) change(moveEntry(layout, d.from, { bar: b, page, slot: s }));
        else change(placeEntry(layout, b, page, s, d.entry));
      });
      row.appendChild(cell);
      return { cell, shade, r, wait: 0, waitLabel: null };
    });
  }

  function draw() {
    for (let b = 0; b < BARS; b++) { drawBar(b); drawPager(b); }
  }

  return {
    els: bars.map((x) => x.el),
    grips: bars.map((x) => x.grip),
    get layout() { return layout; },
    get cells() { return cells; },
    set(next, nextCtx) {
      if (next) layout = next;
      if (nextCtx) ctx = nextCtx;
      draw();
    },
    turn,
    // use acts on a resolved slot as a click on it would (its key was pressed).
    use(r) { if (r) onUse?.(r.e, r); },
    // entryAt is what a bar's slot holds on its current page, resolved, or null.
    entryAt(b, s) { return resolve(layout.bars[b][layout.page[b]][s], ctx); },
    update(me, nowS = 0) {
      if (!me) return;
      for (const bar of cells) {
        for (const c of bar) {
          const look = slotLook(c.r, me, ctx, nowS);
          c.shade.style.background = look.off ? 'rgba(0,0,0,.65)' : look.poor ? 'rgba(160,0,0,.55)' : look.wait ? 'rgba(0,0,0,.6)' : 'transparent';
          c.cell.style.outline = look.on ? '2px solid #3ec7c1' : look.ready && !look.poor ? '1px solid #ffd76a' : 'none';
          if (look.wait === c.wait) continue;
          c.wait = look.wait;
          if (c.waitLabel) c.waitLabel.remove();
          c.waitLabel = look.wait ? label(String(look.wait), SLOT / 2 - 4, SLOT / 2 - 6, 'rgb(255,215,106)') : null;
          if (c.waitLabel) c.cell.appendChild(c.waitLabel);
        }
      }
    },
    place(vw, vh) {
      barPlaces(vw, vh).forEach((p, b) => Object.assign(bars[b].el.style, { left: p.left + 'px', top: p.top + 'px' }));
    },
  };
}
