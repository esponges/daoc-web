// The merchant window: the stock Merchant window, filled with what the
// merchant sells.
//
// Its list box shows a page of the stock (picture and name, the price in the
// row), the arrows turn the pages and the page label says which. Clicking a
// row selects it, and the five coin fields show its price. The quantity
// arrows say how many to buy; BUY asks the server for that many, one `buy` at
// a time, stopping at the first one it refuses. Selling is a click on a pack
// item while the window is open, which ui.js handles. The server decides every
// purchase.

import { buildWindow, itemIconNumber } from './uiwin.js';
import { coins } from './pack.js';

// shopRows turns a shop message into the rows to show, in the order sent.
export function shopRows(shop, formatMoney = String) {
  return ((shop && shop.items) || []).map((it, index) => ({
    index, id: it.id, name: it.nm, detail: it.d || '', price: it.pr || 0, priceText: formatMoney(it.pr || 0), item: it,
  }));
}

// canAfford says whether money (copper) covers a row.
export const canAfford = (row, money) => money >= row.price;

// pageCount is how many pages n rows make at perPage a page; always at least one.
export const pageCount = (n, perPage) => Math.max(1, Math.ceil(n / Math.max(1, perPage)));

// pageRows is the rows of one page.
export const pageRows = (rows, page, perPage) => rows.slice(page * perPage, (page + 1) * perPage);

// MAX_QTY is the most that can be asked for at once.
export const MAX_QTY = 99;

// stepQty moves a quantity by d, kept from 1 to MAX_QTY.
export const stepQty = (q, d) => Math.max(1, Math.min(MAX_QTY, q + d));

// priceOf is what qty of a row cost, in copper; 0 when nothing is selected.
export const priceOf = (row, qty) => (row ? row.price * qty : 0);

// coinFields is a price in copper as the five coin fields show it: mithril,
// platinum, gold, silver and copper.
export const coinFields = (copper) => coins(copper);

// createBuyQueue asks for the same item n times, one at a time: send(id) goes
// out for the first at once, and for each next when ack() says the last one
// went through; fail() drops what is left (the server refused one). pending is
// how many are still to be asked for.
export function createBuyQueue(send) {
  let id = null, left = 0;
  return {
    get pending() { return left; },
    start(item, n) {
      id = item;
      left = Math.max(0, n);
      if (left > 0) { left--; send(id); }
    },
    ack() { if (left > 0) { left--; send(id); } },
    fail() { left = 0; },
  };
}

// The five coin fields' adapters, mithril first.
const COIN_ADAPTERS = ['merchant_mithril', 'merchant_platinum', 'merchant_gold', 'merchant_silver', 'merchant_copper'];

// createStockShop builds the window, hidden. show(shop) fills and opens it.
//   onBuy(itemId)   one unit is to be bought (called once per unit, in turn)
//   onDelve(itemId) a row was right-clicked
//   onClose()       the window was closed
// ack() and fail() tell it how a buy went, so a quantity stops at a refusal.
export function createStockShop(doc, data, { onBuy, onDelve, onClose, formatMoney = String, mover } = {}) {
  let shop = null, money = 0, open = false;
  let page = 0, selected = -1, qty = 1, keep = false;
  const queue = createBuyQueue((id) => onBuy?.(id));
  const win = buildWindow(doc, data, 'merchant', {
    captions: true, labels: true,
    onClose: () => onClose?.(),
    onClick: (id, ev, label) => {
      if (id === '1000') turn(-1);
      else if (id === '1001') turn(1);
      else if (id === '1007') setQty(stepQty(qty, -1));
      else if (id === '1008') setQty(stepQty(qty, 1));
      else if (label === 'BUY') buy();
    },
    // Selling is a click on a pack item, and appraising is not built.
    disabled: (ev, id, label) => label === 'SELL' || label === 'APPRAISE',
    tooltip: (ev, id, label) => (label === 'SELL' ? 'Click an item in your pack to sell it.' : 'Not built yet.'),
  });
  Object.assign(win.el.style, { position: 'fixed', left: '50%', top: '110px', marginLeft: '-175px', zIndex: 15, display: 'none' });
  mover?.makeMovable(win.el, 'merchant');
  const list = win.list('1005');
  const keepBox = win.check('1009');
  if (keepBox) keepBox.onChange((on) => { keep = on; });

  let rows = [];
  const perPage = () => (list ? list.view : 10);

  function fields() {
    const row = rows[selected];
    COIN_ADAPTERS.forEach((ad, i) => win.setText(ad, row ? String(coinFields(priceOf(row, qty))[i]) : ''));
    win.setText('merchant_quantity', String(qty));
    win.setText('merchant_page_display', 'Page ' + (page + 1) + ' of ' + pageCount(rows.length, perPage()));
  }

  function draw() {
    win.setText('merchant_title', shop ? shop.nm : '');
    if (list) {
      list.setRows(pageRows(rows, page, perPage()).map((r) => ({
        text: r.name + (r.price ? '   ' + r.priceText : ''),
        icon: itemIconNumber(r.item),
        on: canAfford(r, money),
        title: [r.detail, r.priceText, 'Right-click for more information.'].filter(Boolean).join(' - '),
      })));
      list.select(selected >= 0 && Math.floor(selected / perPage()) === page ? selected % perPage() : -1);
    }
    fields();
  }

  function turn(d) {
    const last = pageCount(rows.length, perPage()) - 1;
    const next = Math.max(0, Math.min(last, page + d));
    if (next === page) return;
    page = next;
    selected = -1;
    draw();
  }
  function setQty(q) { qty = q; fields(); }
  function buy() {
    const row = rows[selected];
    if (!row) return;
    queue.start(row.id, qty);
    if (!keep) setQty(1);
  }
  if (list) list.onSelect((i) => { selected = page * perPage() + i; fields(); });
  if (list) list.onContext((i) => { const r = rows[page * perPage() + i]; if (r) onDelve?.(r.id); });

  return {
    el: win.el,
    get isOpen() { return open; },
    get npcId() { return shop ? shop.npc : null; },
    get page() { return page; },
    get qty() { return qty; },
    get selected() { return selected; },
    show(s) {
      shop = s;
      rows = shopRows(shop, formatMoney);
      page = 0;
      selected = -1;
      qty = 1;
      open = true;
      win.el.style.display = 'block';
      mover?.restore('merchant');
      draw();
    },
    hide() {
      open = false;
      shop = null;
      queue.fail();
      win.el.style.display = 'none';
    },
    // setMoney redraws the prices that the character can and cannot afford.
    setMoney(c) {
      if (c === money) return;
      money = c;
      if (open) draw();
    },
    // ack: the last unit was bought, ask for the next. fail: it was refused.
    ack() { queue.ack(); },
    fail() { queue.fail(); },
    get pending() { return queue.pending; },
  };
}
