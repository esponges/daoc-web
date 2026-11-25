// node web/quickbar.test.mjs
import assert from 'node:assert/strict';
import { UIData } from './uiwin.js';
import { quickbarPlace, slotFrame, SLOT, BAR_W, SLOT_ART } from './quickbar.js';

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
