// The Interact window: what an NPC says, in the stock Interact window (a title
// and a scrolling text area), with a button or two for what can be done about
// it. A healer's offer is the first use.

import { buildWindow, button } from './uiwin.js';
import { healerText } from './sheet.js';

// healerTalk is the Interact content for a healer's offer: the title, the text
// and the one thing that can be accepted.
export function healerTalk(h, formatMoney = String) {
  return { title: h.nm, text: healerText(h, formatMoney), accept: 'ACCEPT' };
}

// Where the buttons go: below the text area, which is grown to leave room.
const WIN = { w: 384, h: 350 };
const BTN_Y = 322;

// createStockTalk builds the window, hidden. show({title, text, accept})
// opens it; onAccept() and onClose() are what the player chose.
export function createStockTalk(doc, data, { onAccept, onClose, mover } = {}) {
  const win = buildWindow(doc, data, 'interact', { onClose: () => api.hide() });
  win.resize(WIN.w, WIN.h);
  Object.assign(win.el.style, { position: 'fixed', left: '50%', top: '120px', marginLeft: '-192px', zIndex: 15, display: 'none' });
  mover?.makeMovable(win.el, 'interact');
  let open = false;
  const api = {
    el: win.el,
    get isOpen() { return open; },
    show(v) {
      open = true;
      win.setText('interact_title', v.title || '');
      const area = win.area('1000');
      if (area) area.setText(v.text || '');
      if (accept) accept.style.display = v.accept ? '' : 'none'; // a greeting has nothing to accept
      win.el.style.display = 'block';
      mover?.restore('interact');
      return v;
    },
    hide() {
      if (!open) return;
      open = false;
      win.el.style.display = 'none';
      onClose?.();
    },
  };
  // The window has no buttons of its own: ACCEPT does the offer, CLOSE walks away.
  const put = (label, x, fn) => {
    const b = button(doc, data, 'button_large', x, BTN_Y, fn, label);
    win.el.appendChild(b);
    return b;
  };
  const accept = put('ACCEPT', 120, () => { onAccept?.(); api.hide(); });
  put('CLOSE', 200, () => api.hide());
  return api;
}
