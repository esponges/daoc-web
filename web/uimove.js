// Moving windows about and remembering where they were put.
//
// A stock window moves by dragging its title strip. Where it was left is kept
// in the browser (localStorage, which may be absent or refuse, so every use is
// wrapped) and used the next time it opens, moved back inside the screen if
// the browser window is smaller now. The pure parts are tested by
// uimove.test.mjs.

// How much of a window must stay on the screen, so it can always be grabbed.
export const KEEP_ON_SCREEN = 32;
// The strip at the top of a window that drags it, in pixels.
export const TITLE_STRIP = 20;

// clampToScreen moves a window of w by h at (x, y) back inside a viewport of
// vw by vh: at least KEEP_ON_SCREEN pixels of it stay in view sideways, and its
// title strip stays on the screen.
export function clampToScreen(x, y, w, h, vw, vh, keep = KEEP_ON_SCREEN) {
  const minX = keep - w, maxX = vw - keep;
  const minY = 0, maxY = Math.max(0, vh - keep);
  return {
    x: Math.round(Math.max(minX, Math.min(maxX, x))),
    y: Math.round(Math.max(minY, Math.min(maxY, y))),
  };
}

// createPositionStore remembers window positions by name. storage is
// localStorage or anything with getItem, setItem and removeItem; a storage that
// is missing or throws makes it remember nothing, quietly.
export function createPositionStore(storage, prefix = 'ui.pos.') {
  const safe = (f, d) => { try { return f(); } catch { return d; } };
  return {
    get(key) {
      const raw = safe(() => storage.getItem(prefix + key), null);
      if (!raw) return null;
      const p = safe(() => JSON.parse(raw), null);
      return p && Number.isFinite(p.x) && Number.isFinite(p.y) ? { x: p.x, y: p.y } : null;
    },
    set(key, pos) { safe(() => storage.setItem(prefix + key, JSON.stringify({ x: pos.x, y: pos.y }))); },
    remove(key) { safe(() => storage.removeItem(prefix + key)); },
    // clear forgets every position the store holds.
    clear() {
      const keys = safe(() => Array.from({ length: storage.length }, (_, i) => storage.key(i)), []);
      for (const k of keys) if (k && k.startsWith(prefix)) safe(() => storage.removeItem(k));
    },
  };
}

// createMover manages the movable windows of a page. viewport() gives
// {w, h}. makeMovable(el, key, handle) lets el (position: fixed) be dragged
// by its title strip and keeps the place; restore(key) puts a window back
// where it was left; resetAll() puts every one at its default place and
// forgets what was saved.
export function createMover({ store, viewport }) {
  const items = new Map();
  const size = (el) => ({ w: parseFloat(el.style.width) || (el.getBoundingClientRect ? el.getBoundingClientRect().width : 0), h: parseFloat(el.style.height) || (el.getBoundingClientRect ? el.getBoundingClientRect().height : 0) });
  const put = (el, x, y) => {
    el.style.left = x + 'px';
    el.style.top = y + 'px';
    el.style.marginLeft = '0px';
    el.style.right = 'auto';
  };
  // where reads a window's place from its style (a hidden window has no box
  // to measure), counting the margin the character windows are spaced with.
  const where = (el) => {
    const left = parseFloat(el.style.left), top = parseFloat(el.style.top);
    if (Number.isFinite(left) && Number.isFinite(top)) return { x: left + (parseFloat(el.style.marginLeft) || 0), y: top };
    const r = el.getBoundingClientRect ? el.getBoundingClientRect() : { left: 0, top: 0 };
    return { x: r.left, y: r.top };
  };

  return {
    // makeMovable returns the item: {restore(), reset()}.
    makeMovable(el, key, strip = TITLE_STRIP, grab = null) {
      const home = { ...where(el), set: false };
      const item = {
        el, key, home,
        restore() {
          const saved = store.get(key);
          const vp = viewport();
          if (!saved) return false;
          const s = size(el);
          const c = clampToScreen(saved.x, saved.y, s.w, s.h, vp.w, vp.h);
          put(el, c.x, c.y);
          return true;
        },
        reset() {
          store.remove(key);
          put(el, item.home.x, item.home.y);
        },
      };
      items.set(key, item);
      let drag = null;
      el.addEventListener('pointerdown', (e) => {
        const p = where(el);
        // A window with its own grab handle (the chat, whose title strip holds tabs) moves only by that.
        if (grab ? !(e.target && grab.contains && grab.contains(e.target)) : e.clientY - p.y > strip) return;
        drag = { dx: e.clientX - p.x, dy: e.clientY - p.y };
        if (el.setPointerCapture && e.pointerId !== undefined) el.setPointerCapture(e.pointerId);
      });
      el.addEventListener('pointermove', (e) => {
        if (!drag) return;
        const vp = viewport(), s = size(el);
        const c = clampToScreen(e.clientX - drag.dx, e.clientY - drag.dy, s.w, s.h, vp.w, vp.h);
        put(el, c.x, c.y);
      });
      const end = () => {
        if (!drag) return;
        drag = null;
        store.set(key, where(el));
      };
      el.addEventListener('pointerup', end);
      el.addEventListener('pointercancel', end);
      return item;
    },
    restore(key) { const it = items.get(key); return it ? it.restore() : false; },
    // resetAll puts every movable window at its default and forgets the saved places.
    resetAll() {
      store.clear();
      for (const it of items.values()) it.reset();
    },
    // reclamp moves windows that no longer fit a smaller screen back inside it.
    reclamp() {
      const vp = viewport();
      for (const it of items.values()) {
        const p = where(it.el), s = size(it.el);
        const c = clampToScreen(p.x, p.y, s.w, s.h, vp.w, vp.h);
        if (c.x !== Math.round(p.x) || c.y !== Math.round(p.y)) put(it.el, c.x, c.y);
      }
    },
    get items() { return items; },
  };
}
