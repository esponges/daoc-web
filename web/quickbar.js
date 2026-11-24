// The quickbar: rows of slots for styles and spells, drawn from the skin's own
// art.
//
// The game's client draws its quickbar itself: no XML layout describes it. Its
// frames are in the skin's frames_win texture and its place is in default.ini
// (Quickbar=329,644 on a 1024 by 768 screen, Quickbar2 beneath it, 42 pixels
// lower). The rectangles below were measured off the converted texture; they
// are one table so that correcting one is a one-line change. Where the
// converted skin is absent the callers keep their plain bars.

import { icon, text, paint } from './uiwin.js';
import { SPELL_ICON, spellKind } from './charwin.js';
import { slotState } from './spells.js';

// The picture of a book spell: by what it does, as in the spells window.
const spellIconOf = (s) => SPELL_ICON[spellKind(s.tp, s.tg)];

// SLOT_ART is one slot's frame in the frames_win texture: the dark square with
// its gold bevel. SLOT is the size it is drawn at, and BAR_W the width of a
// bar of ten, as the stock default places it.
export const SLOT_ART = { tex: 'frames_win', sx: 141, sy: 10, sw: 44, sh: 41 };
export const SLOT = 36;
export const GAP = 0;
export const BAR_W = 366;
export const ROW_H = 42;

// quickbarPlace is where a row of the bar sits on a screen of vw by vh pixels:
// centred, its first row 124 pixels from the bottom (as default.ini places it
// on 1024 by 768) and each further row 42 below the one before.
export function quickbarPlace(vw, vh, row = 0) {
  return { left: Math.round((vw - BAR_W) / 2), top: Math.round(vh - 124 + row * ROW_H) };
}

// slotFrame builds one slot's frame element, size pixels square.
export function slotFrame(doc, data, size = SLOT) {
  const e = doc.createElement('div');
  Object.assign(e.style, { position: 'absolute', left: '0px', top: '0px', width: size + 'px', height: size + 'px' });
  paint(e, data, { ...SLOT_ART, dw: size, dh: size });
  return e;
}

// createSpellQuickbar builds the row of spell slots (Shift+1 to Shift+8): the
// skin's frame, the spell's picture by kind, its key and its power cost. It is
// always shown, empty slots as bare frames.
//   onCast(id)  the slot was pressed        onAssign(id, slot)  a spell was dropped
//   onClear(slot)  the slot was right-clicked
// set(slots, book, song, held) redraws it; update(me, recasts, nowS) shades a
// slot red while the character cannot pay for it, and darkens it with the
// seconds left while its recast runs.
export function createSpellQuickbar(doc, data, { onCast, onAssign, onClear, findSpell, dragId } = {}) {
  const root = doc.createElement('div');
  Object.assign(root.style, { position: 'fixed', zIndex: 11, display: 'flex', width: BAR_W + 'px', justifyContent: 'center' });
  let cells = [];
  const label = (value, x, y, color = 'rgb(255,255,255)') => {
    const t = text(doc, data, 'arial9', value, color);
    t.style.left = x + 'px';
    t.style.top = y + 'px';
    t.style.pointerEvents = 'none';
    t.style.filter = 'drop-shadow(1px 1px 0 #000)';
    return t;
  };
  return {
    el: root,
    get cells() { return cells; },
    set(slots, book, song = 0, held = null) {
      root.textContent = '';
      cells = slots.map((id, i) => {
        const s = id && findSpell ? findSpell(book, id) : null;
        const cell = doc.createElement('div');
        cell.dataset.slot = String(i);
        cell.className = 'qslot';
        Object.assign(cell.style, { position: 'relative', width: SLOT + 'px', height: SLOT + 'px', cursor: s ? 'pointer' : 'default' });
        cell.appendChild(slotFrame(doc, data));
        if (s) {
          const pic = icon(doc, data, spellIconOf(s), 2, 2);
          pic.style.transform = 'scale(' + (SLOT - 4) / 32 + ')';
          pic.style.transformOrigin = '0 0';
          cell.appendChild(pic);
          cell.appendChild(label(String(s.pw), SLOT - 16, SLOT - 12));
          const on = (song && s.id === song) || (held && held.has(s.id));
          if (on) cell.style.outline = '2px solid #3ec7c1';
          cell.title = s.nm + ' (level ' + s.lv + ')' + (s.pl ? (song === s.id ? ' · playing: click to stop' : ' · a song: click to start, again to stop') : '') + (held && held.has(s.id) ? ' · you hold this buff' : '') + ' · right-click to remove';
          cell.addEventListener('click', () => onCast?.(id));
          cell.addEventListener('contextmenu', (e) => { if (e.preventDefault) e.preventDefault(); onClear?.(i); });
        }
        // The key is Shift and the number; the bitmap font has no arrow for Shift, so the
        // number is drawn in the spell bar's colour and the tooltip says the rest.
        cell.appendChild(label(String(i + 1), 3, 2, 'rgb(120,220,215)'));
        cell.addEventListener('dragover', (e) => e.preventDefault());
        cell.addEventListener('drop', (e) => {
          e.preventDefault();
          const sid = dragId ? dragId(e.dataTransfer?.getData('text/plain')) : null;
          if (sid) onAssign?.(sid, i);
        });
        root.appendChild(cell);
        // The recast: a dark square over the picture and the seconds left on it.
        const shade = s ? doc.createElement('div') : null;
        if (shade) {
          Object.assign(shade.style, { position: 'absolute', left: '2px', top: '2px', width: (SLOT - 4) + 'px', height: (SLOT - 4) + 'px', pointerEvents: 'none', display: 'none', background: 'rgba(0,0,0,.6)' });
          cell.appendChild(shade);
        }
        return { cell, id, spell: s, shade, wait: 0, waitLabel: null };
      });
    },
    // update shades a slot the character cannot pay for (power, or hit points for
    // a Savage spell) and darkens one whose recast is running, with its seconds.
    update(me, recasts = null, nowS = 0) {
      if (!me) return;
      cells.forEach((c) => {
        const { poor, wait } = slotState(c.spell, me, recasts, nowS);
        c.cell.style.filter = poor ? 'brightness(0.55) sepia(1) hue-rotate(-50deg) saturate(3)' : '';
        if (!c.shade || wait === c.wait) return;
        c.wait = wait;
        c.shade.style.display = wait ? 'block' : 'none';
        if (c.waitLabel) c.waitLabel.remove();
        c.waitLabel = wait ? label(String(wait), SLOT / 2 - 4, SLOT / 2 - 6, 'rgb(255,215,106)') : null;
        if (c.waitLabel) c.cell.appendChild(c.waitLabel);
      });
    },
  };
}
