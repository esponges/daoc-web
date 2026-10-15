// The inventory, drawn over the stock paper doll (skin1/pperdoll.bmp, 192 by
// 448): worn items on the doll, the pack as five pages of eight picked by the
// sack tabs, and the money along the bottom. The install has no layout file
// for this window, so the positions below were measured from the art.

import { text, icon, itemIconNumber } from './uiwin.js';

// The server's worn slots: 40 weapon, 41 shield, 42 torso, 43 legs, 44 arms,
// 45 feet, 46 hands, 47 head.
export const WORN_SLOTS = {
  40: { x: 15, y: 235, w: 32, h: 32, name: 'Right hand' },
  41: { x: 58, y: 235, w: 32, h: 32, name: 'Left hand' },
  42: { x: 6, y: 36, w: 32, h: 32, name: 'Torso' },
  43: { x: 6, y: 195, w: 32, h: 32, name: 'Legs' },
  44: { x: 6, y: 76, w: 32, h: 32, name: 'Arms' },
  45: { x: 155, y: 195, w: 32, h: 32, name: 'Feet' },
  46: { x: 155, y: 76, w: 32, h: 32, name: 'Hands' },
  47: { x: 155, y: 36, w: 32, h: 32, name: 'Helm' },
};
// The pack is forty slots in five pages of eight, one page per sack tab.
export const PACK_SLOTS = 40;
export const PAGE_SLOTS = 8;
const TAB_Y = [295, 320, 347, 371, 396]; // sack tab centres on the art

// bagCell places pack slot i on the page it belongs to: four across, two down.
export function bagCell(i) {
  const k = i % PAGE_SLOTS;
  return { page: Math.floor(i / PAGE_SLOTS), x: 14 + (k % 4) * 38, y: 282 + Math.floor(k / 4) * 38 };
}

// coins splits an amount in copper into the five denominations of the game:
// mithril, platinum, gold, silver, copper.
export function coins(c) {
  const out = [];
  let rest = Math.max(0, Math.floor(c));
  for (const unit of [100000000, 1000000, 10000, 100, 1]) {
    out.push(Math.floor(rest / unit));
    rest %= unit;
  }
  return out;
}

// createStockPack builds the inventory window. Items are moved by dragging;
// a click uses one (equips a pack item, takes a worn one off) and a
// shift-click destroys it. The server decides every one of them.
//   onMove(from, to)  one slot dropped on another
//   onUse(slot)       click on a pack item
//   onDestroy(slot)   shift-click
//   onUnequip(slot)   click on a worn item
export function createStockPack(doc, data, { onMove, onUse, onDestroy, onUnequip } = {}) {
  const art = data.art('skin1/pperdoll.bmp');
  const root = doc.createElement('div');
  Object.assign(root.style, { position: 'relative', width: (art ? art.w : 192) + 'px', height: (art ? art.h : 448) + 'px', display: 'none' });
  if (art) {
    root.style.backgroundImage = `url("${art.url}")`;
    root.style.imageRendering = 'pixelated';
  }
  root.style.userSelect = 'none';
  root.style.touchAction = 'none';
  let page = 0;
  let items = new Map();
  let open = false;

  const cells = new Map(); // slot -> element
  const cell = (slot, x, y, w, h) => {
    const e = doc.createElement('div');
    Object.assign(e.style, { position: 'absolute', left: x + 'px', top: y + 'px', width: w + 'px', height: h + 'px' });
    e.dataset.slot = String(slot);
    root.appendChild(e);
    cells.set(slot, e);
    return e;
  };
  for (const [s, r] of Object.entries(WORN_SLOTS)) cell(Number(s), r.x, r.y, r.w, r.h).title = r.name;
  for (let i = 0; i < PACK_SLOTS; i++) {
    const b = bagCell(i);
    cell(i, b.x, b.y, 32, 32);
  }

  const tabs = TAB_Y.map((y, p) => {
    const t = doc.createElement('div');
    Object.assign(t.style, { position: 'absolute', left: '170px', top: y - 11 + 'px', width: '20px', height: '22px', cursor: 'pointer' });
    t.title = `Pack ${p + 1}`;
    t.addEventListener('click', () => {
      page = p;
      layoutPage();
    });
    root.appendChild(t);
    return t;
  });

  // The five money labels (M P G S C) along the bottom; values sit beside them.
  const moneyXs = [34, 73, 112, 151, 188];
  const moneyEls = moneyXs.map((x) => {
    const holder = doc.createElement('div');
    Object.assign(holder.style, { position: 'absolute', left: x - 30 + 'px', top: '428px', width: '28px', height: '14px', pointerEvents: 'none' });
    root.appendChild(holder);
    return holder;
  });

  function layoutPage() {
    for (let i = 0; i < PACK_SLOTS; i++) cells.get(i).style.display = Math.floor(i / PAGE_SLOTS) === page ? 'block' : 'none';
    tabs.forEach((t, p) => (t.style.outline = p === page ? '1px solid #ffd76a' : 'none'));
  }

  // --- dragging ---
  let drag = null;
  const slotAt = (x, y) => {
    const e = doc.elementFromPoint ? doc.elementFromPoint(x, y) : null;
    const c = e && e.closest ? e.closest('[data-slot]') : null;
    return c && root.contains(c) ? Number(c.dataset.slot) : null;
  };
  function press(slot, ev) {
    if (!items.has(slot)) return;
    // Without this the browser starts its own text or image drag, which
    // swallows the pointer events a drop needs.
    ev.preventDefault();
    drag = { slot, x: ev.clientX, y: ev.clientY, ghost: null };
    const move = (e) => {
      if (!drag) return;
      if (!drag.ghost && Math.hypot(e.clientX - drag.x, e.clientY - drag.y) > 4) {
        drag.ghost = icon(doc, data, itemIconNumber(items.get(slot)));
        Object.assign(drag.ghost.style, { position: 'fixed', pointerEvents: 'none', zIndex: 40, opacity: '0.85' });
        doc.body.appendChild(drag.ghost);
      }
      if (drag.ghost) {
        drag.ghost.style.left = e.clientX - 16 + 'px';
        drag.ghost.style.top = e.clientY - 16 + 'px';
      }
    };
    const up = (e) => {
      doc.removeEventListener('pointermove', move);
      doc.removeEventListener('pointerup', up);
      doc.removeEventListener('pointercancel', up);
      const d = drag;
      drag = null;
      if (!d) return;
      if (e.type === 'pointercancel') {
        if (d.ghost) d.ghost.remove();
        return;
      }
      if (d.ghost) {
        d.ghost.remove();
        const to = slotAt(e.clientX, e.clientY);
        if (to !== null && to !== d.slot) onMove?.(d.slot, to);
      } else if (e.shiftKey) onDestroy?.(d.slot);
      else if (d.slot >= PACK_SLOTS) onUnequip?.(d.slot);
      else onUse?.(d.slot);
    };
    doc.addEventListener('pointermove', move);
    doc.addEventListener('pointerup', up);
    doc.addEventListener('pointercancel', up);
  }
  for (const [slot, e] of cells) e.addEventListener('pointerdown', (ev) => press(slot, ev));

  // render takes the server's inventory: items with slot, nm, n, k, d, w, line.
  function render(inv) {
    items = new Map((inv.items || []).map((it) => [it.slot, it]));
    for (const [slot, e] of cells) {
      e.textContent = '';
      const it = items.get(slot);
      if (!it) {
        e.style.cursor = 'default';
        e.title = slot >= PACK_SLOTS ? WORN_SLOTS[slot].name : '';
        continue;
      }
      e.style.cursor = 'grab';
      e.title = it.nm + (it.d ? ' — ' + it.d : '') + (slot >= PACK_SLOTS ? '\nclick: take off' : '\nclick: equip · shift-click: destroy');
      e.appendChild(icon(doc, data, itemIconNumber(it)));
      if (it.n > 1) {
        const t = text(doc, data, 'arial9', String(it.n), 'rgb(255,255,255)');
        t.style.left = '2px';
        t.style.top = '20px';
        t.style.pointerEvents = 'none';
        e.appendChild(t);
      }
    }
    coins(inv.mo || 0).forEach((v, i) => {
      moneyEls[i].textContent = '';
      if (v > 0) {
        const t = text(doc, data, 'arial9', String(v), 'rgb(255,255,255)');
        t.style.pointerEvents = 'none';
        moneyEls[i].appendChild(t);
      }
    });
  }

  layoutPage();
  return {
    el: root,
    render,
    get isOpen() {
      return open;
    },
    toggle() {
      open = !open;
      root.style.display = open ? 'block' : 'none';
      return open;
    },
    freeSlot() {
      for (let i = 0; i < PACK_SLOTS; i++) if (!items.has(i)) return i;
      return -1;
    },
    showPage(p) {
      page = p;
      layoutPage();
    },
  };
}
