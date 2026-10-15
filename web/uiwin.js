// Windows built from the game's own interface data.
//
// `go run ./cmd/uiconv` writes web/data/ui/<skin>/layout.json, the textures
// as PNG, the fonts and the icon sheets. This file reads that and builds the
// DOM: nine-slice frames, sprite buttons with their states, bitmap
// and TrueType text, and windows laid out from their WindowTemplate.
//
// The layout is a generic tree {t: tag, x: text, c: children}. Everything
// that does not touch the DOM is plain functions over that tree, checked
// headlessly in uiwin.test.mjs.

// --- reading the tree ---

export const child = (n, tag) => (n && n.c ? n.c.find((c) => c.t === tag) : undefined);
export const children = (n, tag) => (n && n.c ? n.c.filter((c) => c.t === tag) : []);
export const str = (n, tag, d = '') => {
  const c = child(n, tag);
  return c && c.x !== undefined ? c.x : d;
};
export const num = (n, tag, d = 0) => {
  const v = parseFloat(str(n, tag, ''));
  return Number.isFinite(v) ? v : d;
};
// A point element {X, Y}.
export const pt = (n, tag) => {
  const c = child(n, tag);
  return c ? { x: num(c, 'X'), y: num(c, 'Y') } : undefined;
};
export const flag = (n, tag) => str(n, tag).toLowerCase() === 'true';

// --- the data ---

export class UIData {
  // layout is the parsed layout.json, base the URL its files sit under.
  constructor(layout, base = '') {
    this.layout = layout;
    this.base = base.endsWith('/') || base === '' ? base : base + '/';
    this.templates = new Map(); // "Kind/name" -> node
    this.windows = new Map(); // WindowId -> WindowTemplate node
    this.textures = new Map(); // texture name -> file
    this.fonts = new Map(); // TTF name -> {file, height}
    for (const e of layout.elements) {
      const name = str(e, 'Name');
      if (e.t === 'Texture' && name) this.textures.set(name, str(e, 'File'));
      else if (e.t === 'TTFFont' && name) this.fonts.set(name, { file: str(e, 'File').toLowerCase(), height: num(e, 'Height', 12) });
      else if (e.t === 'WindowTemplate') {
        const id = str(e, 'WindowId');
        if (id) this.windows.set(id, e);
        if (name) this.templates.set('WindowTemplate/' + name, e);
      } else if (name && e.t.endsWith('Template')) this.templates.set(e.t + '/' + name, e);
    }
  }

  tpl(kind, name) {
    return this.templates.get(kind + '/' + name);
  }

  // texture resolves a texture name to its PNG url and size.
  texture(name) {
    const file = this.textures.get(name);
    if (!file) return undefined;
    const size = this.layout.sizes && this.layout.sizes[file];
    return { url: this.base + pngPath(file), w: size ? size[0] : 0, h: size ? size[1] : 0 };
  }

  // art resolves a converted file by its path in the install, such as
  // "skin1/pperdoll.bmp": the classic skin's window art.
  art(file) {
    const size = this.layout.sizes && this.layout.sizes[file];
    if (!size) return undefined;
    return { url: this.base + pngPath(file), w: size[0], h: size[1] };
  }

  bitmapFont(name) {
    const f = this.layout.bitmap && this.layout.bitmap[name];
    if (!f) return undefined;
    const size = this.layout.sizes && this.layout.sizes[f.file];
    const byCode = new Map(f.g.map((g) => [g.c, g]));
    return { url: this.base + pngPath(f.file), h: f.h, glyphs: byCode, w: size ? size[0] : 256, ht: size ? size[1] : 256 };
  }
}

// pngPath is where uiconv put a texture: lower case, with a .png extension.
export function pngPath(file) {
  const f = file.toLowerCase().replace(/\\/g, '/');
  const i = f.lastIndexOf('.');
  return (i < 0 ? f : f.slice(0, i)) + '.png';
}

// --- icons ---

// iconSheet finds the converted sheet that holds icon n: sheets are 10 by 10
// grids of 32-pixel icons, one file per hundred numbers.
export function iconCell(data, n) {
  const first = n - (n % 100);
  const sheet = (data.layout.icons || []).find((s) => s.first === first && s.kind === 'cbt') ||
    (data.layout.icons || []).find((s) => s.first === first);
  if (!sheet) return undefined;
  const k = n % 100;
  return { url: data.base + sheet.file, sx: (k % 10) * 32, sy: Math.floor(k / 10) * 32 };
}

// Style icons: the imported numbers do not match this install's sheets (see
// the stock-ui design notes), so a style is drawn with a picture of its
// weapon line, picked from the combat sheets.
export const SPEC_ICON = {
  Sword: 622, Slash: 622, Thrust: 622, 'Left Axe': 700, Axe: 700, Hammer: 634, Crush: 634, Shield: 705,
};
export const DEFAULT_STYLE_ICON = 705;

// Item pictures: item numbers in the imported data do not match this
// install's sheets either, so an item is drawn by what it is: the piece it
// is worn as, or its weapon line.
const WORN_ICON = { 42: 1204, 43: 1221, 44: 1211, 45: 1225, 46: 1215, 47: 1200 };
const LINE_ICON = { Sword: 622, Hammer: 634, Axe: 700 };
export const SHIELD_ICON = 705;
export const MISC_ICON = 1236;

export function itemIconNumber(it) {
  if (it.k === 'shield') return SHIELD_ICON;
  if (it.k === 'weapon') return LINE_ICON[it.line] ?? 634;
  if (it.k === 'armour') return WORN_ICON[it.w] ?? 1204;
  return MISC_ICON;
}

export function styleIconNumber(spec) {
  return SPEC_ICON[spec] ?? DEFAULT_STYLE_ICON;
}

// icon builds a 32-pixel icon element for icon number n, or an empty frame
// when the install has no sheet for it.
export function icon(doc, data, n, x = 0, y = 0) {
  const e = box(doc, x, y, 32, 32);
  const c = iconCell(data, n);
  if (c) {
    const s = e.style;
    s.backgroundImage = `url("${c.url}")`;
    s.backgroundPosition = `${px(-c.sx)} ${px(-c.sy)}`;
    s.backgroundRepeat = 'no-repeat';
    s.imageRendering = 'pixelated';
  }
  return e;
}

// --- nine-slice frames ---

// nineSlice lays a FullResizeImageTemplate out at w by h: the corners keep
// their size and the edges and middle stretch. Each cell is
// {dx, dy, dw, dh, tex, sx, sy, sw, sh}: draw the source rectangle of the
// texture into the destination one.
export function nineSlice(tpl, w, h) {
  const texNode = child(tpl, 'Texture');
  if (!texNode) return [];
  const tex = str(texNode, 'TextureName');
  const L = num(tpl, 'LeftWidth'), M = num(tpl, 'MiddleWidth'), R = num(tpl, 'RightWidth');
  const T = num(tpl, 'TopHeight'), MH = num(tpl, 'MiddleHeight'), B = num(tpl, 'BottomHeight');
  const midW = Math.max(0, w - L - R), midH = Math.max(0, h - T - B);
  const cols = [
    { dx: 0, dw: Math.min(L, w), sw: L },
    { dx: L, dw: midW, sw: M },
    { dx: L + midW, dw: R, sw: R },
  ];
  const rows = [
    { dy: 0, dh: Math.min(T, h), sh: T },
    { dy: T, dh: midH, sh: MH },
    { dy: T + midH, dh: B, sh: B },
  ];
  const names = [
    ['TopLeft', 'TopMiddle', 'TopRight'],
    ['MiddleLeft', 'MiddleMiddle', 'MiddleRight'],
    ['BottomLeft', 'BottomMiddle', 'BottomRight'],
  ];
  const cells = [];
  for (let r = 0; r < 3; r++) {
    for (let c = 0; c < 3; c++) {
      const src = pt(texNode, names[r][c]);
      const col = cols[c], row = rows[r];
      if (!src || col.dw <= 0 || row.dh <= 0 || col.sw <= 0 || row.sh <= 0) continue;
      cells.push({ dx: col.dx, dy: row.dy, dw: col.dw, dh: row.dh, tex, sx: src.x, sy: src.y, sw: col.sw, sh: row.sh });
    }
  }
  return cells;
}

// --- bitmap text ---

const SPACE_FRACTION = 0.3; // the sheets hold no space glyph

// The sheets are in Windows-1252, which differs from Unicode between 0x80 and 0x9F.
const CP1252 = {
  0x20ac: 0x80, 0x201a: 0x82, 0x0192: 0x83, 0x201e: 0x84, 0x2026: 0x85, 0x2020: 0x86, 0x2021: 0x87,
  0x02c6: 0x88, 0x2030: 0x89, 0x0160: 0x8a, 0x2039: 0x8b, 0x0152: 0x8c, 0x017d: 0x8e,
  0x2018: 0x91, 0x2019: 0x92, 0x201c: 0x93, 0x201d: 0x94, 0x2022: 0x95, 0x2013: 0x96, 0x2014: 0x97,
  0x02dc: 0x98, 0x2122: 0x99, 0x0161: 0x9a, 0x203a: 0x9b, 0x0153: 0x9c, 0x017e: 0x9e, 0x0178: 0x9f,
};

// layoutText places the glyphs of a line from a bitmap font. Characters the
// sheet does not hold take the width of a question mark and draw nothing.
export function layoutText(font, text) {
  const out = [];
  let x = 0;
  for (const ch of String(text)) {
    const cp = ch.codePointAt(0);
    const code = CP1252[cp] || cp;
    if (ch === ' ') {
      x += Math.max(3, Math.round(font.h * SPACE_FRACTION));
      continue;
    }
    const g = font.glyphs.get(code);
    if (!g) {
      const q = font.glyphs.get(63);
      x += q ? q.w : 4;
      continue;
    }
    out.push({ g, x });
    x += g.w;
  }
  return { glyphs: out, width: x, height: font.h };
}

// --- colours ---

export function rgba(node) {
  if (!node) return undefined;
  return `rgba(${num(node, 'R', 255)}, ${num(node, 'G', 255)}, ${num(node, 'B', 255)}, ${num(node, 'A', 255) / 255})`;
}

// --- DOM ---

const px = (n) => Math.round(n * 1000) / 1000 + 'px';

function box(doc, x, y, w, h) {
  const e = doc.createElement('div');
  const s = e.style;
  s.position = 'absolute';
  s.left = px(x);
  s.top = px(y);
  s.width = px(w);
  s.height = px(h);
  return e;
}

// spriteCell draws a source rectangle of a texture into e, scaled to fit.
function paint(e, data, cell) {
  const t = data.texture(cell.tex);
  if (!t) return;
  const kx = cell.dw / cell.sw, ky = cell.dh / cell.sh;
  const s = e.style;
  s.backgroundImage = `url("${t.url}")`;
  s.backgroundRepeat = 'no-repeat';
  s.backgroundSize = `${px(t.w * kx)} ${px(t.h * ky)}`;
  s.backgroundPosition = `${px(-cell.sx * kx)} ${px(-cell.sy * ky)}`;
  s.imageRendering = kx === 1 && ky === 1 ? 'pixelated' : 'auto';
}

// frame builds a nine-slice frame element of a FullResizeImageTemplate.
export function frame(doc, data, tplName, w, h) {
  const e = box(doc, 0, 0, w, h);
  const tpl = data.tpl('FullResizeImageTemplate', tplName) || data.tpl('BackgroundTemplate', tplName);
  if (tpl) {
    for (const c of nineSlice(tpl, w, h)) {
      const cell = box(doc, c.dx, c.dy, c.dw, c.dh);
      paint(cell, data, c);
      e.appendChild(cell);
    }
  }
  return e;
}

// sprite builds an ImageAreaTemplate: a fixed rectangle of a texture.
export function sprite(doc, data, tplName, x = 0, y = 0) {
  const tpl = data.tpl('ImageAreaTemplate', tplName);
  if (!tpl) return box(doc, x, y, 0, 0);
  const size = pt(tpl, 'Size') || { x: 0, y: 0 };
  const tl = pt(tpl, 'TopLeft') || { x: 0, y: 0 };
  const e = box(doc, x, y, size.x, size.y);
  paint(e, data, { tex: str(tpl, 'TextureName'), sx: tl.x, sy: tl.y, sw: size.x, sh: size.y, dw: size.x, dh: size.y });
  return e;
}

// text builds a line of text. Bitmap fonts are drawn glyph by glyph as a mask
// painted in the colour, since the sheets are white; TrueType fonts are plain
// text in the family uiFonts registered.
export function text(doc, data, fontName, value, color = 'rgb(255,255,255)') {
  const bf = data.bitmapFont(fontName);
  if (bf) {
    const lay = layoutText(bf, value);
    const e = box(doc, 0, 0, lay.width, lay.height);
    for (const { g, x } of lay.glyphs) {
      const gl = box(doc, x, 0, g.w, bf.h);
      const s = gl.style;
      s.backgroundColor = color;
      for (const p of ['webkitMask', 'mask']) {
        s[p + 'Image'] = `url("${bf.url}")`;
        s[p + 'Repeat'] = 'no-repeat';
        s[p + 'Position'] = `${px(-g.x)} ${px(-g.y)}`;
        s[p + 'Size'] = `${bf.w}px ${bf.ht}px`;
      }
      s.imageRendering = 'pixelated';
      e.appendChild(gl);
    }
    e.dataset.text = String(value);
    return e;
  }
  const f = data.fonts.get(fontName);
  const e = doc.createElement('div');
  e.textContent = String(value);
  e.style.position = 'absolute';
  e.style.whiteSpace = 'pre';
  e.style.color = color;
  e.style.font = f ? `${f.height}px "${fontFamily(fontName)}", sans-serif` : '11px sans-serif';
  return e;
}

export const fontFamily = (name) => 'daoc-' + name;

// registerFonts adds the skin's TrueType fonts to the page.
export function registerFonts(doc, data) {
  const css = [];
  for (const [name, f] of data.fonts) {
    css.push(`@font-face { font-family: "${fontFamily(name)}"; src: url("${data.base}${f.file}"); }`);
  }
  const style = doc.createElement('style');
  style.textContent = css.join('\n');
  doc.head.appendChild(style);
  return style;
}

// button builds a ButtonTemplate: a sprite with normal, highlit, pressed and
// disabled looks that follows the pointer. It reports a click when released
// over itself.
export function button(doc, data, tplName, x, y, onClick) {
  const tpl = data.tpl('ButtonTemplate', tplName);
  const size = (tpl && pt(tpl, 'Size')) || { x: 0, y: 0 };
  const e = box(doc, x, y, size.x, size.y);
  e.style.cursor = 'pointer';
  const tn = tpl && child(tpl, 'Texture');
  const look = (state) => {
    const p = tn && (pt(tn, state) || pt(tn, 'Normal'));
    if (!p) return;
    paint(e, data, { tex: str(tn, 'TextureName'), sx: p.x, sy: p.y, sw: size.x, sh: size.y, dw: size.x, dh: size.y });
  };
  let over = false, down = false;
  const refresh = () => look(down ? (over ? 'PressedHighlit' : 'Pressed') : over ? 'NormalHighlit' : 'Normal');
  e.addEventListener('pointerenter', () => ((over = true), refresh()));
  e.addEventListener('pointerleave', () => ((over = false), refresh()));
  e.addEventListener('pointerdown', () => ((down = true), refresh()));
  e.addEventListener('pointerup', () => {
    const click = down && over;
    down = false;
    refresh();
    if (click && onClick) onClick();
  });
  refresh();
  e.look = look;
  return e;
}

// strip finds a HorizontalResizeImageTemplate by name and returns what is
// needed to draw it: its source point and size in the texture.
function strip(data, name) {
  const tpl = ['HorizontalResizeImageTemplate', 'ForegroundHResizeTemplate', 'BackgroundHResizeTemplate']
    .map((k) => data.tpl(k, name)).find(Boolean);
  const tn = tpl && child(tpl, 'Texture');
  const src = tn && (pt(tn, 'Repeat') || pt(tn, 'Left'));
  if (!src) return undefined;
  return { tex: str(tn, 'TextureName'), sx: src.x, sy: src.y, sw: num(tpl, 'RepeatWidth', 100) || 100, sh: num(tpl, 'Height', 6) };
}

// statusBar builds a StatusBarTemplate w pixels wide: a back strip with the
// fore strip over it, clipped to a fraction. set(f) takes 0 to 1.
export function statusBar(doc, data, tplName, w) {
  const tpl = data.tpl('StatusBarTemplate', tplName);
  const fg = tpl && strip(data, str(tpl, 'ForegroundHResizeTemplate'));
  const bg = tpl && strip(data, str(tpl, 'BackgroundHResizeTemplate'));
  const h = (fg && fg.sh) || (bg && bg.sh) || num(tpl, 'Height', 6);
  const root = box(doc, 0, 0, w, h);
  const draw = (e, st) => {
    if (st) paint(e, data, { ...st, dw: w, dh: h, sh: st.sh });
  };
  const back = box(doc, 0, 0, w, h);
  draw(back, bg);
  const clip = box(doc, 0, 0, w, h);
  clip.style.overflow = 'hidden';
  const fore = box(doc, 0, 0, w, h);
  draw(fore, fg);
  clip.appendChild(fore);
  root.append(back, clip);
  const set = (f) => {
    clip.style.width = px(Math.max(0, Math.min(1, Number.isFinite(f) ? f : 0)) * w);
  };
  set(0);
  return { el: root, set };
}

// labelText builds a text control of a fixed width whose content can change.
// Alignment follows the layout: TextCentered, EndAligned, else left.
export function labelText(doc, data, c) {
  const w = num(c, 'Width'), h = num(c, 'Height');
  const font = str(c, 'FontName');
  const baseColor = rgba(child(c, 'Color')) || 'rgb(255,255,255)';
  const root = box(doc, 0, 0, w, h);
  const set = (value, color) => {
    root.textContent = '';
    if (value === '' || value === undefined || value === null) return;
    const t = text(doc, data, font, value, color || baseColor);
    const bf = data.bitmapFont(font);
    const tw = bf ? layoutText(bf, value).width : 0;
    if (bf) {
      t.style.left = px(flag(c, 'TextCentered') ? (w - tw) / 2 : flag(c, 'EndAligned') ? w - tw : 0);
      t.style.top = px(Math.max(0, (h - bf.h) / 2));
    } else {
      t.style.left = '0px';
      t.style.width = px(w);
      t.style.textAlign = flag(c, 'TextCentered') ? 'center' : flag(c, 'EndAligned') ? 'right' : 'left';
    }
    root.appendChild(t);
  };
  return { el: root, set };
}

// hrCells lays a HorizontalResizeImageTemplate out at width w: the left and
// right ends keep their size and the middle stretches. Cells are
// {dx, dw, tex, sx, sy, sw, sh} with a common height.
export function hrCells(tpl, w) {
  const tn = child(tpl, 'Texture');
  if (!tn) return { h: 0, cells: [] };
  const tex = str(tn, 'TextureName');
  const L = num(tpl, 'LeftWidth'), M = num(tpl, 'RepeatWidth', 1) || 1, R = num(tpl, 'RightWidth'), h = num(tpl, 'Height');
  const left = pt(tn, 'Left'), mid = pt(tn, 'Repeat'), right = pt(tn, 'Right');
  const cells = [];
  const midW = Math.max(0, w - L - R);
  if (left && L > 0) cells.push({ dx: 0, dw: Math.min(L, w), tex, sx: left.x, sy: left.y, sw: L, sh: h });
  if (mid && midW > 0) cells.push({ dx: L, dw: midW, tex, sx: mid.x, sy: mid.y, sw: M, sh: h });
  if (right && R > 0 && w >= L + R) cells.push({ dx: w - R, dw: R, tex, sx: right.x, sy: right.y, sw: R, sh: h });
  return { h, cells };
}

function hrImage(doc, data, name, w) {
  const tpl = data.tpl('HorizontalResizeImageTemplate', name);
  const e = box(doc, 0, 0, w, 0);
  if (!tpl) return e;
  const { h, cells } = hrCells(tpl, w);
  e.style.height = px(h);
  for (const c of cells) {
    const cell = box(doc, c.dx, 0, c.dw, h);
    paint(cell, data, { ...c, dy: 0, dh: h });
    e.appendChild(cell);
  }
  return e;
}

// hrButton builds a HorizontalResizeButtonTemplate w pixels wide with a
// label: stretchable ends and middle in its normal, highlit and pressed
// looks, the label in the template's font and state colours. It reports a
// click when released over itself. set(label) changes the text; disable()
// greys it.
export function hrButton(doc, data, tplName, label, w, onClick) {
  const tpl = data.tpl('HorizontalResizeButtonTemplate', tplName);
  const font = tpl && child(tpl, 'Font');
  const fontName = font ? str(font, 'Name') : 'arial9';
  const colors = {
    Normal: rgba(child(font, 'ColorNormal')) || 'rgb(192,192,192)',
    Pressed: rgba(child(font, 'ColorPressed')) || 'rgb(255,192,0)',
    Highlit: rgba(child(font, 'ColorHighlit')) || 'rgb(255,255,255)',
    Disabled: rgba(child(font, 'ColorDisabled')) || 'rgb(128,128,128)',
  };
  const h = (tpl && num(tpl, 'Height')) || 16;
  const root = box(doc, 0, 0, w, h);
  root.style.cursor = 'pointer';
  const states = {};
  for (const s of ['Normal', 'Pressed', 'NormalHighlit', 'PressedHighlit', 'Disabled']) {
    const name = tpl ? str(tpl, s) : '';
    const img = hrImage(doc, data, name && name !== 'undefined' ? name : str(tpl, 'Normal'), w);
    img.style.display = 'none';
    states[s] = img;
    root.appendChild(img);
  }
  const labelBox = box(doc, 0, 0, w, h);
  labelBox.style.pointerEvents = 'none';
  root.appendChild(labelBox);
  let text_ = label, over = false, down = false, disabled = false, selected = false;
  const refresh = () => {
    const key = disabled ? 'Disabled' : down || selected ? (over ? 'PressedHighlit' : 'Pressed') : over ? 'NormalHighlit' : 'Normal';
    for (const [s, img] of Object.entries(states)) img.style.display = s === key ? 'block' : 'none';
    const colorKey = disabled ? 'Disabled' : down || selected ? 'Pressed' : over ? 'Highlit' : 'Normal';
    labelBox.textContent = '';
    if (text_) {
      const t = text(doc, data, fontName, text_, colors[colorKey]);
      const bf = data.bitmapFont(fontName);
      if (bf) {
        const tw = layoutText(bf, text_).width;
        t.style.left = px(Math.max(0, (w - tw) / 2));
        t.style.top = px(Math.max(0, (h - bf.h) / 2));
      } else {
        t.style.left = '0px';
        t.style.width = px(w);
        t.style.textAlign = 'center';
        t.style.top = px(Math.max(0, (h - 12) / 2));
      }
      labelBox.appendChild(t);
    }
  };
  root.addEventListener('pointerenter', () => ((over = true), refresh()));
  root.addEventListener('pointerleave', () => ((over = false), refresh()));
  root.addEventListener('pointerdown', () => {
    if (!disabled) ((down = true), refresh());
  });
  root.addEventListener('pointerup', () => {
    const click = down && over && !disabled;
    down = false;
    refresh();
    if (click && onClick) onClick();
  });
  refresh();
  return {
    el: root,
    set(l) {
      text_ = l;
      refresh();
    },
    // select holds the pressed look, as a chosen tab does.
    select(on = true) {
      selected = on;
      refresh();
    },
    disable(on = true) {
      disabled = on;
      root.style.cursor = on ? 'default' : 'pointer';
      refresh();
    },
  };
}

// --- compass ---

// compassOffset is how far to scroll the compass strip, in destination pixels,
// to show heading deg (clockwise from north) in the middle of a window w wide.
// The strip is stripW source pixels for a full turn with north at northX, and
// is drawn scaled by k; it wraps, so the result is taken modulo one turn.
export function compassOffset(deg, w, stripW, northX, k) {
  const turn = stripW * k;
  const x = (northX + (((deg % 360) + 360) % 360) / 360 * stripW) * k;
  return ((x - w / 2) % turn + turn) % turn;
}

// compass builds a CompassControlTemplate: the heading strip, scrolled to a
// heading and wrapping. The strip has north a half-step in from its left
// edge, so that each of the eight points sits in the middle of its own cell.
export function compass(doc, data, tplName, w, h) {
  const tpl = data.tpl('CompassControlTemplate', tplName);
  const t = tpl && data.texture(str(tpl, 'Texture'));
  const e = box(doc, 0, 0, w, h);
  if (!t) return { el: e, set() {} };
  const k = h / t.h;
  const northX = t.w / 16; // 32 of 512: the middle of the first of eight cells
  Object.assign(e.style, {
    backgroundImage: `url("${t.url}")`,
    backgroundRepeat: 'repeat-x',
    backgroundSize: `${px(t.w * k)} ${px(h)}`,
    imageRendering: 'auto',
  });
  const set = (deg) => {
    e.style.backgroundPosition = `${px(-compassOffset(deg, w, t.w, northX, k))} 0px`;
  };
  set(0);
  return { el: e, set };
}

// --- windows ---

// placeIn resolves a control's Position against the window size, honouring
// Alignment: OffsetRight / OffsetBottom measure from the right and bottom
// edges, GrowWidth / GrowHeight stretch with the window.
export function place(ctl, winW, winH, baseW, baseH) {
  const p = pt(ctl, 'Position') || { x: 0, y: 0 };
  const al = child(ctl, 'Alignment');
  const w0 = num(ctl, 'Width', num(ctl, 'width', 0)), h0 = num(ctl, 'Height', num(ctl, 'height', 0));
  const growW = al && flag(al, 'GrowWidth'), growH = al && flag(al, 'GrowHeight');
  const w = growW ? w0 + (winW - baseW) : w0;
  const h = growH ? h0 + (winH - baseH) : h0;
  const x = al && flag(al, 'OffsetRight') ? winW - p.x - w : p.x;
  const y = al && flag(al, 'OffsetBottom') ? winH - p.y - h : p.y;
  return { x, y, w, h };
}

// buildWindow builds the window with a WindowId at w by h (default: its
// own size). It returns {el, ctl, resize}: ctl maps each ControlId to its
// element, so the caller can set text and listen for clicks.
export function buildWindow(doc, data, id, opts = {}) {
  const wt = data.windows.get(id);
  if (!wt) throw new Error(`no window ${id} in the interface data`);
  const baseW = num(wt, 'Width'), baseH = num(wt, 'Height');
  const root = box(doc, 0, 0, baseW, baseH);
  root.dataset.window = id;
  const ctl = {};
  // Controls bound to an adapter, the stock name for a live value: bars take
  // a fraction, labels take text.
  const bars = new Map(), labels = new Map(), buttons = new Map();
  let compasses = [];
  const win = {
    el: root, ctl, width: baseW, height: baseH, resize,
    setBar(name, f) { (bars.get(name) || []).forEach((b) => b.set(f)); },
    setText(name, v, color) { (labels.get(name) || []).forEach((l) => l.set(v, color)); },
    has(name) { return bars.has(name) || labels.has(name); },
    button(event) { return buttons.get(event); },
    setHeading(deg) { compasses.forEach((c) => c.set(deg)); },
  };
  const bind = (map, name, item) => {
    if (!name) return;
    if (!map.has(name)) map.set(name, []);
    map.get(name).push(item);
  };

  function resize(w, h) {
    win.width = w;
    win.height = h;
    root.style.width = px(w);
    root.style.height = px(h);
    root.textContent = '';
    for (const k of Object.keys(ctl)) delete ctl[k];
    bars.clear();
    labels.clear();
    buttons.clear();
    compasses = [];
    populate();
  }

  function populate() {
    const w = win.width, h = win.height;
    for (const c of wt.c || []) {
      let el;
      const r = place(c, w, h, baseW, baseH);
      switch (c.t) {
        case 'FullResizeImageDef': {
          const tn = str(c, 'TemplateName') || opts.background;
          if (!tn) continue;
          el = frame(doc, data, tn, r.w, r.h);
          el.style.left = px(r.x);
          el.style.top = px(r.y);
          break;
        }
        case 'ChatControlDef': {
          const ct = data.tpl('ChatControlTemplate', str(c, 'TemplateName'));
          el = frame(doc, data, (ct && str(ct, 'BackgroundTemplate')) || '', r.w, r.h);
          el.style.left = px(r.x);
          el.style.top = px(r.y);
          el.dataset.font = ct ? str(ct, 'FontName') : '';
          const off = pt(c, 'TextOffset');
          el.dataset.textX = off ? off.x : 0;
          el.dataset.textY = off ? off.y : 0;
          el.dataset.linePadding = num(c, 'LinePadding');
          break;
        }
        case 'ButtonDef':
          el = button(doc, data, str(c, 'TemplateName'), r.x, r.y, opts.onClick && (() => opts.onClick(str(c, 'ControlId'))));
          break;
        case 'ImageAreaDef':
          el = sprite(doc, data, str(c, 'TemplateName') || str(c, 'ImageAreaTemplateName'), r.x, r.y);
          break;
        case 'LabelDef':
        case 'ScalarLabelDef': {
          const lt = labelText(doc, data, c);
          el = lt.el;
          el.style.left = px(r.x);
          el.style.top = px(r.y);
          bind(labels, str(c, 'Adapter') || str(c, 'AdapterName'), lt);
          break;
        }
        case 'InvisibleButtonDef':
          el = box(doc, r.x, r.y, r.w, r.h);
          break;
        case 'HorizontalResizeButtonDef': {
          const id = str(c, 'ControlId');
          const ev = str(c, 'OnClickEvent') || str(c, 'onClickEvent');
          const b = hrButton(doc, data, str(c, 'TemplateName'), str(c, 'Label'), r.w, opts.onClick && (() => opts.onClick(ev || id)));
          if (opts.disabled && opts.disabled(ev || id)) b.disable();
          el = b.el;
          el.style.left = px(r.x);
          el.style.top = px(r.y);
          if (ev) buttons.set(ev, b);
          break;
        }
        case 'CompassControlDef': {
          // The window names a small and a large control at the same place;
          // only the larger is drawn.
          const widest = Math.max(...(wt.c || []).filter((x) => x.t === 'CompassControlDef').map((x) => num(x, 'Width')));
          if (r.w < widest) continue;
          const cc = compass(doc, data, str(c, 'TemplateName'), r.w, r.h);
          el = cc.el;
          el.style.left = px(r.x);
          el.style.top = px(r.y);
          compasses.push(cc);
          break;
        }
        case 'StatusBarDef': {
          const sb = statusBar(doc, data, str(c, 'TemplateName'), r.w);
          el = sb.el;
          el.style.left = px(r.x);
          el.style.top = px(r.y);
          bind(bars, str(c, 'AdapterName'), sb);
          break;
        }
        default:
          continue;
      }
      const cid = str(c, 'ControlId');
      if (cid) {
        el.dataset.control = cid;
        ctl[cid] = el;
      }
      root.appendChild(el);
    }
  }

  populate();
  return win;
}

// loadUI fetches the converted layout of a skin and returns its data, or
// null when the converter has not been run so the caller can fall back.
export async function loadUI(base = 'data/ui/atlantis/', fetchFn = globalThis.fetch) {
  try {
    const res = await fetchFn(base + 'layout.json');
    if (!res.ok) return null;
    return new UIData(await res.json(), base);
  } catch {
    return null;
  }
}
