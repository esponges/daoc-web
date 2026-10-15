// Headless checks of the interface renderer's data handling.
//   node web/uiwin.test.mjs
//
// The window-building code is run against a stand-in document; what it draws
// is looked at in a browser.

import { readFileSync, existsSync } from 'node:fs';
import { UIData, nineSlice, layoutText, place, pngPath, buildWindow, child, num, str } from './uiwin.js';

let failed = 0;
const check = (name, ok, extra = '') => {
  if (!ok) {
    failed++;
    console.log(`FAIL ${name} ${extra}`);
  } else console.log(`ok   ${name}`);
};
const eq = (a, b) => JSON.stringify(a) === JSON.stringify(b);

// n builds a tree node: n('X', {A: 1}) or n('X', 'text') or n('X', [children]).
const n = (t, body) => {
  if (Array.isArray(body)) return { t, c: body };
  if (body && typeof body === 'object') return { t, c: Object.entries(body).map(([k, v]) => n(k, v)) };
  return { t, x: String(body) };
};

const layout = {
  skin: 'test',
  sizes: { 'atlantis/a.tga': [256, 256] },
  bitmap: {
    f: { file: 'fonts/f.tga', h: 10, g: [{ c: 65, x: 0, y: 0, w: 6 }, { c: 66, x: 6, y: 0, w: 5 }, { c: 63, x: 11, y: 0, w: 4 }] },
  },
  elements: [
    n('Texture', { Name: 'page1', File: 'atlantis/a.tga' }),
    n('TTFFont', { Name: 'chat_small', File: 'fonts/MyriadPro-Regular.ttf', Height: 10 }),
    n('FullResizeImageTemplate', {
      Name: 'box',
      TopHeight: 4, MiddleHeight: 2, BottomHeight: 4,
      LeftWidth: 3, MiddleWidth: 2, RightWidth: 3,
      Texture: [
        n('TextureName', 'page1'),
        n('TopLeft', { X: 10, Y: 20 }), n('TopMiddle', { X: 13, Y: 20 }), n('TopRight', { X: 15, Y: 20 }),
        n('MiddleLeft', { X: 10, Y: 24 }), n('MiddleMiddle', { X: 13, Y: 24 }), n('MiddleRight', { X: 15, Y: 24 }),
        n('BottomLeft', { X: 10, Y: 26 }), n('BottomMiddle', { X: 13, Y: 26 }), n('BottomRight', { X: 15, Y: 26 }),
      ],
    }),
    n('ImageAreaTemplate', { Name: 'dot', Size: { X: 8, Y: 6 }, TopLeft: { X: 1, Y: 2 }, TextureName: 'page1' }),
    n('ButtonTemplate', {
      Name: 'btn', Size: { X: 14, Y: 14 },
      Texture: [n('TextureName', 'page1'), n('Normal', { X: 0, Y: 43 }), n('Pressed', { X: 14, Y: 43 }), n('NormalHighlit', { X: 0, Y: 43 }), n('PressedHighlit', { X: 28, Y: 43 })],
    }),
    n('HorizontalResizeImageTemplate', { Name: 'red_full', Height: 6, LeftWidth: 0, RepeatWidth: 100, RightWidth: 0, Texture: [n('TextureName', 'page1'), n('Left', { X: 154, Y: 219 }), n('Repeat', { X: 154, Y: 219 }), n('Right', { X: 154, Y: 219 })] }),
    n('HorizontalResizeImageTemplate', { Name: 'red_back', Height: 6, LeftWidth: 0, RepeatWidth: 100, RightWidth: 0, Texture: [n('TextureName', 'page1'), n('Left', { X: 154, Y: 213 }), n('Repeat', { X: 154, Y: 213 }), n('Right', { X: 154, Y: 213 })] }),
    n('StatusBarTemplate', { Name: 'red_status', ForegroundHResizeTemplate: 'red_full', BackgroundHResizeTemplate: 'red_back', Height: 16, ForegroundOffset: { X: 0, Y: 0 } }),
    n('WindowTemplate', [
      n('Name', 'bars'), n('WindowId', 'Bars'), n('Width', 160), n('Height', 90),
      n('StatusBarDef', [n('TemplateName', 'red_status'), n('AdapterName', 'hits'), n('Width', 125), n('Position', { X: 26, Y: 31 })]),
      n('ScalarLabelDef', [n('FontName', 'f'), n('Adapter', 'hits'), n('Width', 40), n('Height', 12), n('TextCentered', 'true'), n('Position', { X: 5, Y: 28 })]),
      n('ScalarLabelDef', [n('FontName', 'f'), n('Adapter', 'hits'), n('Width', 40), n('Height', 12), n('Position', { X: 5, Y: 50 })]),
    ]),
    n('WindowTemplate', [
      n('Name', 'win'), n('WindowId', 'Win'), n('Width', 100), n('Height', 60),
      n('FullResizeImageDef', [n('TemplateName', 'box'), n('ControlId', 'bg'), n('Width', 100), n('Height', 60), n('Position', { X: 0, Y: 0 }), n('Alignment', { GrowWidth: 'true', GrowHeight: 'true' })]),
      n('ButtonDef', [n('TemplateName', 'btn'), n('ControlId', 'close'), n('Position', { X: 3, Y: 4 }), n('Alignment', { OffsetRight: 'true' })]),
      n('LabelDef', [n('FontName', 'f'), n('ControlId', 'title'), n('Width', 50), n('Height', 10), n('Position', { X: 5, Y: 5 })]),
    ]),
  ],
};

// A stand-in document: enough surface for the builders. Setting textContent
// to '' clears the children, as in a browser.
const doc = {
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

const data = new UIData(layout, 'data/ui/test');

check('base gets a slash', data.base === 'data/ui/test/');
check('texture resolves', eq(data.texture('page1'), { url: 'data/ui/test/atlantis/a.png', w: 256, h: 256 }), JSON.stringify(data.texture('page1')));
check('unknown texture', data.texture('nope') === undefined);
check('png path is lower case', pngPath('Fonts\\MyriadPro.TGA') === 'fonts/myriadpro.png');
check('window by id', data.windows.has('Win'));
check('ttf font by name', data.fonts.get('chat_small').file === 'fonts/myriadpro-regular.ttf');

// Nine-slice: corners keep their size, the middle stretches.
const cells = nineSlice(data.tpl('FullResizeImageTemplate', 'box'), 20, 30);
check('nine cells', cells.length === 9, String(cells.length));
const mid = cells[4];
check('middle stretches', mid.dx === 3 && mid.dy === 4 && mid.dw === 14 && mid.dh === 22, JSON.stringify(mid));
check('middle source', mid.sx === 13 && mid.sy === 24 && mid.sw === 2 && mid.sh === 2);
check('corner keeps size', eq([cells[8].dx, cells[8].dy, cells[8].dw, cells[8].dh], [17, 26, 3, 4]), JSON.stringify(cells[8]));
const tiny = nineSlice(data.tpl('FullResizeImageTemplate', 'box'), 4, 6);
check('smaller than the corners drops the middle', tiny.length < 9 && tiny.every((c) => c.dw > 0 && c.dh > 0));

// Text.
const font = data.bitmapFont('f');
const lay = layoutText(font, 'AB A?x');
check('advances add up', lay.glyphs.length === 4 && lay.glyphs[1].x === 6 && lay.glyphs[2].x === 11 + 3, JSON.stringify(lay.glyphs.map((g) => g.x)));
check('width counts space and unknown', lay.width === 6 + 5 + 3 + 6 + 4 + 4, String(lay.width));
check('no font', data.bitmapFont('missing') === undefined);
const dash = layoutText({ h: 10, glyphs: new Map([[0x97, { c: 0x97, x: 0, y: 0, w: 7 }]]) }, '—');
check('cp1252 mapping', dash.glyphs.length === 1 && dash.width === 7, JSON.stringify(dash));

// Placement.
const win = data.windows.get('Win');
const bg = child(win, 'FullResizeImageDef');
check('grow width and height', eq(place(bg, 140, 90, 100, 60), { x: 0, y: 0, w: 140, h: 90 }));
const close = child(win, 'ButtonDef');
check('offset from the right', place(close, 140, 90, 100, 60).x === 140 - 3 - 0);

// Building.
const w = buildWindow(doc, data, 'Win');
check('window size', w.el.style.width === '100px' && w.el.style.height === '60px');
check('controls by id', 'bg' in w.ctl && 'close' in w.ctl && 'title' in w.ctl);
check('frame has nine cells', w.ctl.bg.children.length === 9, String(w.ctl.bg.children.length));
check('button is sprited', w.ctl.close.style.backgroundImage.includes('a.png'));
w.resize(140, 90);
check('resize rebuilds', w.el.style.width === '140px' && w.ctl.bg.style.width === '140px' && w.ctl.bg.children.length === 9);
check('unknown window throws', (() => { try { buildWindow(doc, data, 'Nope'); } catch { return true; } return false; })());

// Bars and labels bound to adapters.
const bw = buildWindow(doc, data, 'Bars');
check('adapter known', bw.has('hits') && !bw.has('nope'));
bw.setBar('hits', 0.5);
const barEl = bw.el.children[0];
check('bar clips to the fraction', barEl.children[1].style.width === '62.5px', barEl.children[1].style.width);
check('bar back is full width', barEl.children[0].style.width === '125px');
bw.setBar('hits', 7);
check('bar clamps above one', barEl.children[1].style.width === '125px');
bw.setBar('hits', NaN);
check('bar treats NaN as empty', barEl.children[1].style.width === '0px');
bw.setText('hits', 42);
check('every label on the adapter updates', bw.el.children[1].children.length === 1 && bw.el.children[2].children.length === 1, bw.el.children.map((c) => c.children.length).join());
bw.setText('hits', '');
check('empty text clears', bw.el.children[1].children.length === 0);
bw.resize(180, 90);
check('adapters survive a resize', bw.has('hits'));
bw.setText('nope', 'x');

// Button states.
let clicks = 0;
const b = buildWindow(doc, data, 'Win', { onClick: () => clicks++ }).ctl.close;
b.listeners.pointerenter(); b.listeners.pointerdown();
check('pressed look', b.style.backgroundPosition === '-28px -43px', b.style.backgroundPosition);
b.listeners.pointerup();
check('click reported once', clicks === 1);
b.listeners.pointerdown(); b.listeners.pointerleave(); b.listeners.pointerup();
check('release off the button is no click', clicks === 1);

// The real converted chat window, when the converter has been run.
const real = 'web/data/ui/atlantis/layout.json';
if (existsSync(real)) {
  const rd = new UIData(JSON.parse(readFileSync(real, 'utf8')), 'data/ui/atlantis/');
  const chat = rd.windows.get('Chat');
  check('real chat window', chat && num(chat, 'Width') === 350 && num(chat, 'Height') === 200);
  const cw = buildWindow(doc, rd, 'Chat');
  check('real chat controls', '1000' in cw.ctl && '1002' in cw.ctl, Object.keys(cw.ctl).join());
  check('real chat frame drawn', cw.ctl['1000'].children.length === 9, String(cw.ctl['1000'].children.length));
  check('real bitmap fonts', rd.bitmapFont('arial11') && rd.bitmapFont('arial11').glyphs.size >= 223);
  check('real icon sheets', rd.layout.icons.length === 19);
} else console.log('skip real layout (run go run ./cmd/uiconv)');

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
