// node web/quickbar.test.mjs
import assert from 'node:assert/strict';
import { UIData } from './uiwin.js';
import { quickbarPlace, slotFrame, createSpellQuickbar, SLOT, BAR_W, SLOT_ART } from './quickbar.js';

function fakeDoc() {
  return {
    createElement(tag) {
      const e = {
        tag, style: {}, dataset: {}, children: [], listeners: {}, className: '',
        appendChild(c) { this.children.push(c); return c; },
        append(...cs) { cs.forEach((c) => this.children.push(c)); },
        addEventListener(t, f) { this.listeners[t] = f; },
      };
      let text = '';
      Object.defineProperty(e, 'textContent', { get: () => text, set: (v) => { text = v; if (v === '') e.children = []; } });
      return e;
    },
  };
}
const data = new UIData({ skin: 't', sizes: { 'atlantis/frames_win.tga': [512, 256], 'icons/spl_0.png': [320, 320] }, icons: [{ first: 0, kind: 'spl', file: 'icons/spl_0.png' }], bitmap: {},
  elements: [{ t: 'Texture', c: [{ t: 'Name', x: 'frames_win' }, { t: 'File', x: 'atlantis/frames_win.tga' }] }] }, 'data/ui/test');
const doc = fakeDoc();

// Where the stock bar sits: as default.ini places it on its own screen, and centred on others.
{
  assert.deepEqual(quickbarPlace(1024, 768, 0), { left: 329, top: 644 }, 'Quickbar=329,644 on 1024 by 768');
  assert.deepEqual(quickbarPlace(1024, 768, 1), { left: 329, top: 686 }, 'Quickbar2 is 42 below (328,686 in the file, one pixel off)');
  assert.deepEqual(quickbarPlace(1280, 1024, 0), { left: 457, top: 900 }, 'centred, the same distance from the bottom');
  assert.equal(BAR_W, 366);
  console.log('ok: the quickbar sits where the game puts it');
}

// A slot frame is the skin's art at the slot size.
{
  const f = slotFrame(doc, data);
  assert.equal(f.style.width, SLOT + 'px');
  assert.ok(f.style.backgroundImage.includes('frames_win.png'), 'it is cut from frames_win');
  const r3 = (n) => Math.round(n * 1000) / 1000;
  assert.equal(f.style.backgroundSize, `${r3(512 * SLOT / SLOT_ART.sw)}px ${r3(256 * SLOT / SLOT_ART.sh)}px`);
  console.log('ok: a slot frame is drawn from the skin');
}

// The spell bar: always eight slots, spells drawn by kind with their cost, drop to assign, right-click to clear.
{
  const book = { lines: [{ nm: 'Darkness', sp: [{ id: 3, nm: 'Minor Dusk Dart', lv: 1, pw: 2, ct: 2.6, ok: true, tp: 'DirectDamage' }, { id: 7, nm: 'Song', lv: 2, pw: 4, ct: 0, ok: true, tp: 'Buff', pl: true }] }] };
  const findSpell = (b, id) => b.lines.flatMap((l) => l.sp).find((s) => s.id === id) || null;
  const log = [];
  const bar = createSpellQuickbar(doc, data, {
    onCast: (id) => log.push(['cast', id]), onAssign: (id, slot) => log.push(['assign', id, slot]), onClear: (slot) => log.push(['clear', slot]),
    findSpell, dragId: (d) => { const m = /^spell:(\d+)$/.exec(d || ''); return m ? Number(m[1]) : null; },
  });
  bar.set([3, 0, 0, 0, 0, 0, 0, 7], book, 7, null);
  assert.equal(bar.cells.length, 8, 'eight slots, shown even when empty');
  assert.equal(bar.el.children[1].children.length, 2, 'an empty slot: its frame and its key');
  const full = bar.el.children[0];
  assert.ok(full.children.length >= 4, 'a full slot: frame, picture, cost, key');
  assert.equal(full.title.includes('Minor Dusk Dart'), true);
  full.listeners.click();
  full.listeners.contextmenu({ preventDefault() {} });
  assert.deepEqual(log, [['cast', 3], ['clear', 0]]);
  assert.equal(bar.el.children[7].style.outline, '2px solid #3ec7c1', 'the song that is on is ringed');
  assert.ok(bar.el.children[7].title.includes('playing'));
  const empty = bar.el.children[2];
  empty.listeners.drop({ preventDefault() {}, dataTransfer: { getData: () => 'spell:3' } });
  empty.listeners.drop({ preventDefault() {}, dataTransfer: { getData: () => 'style:5' } });
  assert.deepEqual(log.slice(2), [['assign', 3, 2]], 'a spell dropped on a slot is assigned; a style is not');
  empty.listeners.click;
  // Not enough power: shaded.
  bar.update({ power: 1 }, [3], book);
  assert.notEqual(bar.cells[0].cell.style.filter, '', 'a spell that costs more than the power left is shaded');
  bar.update({ power: 50 }, [3], book);
  assert.equal(bar.cells[0].cell.style.filter, '');
  console.log('ok: the spell bar shows its slots, assigns, clears, casts and shades');
}
