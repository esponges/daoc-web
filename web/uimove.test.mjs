// node web/uimove.test.mjs
import assert from 'node:assert/strict';
import { clampToScreen, createPositionStore, createMover, KEEP_ON_SCREEN, TITLE_STRIP } from './uimove.js';

// A window is kept where it can be grabbed.
{
  assert.deepEqual(clampToScreen(100, 80, 200, 300, 1000, 800), { x: 100, y: 80 }, 'inside: untouched');
  assert.deepEqual(clampToScreen(-500, -40, 200, 300, 1000, 800), { x: KEEP_ON_SCREEN - 200, y: 0 }, 'far left and above: a strip stays in view, the title on screen');
  assert.deepEqual(clampToScreen(5000, 5000, 200, 300, 1000, 800), { x: 1000 - KEEP_ON_SCREEN, y: 800 - KEEP_ON_SCREEN }, 'far right and below');
  assert.deepEqual(clampToScreen(900, 700, 200, 300, 640, 480), { x: 640 - KEEP_ON_SCREEN, y: 480 - KEEP_ON_SCREEN }, 'a smaller screen pulls a window back');
  assert.deepEqual(clampToScreen(10.4, 20.6, 100, 100, 500, 500), { x: 10, y: 21 }, 'whole pixels');
  console.log('ok: a window is kept inside the screen');
}

// The store remembers by name and survives a storage that fails.
{
  const mem = new Map();
  const storage = { getItem: (k) => mem.get(k) ?? null, setItem: (k, v) => mem.set(k, v), removeItem: (k) => mem.delete(k), get length() { return mem.size; }, key: (i) => [...mem.keys()][i] };
  const store = createPositionStore(storage);
  assert.equal(store.get('spells'), null);
  store.set('spells', { x: 120, y: 80 });
  store.set('pack', { x: 5, y: 6 });
  mem.set('other.thing', 'keep');
  assert.deepEqual(store.get('spells'), { x: 120, y: 80 });
  store.remove('pack');
  assert.equal(store.get('pack'), null);
  store.clear();
  assert.equal(store.get('spells'), null);
  assert.equal(mem.get('other.thing'), 'keep', 'clear forgets only window places');
  mem.set('ui.pos.bad', 'not json');
  assert.equal(store.get('bad'), null, 'a damaged entry is ignored');
  mem.set('ui.pos.nan', JSON.stringify({ x: 'a', y: 1 }));
  assert.equal(store.get('nan'), null);

  const broken = createPositionStore({ getItem() { throw new Error('blocked'); }, setItem() { throw new Error('blocked'); }, removeItem() { throw new Error('blocked'); }, get length() { throw new Error('blocked'); } });
  assert.equal(broken.get('x'), null);
  broken.set('x', { x: 1, y: 2 });
  broken.remove('x');
  broken.clear();
  const none = createPositionStore(undefined);
  assert.equal(none.get('x'), null, 'no storage at all');
  none.set('x', { x: 1, y: 2 });
  console.log('ok: positions are remembered, and a blocked or missing storage is harmless');
}

// A window drags by its title strip and the place is kept.
function fakeEl(x, y, w, h) {
  const el = { style: { width: w + 'px', height: h + 'px', left: x + 'px', top: y + 'px' }, listeners: {},
    addEventListener(t, f) { this.listeners[t] = f; },
    getBoundingClientRect() { return { left: parseFloat(this.style.left), top: parseFloat(this.style.top), width: w, height: h }; } };
  return el;
}
{
  const mem = new Map();
  const store = createPositionStore({ getItem: (k) => mem.get(k) ?? null, setItem: (k, v) => mem.set(k, v), removeItem: (k) => mem.delete(k), get length() { return mem.size; }, key: (i) => [...mem.keys()][i] });
  const vp = { w: 1000, h: 700 };
  const mover = createMover({ store, viewport: () => vp });
  const el = fakeEl(50, 60, 200, 300);
  mover.makeMovable(el, 'spells');

  // Pressing below the title strip does not drag.
  el.listeners.pointerdown({ clientX: 100, clientY: 60 + TITLE_STRIP + 10 });
  el.listeners.pointermove({ clientX: 300, clientY: 300 });
  assert.equal(el.style.left, '50px', 'the body of a window is not a handle');

  // Pressing in the strip drags, and releasing keeps the place.
  el.listeners.pointerdown({ clientX: 100, clientY: 70 });
  el.listeners.pointermove({ clientX: 400, clientY: 270 });
  assert.equal(el.style.left, '350px');
  assert.equal(el.style.top, '260px');
  el.listeners.pointerup();
  assert.deepEqual(store.get('spells'), { x: 350, y: 260 });
  el.listeners.pointermove({ clientX: 10, clientY: 10 });
  assert.equal(el.style.left, '350px', 'after the release it no longer follows');

  // Dragged off the screen it is held in view.
  el.listeners.pointerdown({ clientX: 360, clientY: 270 });
  el.listeners.pointermove({ clientX: 9000, clientY: 9000 });
  assert.equal(el.style.left, 1000 - KEEP_ON_SCREEN + 'px');
  assert.equal(el.style.top, 700 - KEEP_ON_SCREEN + 'px');
  el.listeners.pointerup();

  // The next session puts it back where it was left.
  const el2 = fakeEl(50, 60, 200, 300);
  const mover2 = createMover({ store, viewport: () => vp });
  mover2.makeMovable(el2, 'spells');
  assert.equal(mover2.restore('spells'), true);
  assert.equal(el2.style.left, 1000 - KEEP_ON_SCREEN + 'px');
  assert.equal(mover2.restore('nothing'), false);

  // On a smaller screen it opens moved back inside.
  const small = createMover({ store, viewport: () => ({ w: 400, h: 300 }) });
  const el3 = fakeEl(50, 60, 200, 300);
  small.makeMovable(el3, 'spells');
  small.restore('spells');
  assert.equal(el3.style.left, 400 - KEEP_ON_SCREEN + 'px');
  assert.equal(el3.style.top, 300 - KEEP_ON_SCREEN + 'px');

  // Shrinking the browser window pulls windows back.
  const shrink = { w: 1000, h: 700 };
  const m4 = createMover({ store: createPositionStore(undefined), viewport: () => shrink });
  const el4 = fakeEl(800, 600, 200, 300);
  m4.makeMovable(el4, 'pack');
  shrink.w = 500; shrink.h = 400;
  m4.reclamp();
  assert.equal(el4.style.left, 500 - KEEP_ON_SCREEN + 'px');
  assert.equal(el4.style.top, 400 - KEEP_ON_SCREEN + 'px');

  // /resetui: every window home again, nothing saved.
  mover2.resetAll();
  assert.equal(el2.style.left, '50px');
  assert.equal(el2.style.top, '60px');
  assert.equal(store.get('spells'), null);
  console.log('ok: a window drags by its title, is remembered, restored inside the screen, and reset');
}
