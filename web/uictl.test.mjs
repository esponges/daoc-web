// Headless checks of the stock controls drawn by uictl.js.
//   node web/uictl.test.mjs
// What they look like is checked in a browser (/uidemo.html).

import assert from 'node:assert/strict';
import { UIData, buildWindow, child, num, str, pt } from './uiwin.js';
import { vrCells, clampTop, scrollThumb, topFromThumb, rowHeight } from './uictl.js';

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
    n('FullResizeImageTemplate', [n('Name', 'interior'), n('TopHeight', 4), n('MiddleHeight', 4), n('BottomHeight', 4), n('LeftWidth', 4), n('MiddleWidth', 4), n('RightWidth', 4),
      n('Texture', [n('TextureName', 'page1'), n('TopLeft', { X: 0, Y: 41 }), n('TopMiddle', { X: 41, Y: 41 }), n('TopRight', { X: 82, Y: 41 }), n('MiddleLeft', { X: 0, Y: 82 }), n('MiddleMiddle', { X: 41, Y: 82 }), n('MiddleRight', { X: 82, Y: 82 }), n('BottomLeft', { X: 0, Y: 123 }), n('BottomMiddle', { X: 41, Y: 123 }), n('BottomRight', { X: 82, Y: 123 })])]),
    n('ButtonTemplate', [n('Name', 'up'), n('Size', { X: 15, Y: 17 })]),
    n('ButtonTemplate', [n('Name', 'down'), n('Size', { X: 15, Y: 17 })]),
    n('VerticalResizeImageTemplate', [n('Name', 'thumb'), n('Width', 11), n('TopHeight', 8), n('RepeatHeight', 16), n('BottomHeight', 8),
      n('Texture', [n('TextureName', 'page1'), n('Top', { X: 227, Y: 219 }), n('Repeat', { X: 227, Y: 227 }), n('Bottom', { X: 227, Y: 243 })])]),
    n('VerticalScrollbarTemplate', [n('Name', 'vscroll'), n('UpButtonTemplate', 'up'), n('DownButtonTemplate', 'down'), n('ThumbVResizeTemplate', 'thumb'), n('GutterVResizeTemplate', 'gutter'), n('Width', 15), n('ThumbOffset', { X: 2, Y: 0 })]),
    n('ListBoxTemplate', [n('Name', 'lb'), n('ScrollbarTemplate', 'vscroll'), n('DrawScroll', 'true'), n('BackgroundTemplate', 'interior'), n('TextFont', [n('Name', 'arial9'), n('LineOffset', 2), n('TextLeftOffset', 4)])]),
    n('WindowTemplate', [
      n('Name', 'lists'), n('WindowId', 'Lists'), n('Width', 200), n('Height', 140),
      n('ListBoxDef', [n('TemplateName', 'lb'), n('ControlId', '1005'), n('AdapterName', 'items'), n('Width', 120), n('Height', 108), n('Position', { X: 10, Y: 20 })]),
    ]),
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


// The scroll arithmetic.
{
  assert.equal(clampTop(-3, 30, 5), 0);
  assert.equal(clampTop(99, 30, 5), 25, 'the last position that still fills the view');
  assert.equal(clampTop(4.6, 30, 5), 5);
  assert.equal(clampTop(3, 4, 5), 0, 'a short list never scrolls');
  assert.equal(scrollThumb({ total: 10, view: 10, top: 0, track: 200 }), null, 'everything fits: no thumb');
  assert.deepEqual(scrollThumb({ total: 100, view: 10, top: 0, track: 200 }), { h: 20, y: 0, maxTop: 90 });
  assert.deepEqual(scrollThumb({ total: 100, view: 10, top: 45, track: 200 }), { h: 20, y: 90, maxTop: 90 }, 'halfway down the list, halfway down the track');
  assert.equal(scrollThumb({ total: 100, view: 10, top: 90, track: 200 }).y, 180, 'at the end the thumb is at the bottom');
  assert.equal(scrollThumb({ total: 1000, view: 5, top: 0, track: 100 }).h, 16, 'never thinner than the minimum');
  assert.equal(topFromThumb(90, 200, 20, 100, 10), 45);
  assert.equal(topFromThumb(500, 200, 20, 100, 10), 90, 'a thumb dragged past the end stops at the end');
  assert.equal(topFromThumb(-5, 200, 20, 100, 10), 0);
  assert.equal(rowHeight(11, 16, 2), 18);
  assert.equal(rowHeight(11, 0, 2), 13);
  console.log('ok: the scroll arithmetic');
}

// A list box shows its rows, scrolls, selects and offers rows to drag.
{
  const lw = buildWindow(doc, data, 'Lists');
  const list = lw.list('items');
  assert.ok(list && lw.list('1005') === list, 'found by adapter and by control id');
  const body = (l) => l.el.children.find((c) => c.style.overflow === 'hidden');
  assert.equal(lw.list('nope'), undefined);
  const rows = Array.from({ length: 30 }, (_, i) => ({ text: 'Row ' + i, drag: i === 2 ? 'spell:7' : undefined, header: i === 0, on: i !== 3 }));
  list.setRows(rows);
  assert.equal(list.view, 5, 'the rows that fit: 100px of 18px rows');
  assert.equal(body(list).children.length, 5);
  assert.equal(body(list).children[0].dataset.row, '0');
  assert.ok(list.el.children.some((c) => c.style.left === '102px'), 'the scrollbar sits at the right edge');
  list.scrollTo(10);
  assert.equal(list.top, 10);
  assert.equal(body(list).children[0].dataset.row, '10');
  list.el.listeners.wheel({ deltaY: 100, preventDefault() {} });
  assert.equal(list.top, 13, 'the wheel moves three rows');
  list.el.listeners.wheel({ deltaY: -100, preventDefault() {} });
  list.scrollTo(999);
  assert.equal(list.top, 25, 'scrolling past the end stops at the end');
  list.scrollTo(0);

  // Selecting: a click on a row, not on a header.
  const picked = [];
  list.onSelect((i, row) => picked.push([i, row.text]));
  body(list).children[0].listeners.click(); // the header
  assert.deepEqual(picked, [], 'a header cannot be selected');
  body(list).children[1].listeners.click();
  assert.deepEqual(picked, [[1, 'Row 1']]);
  assert.equal(list.selected, 1);
  assert.equal(body(list).children[1].style.background, 'rgba(255,192,0,0.2)', 'the selected row is marked');
  const act = [];
  list.onActivate((i) => act.push(i));
  body(list).children[1].listeners.dblclick();
  assert.deepEqual(act, [1]);

  // Rows with a drag payload can be dragged.
  const dragRow = body(list).children[2];
  assert.equal(dragRow.draggable, true);
  const payload = {};
  dragRow.listeners.dragstart({ dataTransfer: { setData: (t, v) => { payload[t] = v; } } });
  assert.deepEqual(payload, { 'text/plain': 'spell:7' });
  assert.equal(body(list).children[1].draggable, undefined, 'plain rows are not draggable');

  // A short list has no thumb; setting fewer rows drops a selection that is gone.
  list.setRows(rows.slice(0, 3));
  assert.equal(list.selected, 1);
  list.setRows(rows.slice(0, 1));
  assert.equal(list.selected, -1);
  assert.equal(list.top, 0);
  list.setRows([]);
  assert.equal(body(list).children.length, 0);
  console.log('ok: a list box shows, scrolls, selects and drags its rows');
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
