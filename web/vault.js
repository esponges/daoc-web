// The vault window: a character's 40 vault slots, opened by a vault keeper
// beside the pack. Items move by clicking (a vault item goes to the first free
// pack slot, a pack item to the first free vault slot) or by dragging a cell
// onto another. Built on the stock window frame when the interface is there,
// and as a plain panel when it is not.

import { frame, text, hrButton, layoutText } from './uiwin.js';

export const VAULT_BASE = 1000;
export const VAULT_SLOTS = 40;
export const PACK = 40;
const COLS = 8;
const CELL = 38;
const WIDTH = COLS * CELL + 20;
const HEIGHT = 26 + (VAULT_SLOTS / COLS) * CELL + 44;

// vaultCells is the 40 cells in order, each with the item in it or null.
export function vaultCells(vault) {
  const by = new Map(((vault && vault.items) || []).map((i) => [i.slot, i]));
  return Array.from({ length: VAULT_SLOTS }, (_, i) => ({ slot: VAULT_BASE + i, item: by.get(VAULT_BASE + i) || null }));
}

// firstFreeVault is the first empty vault slot, -1 when it is full.
export function firstFreeVault(vault) {
  const c = vaultCells(vault).find((x) => !x.item);
  return c ? c.slot : -1;
}

// firstFreePack is the first empty pack slot, -1 when the pack is full.
export function firstFreePack(inv) {
  const used = new Set(((inv && inv.items) || []).map((i) => i.slot));
  for (let s = 0; s < PACK; s++) if (!used.has(s)) return s;
  return -1;
}

// createVault builds the window, hidden.
//   onMove(from, to)  an item was clicked or dragged
//   onClose()         the Close button was pressed
export function createVault(doc, data, { onMove, onClose, getInv = () => null } = {}) {
  const root = doc.createElement('div');
  root.className = data ? 'vault-stock' : 'panel vault';
  if (data) Object.assign(root.style, { position: 'relative', width: WIDTH + 'px', height: HEIGHT + 'px', display: 'none' });
  let vault = null;
  let open = false;

  const cellEl = (c) => {
    const d = doc.createElement('div');
    d.className = 'cell' + (c.item ? ' full ' + (c.item.k || '') : '');
    d.dataset.slot = c.slot;
    Object.assign(d.style, { width: CELL - 4 + 'px', height: CELL - 4 + 'px', boxSizing: 'border-box', overflow: 'hidden', fontSize: '10px', lineHeight: '11px',
      border: '1px solid rgba(120,170,220,.25)', borderRadius: '3px', padding: '2px', color: '#e8eef5', cursor: c.item ? 'pointer' : 'default',
      wordBreak: 'break-word', background: c.item ? 'rgba(255,255,255,.07)' : 'transparent' });
    if (c.item) {
      d.textContent = c.item.nm + (c.item.n > 1 ? ' ×' + c.item.n : '');
      d.title = c.item.nm + (c.item.d ? ' — ' + c.item.d : '') + '\nclick: take it to your pack';
      d.draggable = true;
      d.addEventListener('dragstart', (e) => e.dataTransfer?.setData('text/plain', String(c.slot)));
      d.addEventListener('click', () => {
        const free = firstFreePack(getInv());
        if (free >= 0) onMove?.(c.slot, free);
      });
    }
    d.addEventListener('dragover', (e) => e.preventDefault());
    d.addEventListener('drop', (e) => {
      e.preventDefault();
      const from = Number(e.dataTransfer?.getData('text/plain'));
      if (Number.isFinite(from) && from !== c.slot) onMove?.(from, c.slot);
    });
    return d;
  };

  function render() {
    root.textContent = '';
    const grid = doc.createElement('div');
    Object.assign(grid.style, { display: 'grid', gridTemplateColumns: `repeat(${COLS}, ${CELL}px)`, gap: '0' });
    for (const c of vaultCells(vault)) grid.appendChild(cellEl(c));
    if (data) {
      root.appendChild(frame(doc, data, 'dlg_background_noresize', WIDTH, HEIGHT));
      const title = text(doc, data, 'arial11', 'Vault', 'rgb(255,255,255)');
      const bf = data.bitmapFont('arial11');
      const tw = bf ? layoutText(bf, 'Vault').width : 0;
      title.style.left = Math.max(8, (WIDTH - tw) / 2) + 'px';
      title.style.top = '7px';
      title.style.pointerEvents = 'none';
      root.appendChild(title);
      Object.assign(grid.style, { position: 'absolute', left: '10px', top: '26px' });
      root.appendChild(grid);
      const close = hrButton(doc, data, 'generic_hr_button', 'CLOSE', 90, () => onClose?.());
      Object.assign(close.el.style, { left: (WIDTH - 90) / 2 + 'px', top: HEIGHT - 34 + 'px' });
      root.appendChild(close.el);
    } else {
      const h = doc.createElement('h2');
      h.textContent = vault && vault.nm ? 'Vault — ' + vault.nm : 'Vault';
      root.appendChild(h);
      root.appendChild(grid);
      const b = doc.createElement('button');
      b.textContent = 'Close';
      b.addEventListener('click', () => onClose?.());
      root.appendChild(b);
    }
  }

  const setOpen = (v) => {
    open = v;
    if (data) root.style.display = v ? 'block' : 'none';
    else root.classList.toggle('open', v);
  };

  return {
    el: root,
    get isOpen() { return open; },
    get vault() { return vault; },
    show(m) { vault = m; render(); setOpen(true); },
    hide() { vault = null; setOpen(false); },
  };
}
