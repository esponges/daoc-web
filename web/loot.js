// The loot window, built on the stock window frame. The install has no layout
// for it (the nearest are the merchant and vault windows), so it is made the
// way the inventory is: a stock frame, rows of picture, name and detail in
// stock fonts, and a stock button for Take all.

import { frame, text, icon, itemIconNumber, hrButton, layoutText } from './uiwin.js';

const WIDTH = 270;
const ROW_H = 36;
const TOP = 26;
const BOTTOM = 34;

// lootHeight is the window height for a corpse with this many rows.
export function lootHeight(rows) {
  return TOP + Math.max(1, rows) * ROW_H + BOTTOM;
}

// lootRows turns a corpse message into the rows to show: money first, then
// each item with the slot a take request names. Money is slot -1.
export function lootRows(corpse, formatMoney) {
  const rows = [];
  if (corpse && corpse.mo) rows.push({ slot: -1, name: formatMoney(corpse.mo), detail: '', money: true });
  for (const it of (corpse && corpse.items) || []) rows.push({ slot: it.slot, name: it.nm, detail: it.d || '', item: it });
  return rows;
}

// createStockLoot builds the window, hidden. show(corpse) fills and opens it.
//   onTake(corpseId, slot)  a row was clicked; slot -1 is the money
export function createStockLoot(doc, data, { onTake, formatMoney = String } = {}) {
  const root = doc.createElement('div');
  Object.assign(root.style, { position: 'relative', width: WIDTH + 'px', display: 'none' });
  let corpse = null;
  let open = false;

  function render() {
    root.textContent = '';
    const rows = lootRows(corpse, formatMoney);
    const h = lootHeight(rows.length);
    root.style.height = h + 'px';
    const bg = frame(doc, data, 'dlg_background_noresize', WIDTH, h);
    root.appendChild(bg);

    const title = text(doc, data, 'arial11', corpse ? corpse.nm : '', 'rgb(255,255,255)');
    const bf = data.bitmapFont('arial11');
    const tw = bf && corpse ? layoutText(bf, corpse.nm).width : 0;
    title.style.left = Math.max(8, (WIDTH - tw) / 2) + 'px';
    title.style.top = '7px';
    title.style.pointerEvents = 'none';
    root.appendChild(title);

    rows.forEach((r, i) => {
      const row = doc.createElement('div');
      Object.assign(row.style, {
        position: 'absolute', left: '10px', top: TOP + i * ROW_H + 'px', width: WIDTH - 20 + 'px', height: ROW_H - 2 + 'px', cursor: 'pointer',
      });
      row.addEventListener('pointerenter', () => (row.style.background = 'rgba(255,255,255,.10)'));
      row.addEventListener('pointerleave', () => (row.style.background = 'transparent'));
      row.addEventListener('click', () => corpse && onTake?.(corpse.id, r.slot));
      if (r.item) row.appendChild(icon(doc, data, itemIconNumber(r.item), 2, 1));
      const name = text(doc, data, 'arial11', r.name, r.money ? 'rgb(255,215,106)' : 'rgb(255,255,255)');
      name.style.left = (r.item ? 40 : 6) + 'px';
      name.style.top = (r.detail ? 3 : 10) + 'px';
      name.style.pointerEvents = 'none';
      row.appendChild(name);
      if (r.detail) {
        const d = text(doc, data, 'arial9', r.detail, 'rgb(170,180,190)');
        d.style.left = '40px';
        d.style.top = '19px';
        d.style.pointerEvents = 'none';
        row.appendChild(d);
      }
      root.appendChild(row);
    });

    const all = hrButton(doc, data, 'generic_hr_button', 'TAKE ALL', 110, () => {
      if (!corpse) return;
      // Money first, then items from the last slot back, so slots do not shift under the clicks.
      for (const r of lootRows(corpse, formatMoney).sort((a, b) => b.slot - a.slot)) onTake?.(corpse.id, r.slot);
    });
    Object.assign(all.el.style, { left: (WIDTH - 110) / 2 + 'px', top: h - BOTTOM + 8 + 'px' });
    root.appendChild(all.el);
  }

  return {
    el: root,
    get isOpen() {
      return open;
    },
    get corpseId() {
      return corpse ? corpse.id : null;
    },
    // show fills the window from a corpse message and opens it; a corpse with
    // nothing left closes it.
    show(c) {
      const empty = !c || ((!c.items || !c.items.length) && !c.mo);
      if (empty) {
        this.hide();
        return;
      }
      corpse = c;
      render();
      open = true;
      root.style.display = 'block';
    },
    hide() {
      open = false;
      corpse = null;
      root.style.display = 'none';
    },
  };
}
