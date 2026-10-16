// The merchant window, built on the stock window frame the way the loot window
// is: the install has art for a merchant window but no layout for it, so this
// is a stock frame with rows of picture, name, detail and price in stock fonts
// and a stock button. A merchant can sell over a hundred things, so the rows
// scroll inside a window of fixed height.

import { frame, text, icon, itemIconNumber, hrButton, layoutText } from './uiwin.js';

export const WIDTH = 300;
export const ROW_H = 36;
const TOP = 26;
const BOTTOM = 52;
export const MAX_ROWS = 8;

// shopHeight is the window height for a shop with this many rows.
export function shopHeight(rows) {
  return TOP + Math.min(MAX_ROWS, Math.max(1, rows)) * ROW_H + BOTTOM;
}

// shopRows turns a shop message into the rows to show, in the order sent.
export function shopRows(shop, formatMoney = String) {
  return ((shop && shop.items) || []).map((it, index) => ({
    index, id: it.id, name: it.nm, detail: it.d || '', price: it.pr || 0, priceText: formatMoney(it.pr || 0), item: it,
  }));
}

// canAfford says whether money (copper) covers a row.
export const canAfford = (row, money) => money >= row.price;

// createStockShop builds the window, hidden. show(shop) fills and opens it.
//   onBuy(itemId)   a row was clicked
//   onClose()       the Close button was pressed
export function createStockShop(doc, data, { onBuy, onClose, formatMoney = String } = {}) {
  const root = doc.createElement('div');
  Object.assign(root.style, { position: 'relative', width: WIDTH + 'px', display: 'none' });
  let shop = null;
  let money = 0;
  let open = false;

  function render() {
    root.textContent = '';
    const rows = shopRows(shop, formatMoney);
    const h = shopHeight(rows.length);
    root.style.height = h + 'px';
    root.appendChild(frame(doc, data, 'dlg_background_noresize', WIDTH, h));

    const title = text(doc, data, 'arial11', shop ? shop.nm : '', 'rgb(255,255,255)');
    const bf = data.bitmapFont('arial11');
    const tw = bf && shop ? layoutText(bf, shop.nm).width : 0;
    title.style.left = Math.max(8, (WIDTH - tw) / 2) + 'px';
    title.style.top = '7px';
    title.style.pointerEvents = 'none';
    root.appendChild(title);

    const visible = Math.min(MAX_ROWS, Math.max(1, rows.length));
    const list = doc.createElement('div');
    Object.assign(list.style, {
      position: 'absolute', left: '10px', top: TOP + 'px', width: WIDTH - 20 + 'px', height: visible * ROW_H + 'px',
      overflowY: 'auto', overflowX: 'hidden',
    });
    rows.forEach((r, i) => {
      const row = doc.createElement('div');
      Object.assign(row.style, { position: 'absolute', left: '0', top: i * ROW_H + 'px', width: WIDTH - 36 + 'px', height: ROW_H - 2 + 'px', cursor: 'pointer' });
      row.addEventListener('pointerenter', () => (row.style.background = 'rgba(255,255,255,.10)'));
      row.addEventListener('pointerleave', () => (row.style.background = 'transparent'));
      row.addEventListener('click', () => onBuy?.(r.id));
      row.appendChild(icon(doc, data, itemIconNumber(r.item), 2, 1));
      const name = text(doc, data, 'arial11', r.name, 'rgb(255,255,255)');
      name.style.left = '40px';
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
      const price = text(doc, data, 'arial9', r.priceText, canAfford(r, money) ? 'rgb(255,215,106)' : 'rgb(255,110,110)');
      const pw = bf ? layoutText(data.bitmapFont('arial9') || bf, r.priceText).width : 0;
      price.style.left = Math.max(40, WIDTH - 36 - pw - 4) + 'px';
      price.style.top = '3px';
      price.style.pointerEvents = 'none';
      row.appendChild(price);
      list.appendChild(row);
    });
    if (rows.length) {
      const spacer = doc.createElement('div');
      Object.assign(spacer.style, { position: 'absolute', top: rows.length * ROW_H + 'px', height: '1px', width: '1px' });
      list.appendChild(spacer);
    }
    root.appendChild(list);

    const cash = text(doc, data, 'arial9', formatMoney(money), 'rgb(255,215,106)');
    cash.style.left = '12px';
    cash.style.top = h - BOTTOM + 8 + 'px';
    cash.style.pointerEvents = 'none';
    root.appendChild(cash);
    const close = hrButton(doc, data, 'generic_hr_button', 'CLOSE', 90, () => onClose?.());
    Object.assign(close.el.style, { left: (WIDTH - 90) / 2 + 'px', top: h - BOTTOM + 24 + 'px' });
    root.appendChild(close.el);
  }

  return {
    el: root,
    get isOpen() {
      return open;
    },
    get npcId() {
      return shop ? shop.npc : null;
    },
    show(s) {
      shop = s;
      render();
      open = true;
      root.style.display = 'block';
    },
    hide() {
      open = false;
      shop = null;
      root.style.display = 'none';
    },
    // setMoney redraws the prices that the character can and cannot afford.
    setMoney(c) {
      if (c === money) return;
      money = c;
      if (open) render();
    },
  };
}
