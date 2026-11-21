// Headless checks of the stock controls drawn by uictl.js.
//   node web/uictl.test.mjs
// What they look like is checked in a browser (/uidemo.html).

import assert from 'node:assert/strict';
import { UIData, buildWindow, child, num, str, pt } from './uiwin.js';
import { vrCells } from './uictl.js';

const n = (t, body) => {
  if (Array.isArray(body)) return { t, c: body };
  if (body && typeof body === 'object') return { t, c: Object.entries(body).map(([k, v]) => n(k, v)) };
  return { t, x: String(body) };
};

// A stand-in document, as in uiwin.test.mjs.
export function fakeDoc() {
  return {
    createElement(tag) {
      const e = {
        tag, style: {}, dataset: {}, children: [], listeners: {},
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

const layout = {
  skin: 'test',
  sizes: { 'atlantis/a.tga': [256, 256] },
  bitmap: {},
  elements: [
    n('Texture', { Name: 'page1', File: 'atlantis/a.tga' }),
    n('HorizontalResizeImageTemplate', [n('Name', 'inset'), n('Height', 15), n('LeftWidth', 10), n('RepeatWidth', 20), n('RightWidth', 10),
      n('Texture', [n('TextureName', 'page1'), n('Left', { X: 117, Y: 240 }), n('Repeat', { X: 128, Y: 240 }), n('Right', { X: 149, Y: 240 })])]),
    n('VerticalResizeImageTemplate', [n('Name', 'gutter'), n('Width', 16), n('TopHeight', 1), n('RepeatHeight', 18), n('BottomHeight', 1),
      n('Texture', [n('TextureName', 'page1'), n('Top', { X: 186, Y: 95 }), n('Repeat', { X: 186, Y: 75 }), n('Bottom', { X: 186, Y: 95 })])]),
    n('ImageAreaTemplate', [n('Name', 'emote'), n('Size', { X: 20, Y: 20 }), n('TopLeft', { X: 80, Y: 100 }), n('TextureName', 'page1')]),
    n('HorizontalResizeButtonTemplate', [n('Name', 'hrb'), n('Height', 20), n('Normal', 'inset'), n('Pressed', 'inset'), n('NormalHighlit', 'inset'), n('PressedHighlit', 'inset'),
      n('Font', [n('Name', 'arial9')])]),
    n('IconTemplate', [n('Name', 'plain_icon_large'), n('Size', { X: 32, Y: 32 })]),
    n('StatusIconTemplate', [n('Name', 'pip'), n('Width', 8), n('Height', 8), n('TextureName', 'page1'), n('TextureStart', { X: 190, Y: 97 })]),
    n('WindowTemplate', [
      n('Name', 'imgs'), n('WindowId', 'Imgs'), n('Width', 200), n('Height', 120),
      n('HorizontalResizeImageDef', [n('TemplateName', 'inset'), n('Width', 100), n('Position', { X: 5, Y: 6 }), n('ControlId', 'h')]),
      n('VerticalResizeImageDef', [n('TemplateName', 'gutter'), n('Height', 60), n('Position', { X: 7, Y: 8 }), n('ControlId', 'v')]),
      n('HorizontalResizeImageButtonDef', [n('ImageAreaTemplateName', 'emote'), n('HRButtonTemplateName', 'hrb'), n('ControlId', '1001'), n('Width', 30), n('Position', { X: 1, Y: 2 })]),
      n('IconDef', [n('TemplateName', 'plain_icon_large'), n('AdapterName', 'pic'), n('ControlId', 'ic'), n('Position', { X: 40, Y: 40 })]),
      n('StatusIconDef', [n('TemplateName', 'pip'), n('AdapterName', 'pw'), n('ControlId', 'st'), n('Position', { X: 60, Y: 60 })]),
      n('DynamicImageDef', [n('ControlId', 'dyn'), n('TextureName', 'page1'), n('TextureCoords', { X: 4, Y: 8 }), n('Dimensions', { X: 50, Y: 40 }), n('Position', { X: 70, Y: 70 })]),
    ]),
  ],
};
const data = new UIData(layout, 'data/ui/test');
const doc = fakeDoc();
const h = { child, num, str, pt };

// The vertical strip: both ends keep their size, the middle takes the rest.
{
  const tpl = data.tpl('VerticalResizeImageTemplate', 'gutter');
  const { w, cells } = vrCells(tpl, 362, h);
  assert.equal(w, 16);
  assert.deepEqual(cells.map((c) => [c.dy, c.dh, c.sy, c.sh]), [[0, 1, 95, 1], [1, 360, 75, 18], [361, 1, 95, 1]]);
  assert.deepEqual(vrCells(tpl, 1, h).cells.map((c) => c.dh), [1], 'too short for both ends: only the top');
  assert.equal(vrCells(tpl, 2, h).cells.length, 2, 'just both ends');
  assert.deepEqual(vrCells({ c: [] }, 50, h), { w: 0, cells: [] }, 'a template with no art draws nothing');
  console.log('ok: a vertical strip lays out its ends and middle');
}

const win = buildWindow(doc, data, 'Imgs', { onClick: (id) => clicked.push(id) });
const clicked = [];

// The horizontal strip is at its place, as wide as the control says, in three pieces.
{
  assert.equal(win.ctl.h.style.left, '5px');
  assert.equal(win.ctl.h.style.top, '6px');
  assert.equal(win.ctl.h.style.width, '100px');
  assert.equal(win.ctl.h.children.length, 3, 'left, middle and right');
  assert.equal(win.ctl.v.children.length, 3);
  assert.equal(win.ctl.v.style.height, '60px');
  assert.ok(win.ctl.v.children[1].style.backgroundImage.includes('a.png'));
  console.log('ok: insets and gutters are drawn at their size and place');
}

// The image button draws its picture on the stretchable button and reports its click by control id.
{
  const b = win.ctl['1001'];
  assert.equal(b.style.width, '30px');
  assert.ok(b.children.some((c) => c.style.backgroundImage && c.style.backgroundImage.includes('a.png')), 'the picture is on the button');
  b.listeners.pointerenter();
  b.listeners.pointerdown();
  b.listeners.pointerup();
  assert.deepEqual(clicked, ['1001']);
  console.log('ok: the image button shows its picture and reports its click');
}

// Icons and status markers take values by adapter.
{
  assert.equal(win.ctl.ic.style.width, '32px');
  assert.equal(win.ctl.ic.children.length, 0, 'empty until it has an icon');
  win.setIcon('pic', 634);
  assert.equal(win.ctl.ic.children.length, 1);
  win.setIcon('pic', 0);
  assert.equal(win.ctl.ic.children.length, 0, 'zero clears it');
  win.setIcon('nobody', 5); // an adapter nothing is bound to is ignored
  assert.equal(win.ctl.st.style.display, 'none');
  win.setStatus('pw', 3);
  assert.equal(win.ctl.st.style.display, 'block');
  win.setStatus('pw', 0);
  assert.equal(win.ctl.st.style.display, 'none');
  console.log('ok: icons and status markers follow their adapters');
}

// A dynamic image is a rectangle of a texture.
{
  const d = win.ctl.dyn;
  assert.equal(d.style.width, '50px');
  assert.equal(d.style.height, '40px');
  assert.equal(d.style.backgroundPosition, '-4px -8px');
  console.log('ok: a dynamic image shows its part of the texture');
}

// A rebuild keeps the controls and their adapters.
{
  win.resize(220, 130);
  win.setIcon('pic', 700);
  assert.equal(win.ctl.ic.children.length, 1);
  console.log('ok: resizing keeps the image controls working');
}
