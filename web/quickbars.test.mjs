// node web/quickbars.test.mjs
import assert from 'node:assert/strict';
import {
  BARS, PAGES, SLOTS, emptyLayout, parseLayout, placeEntry, moveEntry, clearEntry, turnPage, dragEntry,
  knownEntries, fillNew, resolve, slotLook, keySlot, keyName, barPlaces, createQuickbars, QBAR_W,
} from './quickbars.js';

const bar = [
  { id: 11, nm: 'Plan', spec: 'Axe', cost: 6, opens: 'any', lv: 2 },
  { id: 12, nm: 'Hack', spec: 'Axe', cost: 8, opens: 'after Plan', lv: 4 },
];
const also = [{ id: 13, nm: 'Shield Bash', spec: 'Shields', cost: 5, nd: 'a shield', lv: 1 }];
const book = { lines: [{ nm: 'Darkness', sp: [
  { id: 3, nm: 'Minor Dusk Dart', lv: 1, pw: 2, ct: 2.6, ok: true, tp: 'DirectDamage' },
  { id: 4, nm: 'Summon Thing', lv: 1, pw: 5, ct: 0, ok: false, tp: 'Summon' },
  { id: 7, nm: 'Song', lv: 2, pw: 4, ct: 0, ok: true, tp: 'Buff', pl: true },
] }] };
const ctx = { bar, also, book, abilities: ['guard', 'critical shot'] };

// Three bars, ten pages of ten.
{
  const l = emptyLayout();
  assert.equal(l.bars.length, BARS);
  assert.ok(l.bars.every((b) => b.length === PAGES && b.every((p) => p.length === SLOTS && p.every((s) => s === null))));
  assert.deepEqual(l.page, [0, 0, 0]);
  console.log('ok: three bars of ten pages of ten slots');
}

// A saved layout is read back; anything malformed is dropped.
{
  let l = placeEntry(emptyLayout(), 1, 3, 4, { k: 'spell', id: 3 });
  l = turnPage(l, 1, 3);
  const back = parseLayout(JSON.stringify(l));
  assert.deepEqual(back.bars[1][3][4], { k: 'spell', id: 3 });
  assert.equal(back.page[1], 3);
  assert.ok(back.seen.includes('spell:3'));
  assert.deepEqual(parseLayout('not json'), emptyLayout());
  assert.deepEqual(parseLayout({ v: 2 }), emptyLayout());
  const bad = parseLayout({ v: 1, page: [12, -1, 'x'], bars: [[[{ k: 'macro', id: 1 }, { k: 'style', id: -3 }, { k: 'ability', id: '' }]]] });
  assert.deepEqual(bad.page, [0, 0, 0]);
  assert.deepEqual(bad.bars[0][0].slice(0, 3), [null, null, null]);
  console.log('ok: layouts are saved and read back, bad entries dropped');
}

// Placing, moving, clearing, turning pages; the original is never changed.
{
  const l0 = emptyLayout();
  const l1 = placeEntry(l0, 0, 0, 2, { k: 'style', id: 11 });
  assert.equal(l0.bars[0][0][2], null, 'placing copies');
  const l2 = placeEntry(l1, 2, 5, 0, { k: 'style', id: 11 });
  assert.deepEqual(l2.bars[0][0][2], l2.bars[2][5][0], 'one thing may sit on several slots, as in the game');
  const l3 = moveEntry(placeEntry(l2, 0, 0, 3, { k: 'spell', id: 3 }), { bar: 0, page: 0, slot: 2 }, { bar: 0, page: 0, slot: 3 });
  assert.deepEqual([l3.bars[0][0][2], l3.bars[0][0][3]], [{ k: 'spell', id: 3 }, { k: 'style', id: 11 }], 'a move swaps');
  assert.equal(clearEntry(l3, 0, 0, 3).bars[0][0][3], null);
  assert.equal(turnPage(l0, 0, -1).page[0], 9, 'back from the first page is the tenth');
  assert.equal(turnPage(turnPage(l0, 0, -1), 0, 1).page[0], 0);
  console.log('ok: place, move, clear and turn pages');
}

// What a drag carries.
{
  assert.deepEqual(dragEntry('style:12'), { entry: { k: 'style', id: 12 } });
  assert.deepEqual(dragEntry('spell:3'), { entry: { k: 'spell', id: 3 } });
  assert.deepEqual(dragEntry('ability:critical shot'), { entry: { k: 'ability', id: 'critical shot' } });
  assert.deepEqual(dragEntry('qb:1:4:9'), { from: { bar: 1, page: 4, slot: 9 } });
  assert.equal(dragEntry('17'), null);
  assert.equal(dragEntry(undefined), null);
  console.log('ok: drags from the windows and from other slots are read');
}

// New things go to their bar: styles to the first (its first slot left for
// attack), spells the server casts to the second, abilities to the third; a
// thing taken off is not put back.
{
  const [st, sp, ab] = knownEntries(ctx);
  assert.deepEqual(st.map((e) => e.id), [11, 12, 13], 'held-back styles go on the bar too');
  assert.deepEqual(sp.map((e) => e.id), [3, 7], 'a spell the server cannot cast is left off');
  assert.deepEqual(ab.map((e) => e.id), ['guard', 'critical shot']);
  let l = fillNew(emptyLayout(), ctx);
  assert.equal(l.bars[0][0][0], null, 'the first slot is left for the attack key');
  assert.deepEqual(l.bars[0][0].slice(1, 4).map((e) => e.id), [11, 12, 13]);
  assert.deepEqual(l.bars[1][0].slice(0, 2).map((e) => e.id), [3, 7]);
  assert.deepEqual(l.bars[2][0].slice(0, 2).map((e) => e.id), ['guard', 'critical shot']);
  assert.equal(fillNew(l, ctx), l, 'nothing new, the same layout');
  l = clearEntry(l, 0, 0, 1);
  assert.equal(fillNew(l, ctx).bars[0][0][1], null, 'a style taken off stays off');
  const more = { ...ctx, bar: [...bar, { id: 14, nm: 'Chop', spec: 'Axe', cost: 9, opens: 'any', lv: 6 }] };
  assert.equal(fillNew(l, more).bars[0][0][1].id, 14, 'a newly learned style takes the first empty slot');
  console.log('ok: new styles, spells and abilities are placed once');
}

// A full first page spills onto the next.
{
  const many = { bar: Array.from({ length: 12 }, (_, i) => ({ id: 100 + i, nm: 'S' + i, spec: 'Axe', cost: 1, opens: 'any', lv: 1 })) };
  const l = fillNew(emptyLayout(), many);
  assert.equal(l.bars[0][0].filter(Boolean).length, 9);
  assert.deepEqual(l.bars[0][1].slice(0, 3).map((e) => e.id), [109, 110, 111]);
  console.log('ok: a full page spills onto the next');
}

// The old eight-slot spell bar carries over to the second bar's first page.
{
  const l = fillNew(emptyLayout(), ctx, [0, 7, 0, 0, 0, 0, 0, 3]);
  assert.equal(l.bars[1][0][1].id, 7);
  assert.equal(l.bars[1][0][7].id, 3);
  assert.equal(l.bars[1][0].filter(Boolean).length, 2, 'the carried spells are not placed twice');
  console.log('ok: the old spell bar carries over');
}

// What a slot shows and how it is shaded.
{
  const plan = resolve({ k: 'style', id: 11 }, ctx);
  assert.equal(plan.name, 'Plan');
  assert.equal(plan.index, 0);
  const bash = resolve({ k: 'style', id: 13 }, ctx);
  assert.ok(bash.locked && bash.title.includes('needs a shield'));
  assert.equal(resolve({ k: 'style', id: 99 }, ctx), null, 'a style no longer known shows as empty');
  assert.equal(resolve({ k: 'ability', id: 'engage' }, ctx), null);
  assert.equal(resolve({ k: 'ability', id: 'critical shot' }, ctx).name, 'Critical Shot');
  const me = { end: 7, avail: [0], power: 1, hp: 10 };
  assert.deepEqual(slotLook(plan, me, ctx), { ready: true, poor: false, wait: 0, off: false, on: false });
  assert.equal(slotLook(resolve({ k: 'style', id: 12 }, ctx), me, ctx).poor, true, 'not enough endurance');
  assert.equal(slotLook(bash, me, ctx).off, true);
  const dart = resolve({ k: 'spell', id: 3 }, ctx);
  assert.equal(slotLook(dart, me, ctx).poor, true, 'not enough power');
  const recasts = new Map([[3, 15]]);
  assert.equal(slotLook(dart, { ...me, power: 9 }, { ...ctx, recasts }, 12.5).wait, 3, 'the recast seconds left');
  const song = resolve({ k: 'spell', id: 7 }, { ...ctx, song: 7 });
  assert.equal(slotLook(song, me, { ...ctx, song: 7 }).on, true, 'the song playing is ringed');
  console.log('ok: slots resolve and shade');
}

// Keys: 1 to 0, Shift for the second bar, Alt for the third; Ctrl is the browser's.
{
  assert.deepEqual(keySlot({ code: 'Digit1' }), { bar: 0, slot: 0 });
  assert.deepEqual(keySlot({ code: 'Digit0' }), { bar: 0, slot: 9 });
  assert.deepEqual(keySlot({ code: 'Digit3', shiftKey: true }), { bar: 1, slot: 2 });
  assert.deepEqual(keySlot({ code: 'Digit5', altKey: true }), { bar: 2, slot: 4 });
  assert.equal(keySlot({ code: 'Digit5', ctrlKey: true }), null);
  assert.equal(keySlot({ code: 'KeyA' }), null);
  assert.equal(keyName(1, 9), 'Shift+0');
  assert.equal(keyName(2, 0), 'Alt+1');
  console.log('ok: the keys of the three bars');
}

// Where the bars sit: the first two as default.ini puts them, the third beside
// the second when it fits, above the first when not.
{
  const [a, b, c] = barPlaces(1600, 900);
  assert.equal(a.left, b.left);
  assert.equal(b.top - a.top, 42);
  assert.equal(c.left, b.left + QBAR_W + 8);
  assert.equal(c.top, b.top);
  const [a2, , c2] = barPlaces(700, 600);
  assert.equal(c2.left, a2.left);
  assert.equal(c2.top, a2.top - 42);
  console.log('ok: the bars are placed');
}

// The plain bars (no converted interface): use, delve, clear, drop, turn a page.
{
  const fakeDoc = {
    createElement(tag) {
      const e = {
        tag, style: {}, dataset: {}, children: [], listeners: {}, className: '', title: '',
        appendChild(c) { this.children.push(c); return c; },
        append(...cs) { cs.forEach((c) => this.children.push(c)); },
        addEventListener(t, f) { this.listeners[t] = f; },
        remove() {},
      };
      let text = '';
      Object.defineProperty(e, 'textContent', { get: () => text, set: (v) => { text = v; if (v === '') e.children = []; } });
      return e;
    },
  };
  const log = [];
  let saved = null;
  const q = createQuickbars(fakeDoc, null, { onUse: (e) => log.push(['use', e.k, e.id]), onDelve: (e) => log.push(['delve', e.k, e.id]), onChange: (l) => { saved = l; } });
  q.set(fillNew(emptyLayout(), ctx), ctx);
  assert.equal(q.els.length, 3);
  const row0 = q.els[0].children[0];
  assert.equal(row0.children.length, 10);
  row0.children[1].listeners.click();
  row0.children[1].listeners.contextmenu({ preventDefault() {} });
  q.use(q.entryAt(2, 0));
  assert.deepEqual(log, [['use', 'style', 11], ['delve', 'style', 11], ['use', 'ability', 'guard']]);
  assert.ok(row0.children[1].title.includes('Plan') && row0.children[1].title.includes('Shift+Right-click to remove.'));
  row0.children[1].listeners.contextmenu({ preventDefault() {}, shiftKey: true });
  assert.equal(saved.bars[0][0][1], null, 'Shift+right-click clears the slot');
  q.els[0].children[0].children[5].listeners.drop({ preventDefault() {}, dataTransfer: { getData: () => 'spell:7' } });
  assert.deepEqual(saved.bars[0][0][5], { k: 'spell', id: 7 }, 'a spell dropped on the first bar stays there');
  q.els[1].listeners.wheel({ preventDefault() {}, deltaY: 1 });
  assert.equal(saved.page[1], 1, 'the wheel turns the page');
  assert.equal(q.entryAt(1, 0), null, 'the second page is empty');
  q.update({ end: 0, avail: [], power: 0, hp: 1 });
  console.log('ok: the plain bars use, delve, clear, take drops and turn pages');
}

// The bars move like the windows: by the grip on the page number, not by a slot.
{
  const { createMover } = await import('./uimove.js');
  const mk = (tag) => {
    const e = { tag, style: {}, dataset: {}, children: [], listeners: {}, className: '', title: '', parentElement: null,
      appendChild(c) { c.parentElement = this; this.children.push(c); return c; },
      append(...cs) { cs.forEach((c) => this.appendChild(c)); },
      addEventListener(t, f) { (this.listeners[t] ||= []).push(f); },
      contains(n) { for (let x = n; x; x = x.parentElement) if (x === this) return true; return false; },
      fire(t, ev) { (this.listeners[t] || []).forEach((f) => f(ev)); },
      getBoundingClientRect() { return { left: 0, top: 0, width: 376, height: 36 }; },
      remove() {} };
    let text = '';
    Object.defineProperty(e, 'textContent', { get: () => text, set: (v) => { text = v; if (v === '') e.children = []; } });
    return e;
  };
  const saved = new Map();
  const store = { get: (k) => saved.get(k), set: (k, v) => saved.set(k, v), remove: (k) => saved.delete(k), clear: () => saved.clear() };
  const mover = createMover({ store, viewport: () => ({ w: 1600, h: 900 }) });
  const q = createQuickbars({ createElement: mk }, null, {});
  q.set(emptyLayout(), {});
  q.place(1600, 900);
  assert.equal(q.grips.length, 3);
  const el = q.els[0];
  const item = mover.makeMovable(el, 'quickbar1', 0, q.grips[0]);
  const x0 = parseFloat(el.style.left), y0 = parseFloat(el.style.top);
  const slot = el.children[0].children[3];
  el.fire('pointerdown', { target: slot, clientX: x0 + 100, clientY: y0 + 10 });
  el.fire('pointermove', { clientX: x0 - 200, clientY: y0 - 300 });
  el.fire('pointerup', {});
  assert.equal(parseFloat(el.style.left), x0, 'pressing a slot does not move the bar');
  el.fire('pointerdown', { target: q.grips[0], clientX: x0 + 365, clientY: y0 + 18 });
  el.fire('pointermove', { clientX: x0 + 365 - 200, clientY: y0 + 18 - 300 });
  el.fire('pointerup', {});
  assert.deepEqual([parseFloat(el.style.left), parseFloat(el.style.top)], [x0 - 200, y0 - 300], 'the grip moves it');
  assert.deepEqual(saved.get('quickbar1'), { x: x0 - 200, y: y0 - 300 }, 'and the place is kept');
  item.reset();
  assert.deepEqual([parseFloat(el.style.left), parseFloat(el.style.top)], [x0, y0], '/resetui puts it back');
  console.log('ok: a quickbar moves by its grip and keeps its place');
}
