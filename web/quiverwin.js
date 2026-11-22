// The quiver window: the stock WindowQuiver, four rows for the four quiver
// slots (inventory slots 49 to 52) with what is in each and how many.
//
// The server fires from the first quiver slot that holds arrows, so that slot
// is the active one and the window marks it. Picking another row moves its
// stack into the first slot (the two swap), which makes it the one in use.

import { buildWindow, itemIconNumber } from './uiwin.js';

export const QUIVER_FIRST = 49;
export const QUIVER_SLOTS = 4;

// quiverRows are the four rows for an inventory message, each with its slot,
// the item (or null), the line to show and the stock icon number (0 for none).
// active is the index of the row the server will fire from, -1 when empty.
export function quiverRows(inv) {
  const by = new Map(((inv && inv.items) || []).map((i) => [i.slot, i]));
  const rows = Array.from({ length: QUIVER_SLOTS }, (_, i) => {
    const slot = QUIVER_FIRST + i;
    const item = by.get(slot) || null;
    return { slot, item, text: item ? item.nm + (item.n > 1 ? '  x' + item.n : '') : '', icon: item ? itemIconNumber(item) : 0 };
  });
  return { rows, active: rows.findIndex((r) => r.item) };
}

// quiverMove is the move that makes row i the active one: its stack to the
// first slot. null when it is empty or already first in use.
export function quiverMove(view, i) {
  const r = view.rows[i];
  if (!r || !r.item || i === view.active) return null;
  return { from: r.slot, to: view.rows[view.active].slot };
}

const ROW_Y = 15, ROW_H = 17;

// createStockQuiver builds the window, hidden. onMove(from, to) asks the
// server to move a stack; update(inv) redraws; toggle() opens or closes it.
export function createStockQuiver(doc, data, { onMove, mover } = {}) {
  let view = quiverRows(null), open = false;
  const win = buildWindow(doc, data, 'quiver', {
    captions: true, labels: true,
    onClose: () => api.toggle(false),
    onClick: (id, ev) => {
      const m = /^QuiverSelect(\d)$/.exec(ev || '');
      if (!m) return;
      const mv = quiverMove(view, Number(m[1]));
      if (mv) onMove?.(mv.from, mv.to);
    },
  });
  Object.assign(win.el.style, { position: 'fixed', left: '50%', top: '300px', marginLeft: '-120px', zIndex: 14, display: 'none' });
  mover?.makeMovable(win.el, 'quiver');
  const mark = win.ctl['5000'];
  const draw = () => {
    view.rows.forEach((r, i) => {
      win.setIcon('quiver_icon' + i, r.icon);
      win.setText('quiver_count_string' + i, r.text);
    });
    if (mark) {
      mark.style.display = view.active >= 0 ? 'block' : 'none';
      mark.style.top = ROW_Y + Math.max(0, view.active) * ROW_H + 'px';
    }
  };
  const api = {
    el: win.el,
    get isOpen() { return open; },
    get view() { return view; },
    update(inv) { view = quiverRows(inv); draw(); },
    toggle(on = !open) {
      open = on;
      win.el.style.display = open ? 'block' : 'none';
      if (open) mover?.restore('quiver');
    },
  };
  draw();
  return api;
}
