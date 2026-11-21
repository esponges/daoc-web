// The controls a stock window layout holds besides frames, labels and buttons:
// image strips, icons and (as they are added) lists, trees, checkboxes, tabs,
// sliders, combo boxes and text areas.
//
// uiwin.js owns the window builder and the DOM primitives. This file takes
// those primitives as a parameter (createControls(h)) rather than importing
// them, so the two do not import each other, and returns the drawer of each
// control type for the builder's registry. The layout arithmetic is pure and
// exported, checked in uictl.test.mjs.

// vrCells lays a VerticalResizeImageTemplate out at height h: the top and
// bottom ends keep their size and the middle repeats. Cells are
// {dy, dh, tex, sx, sy, sw, sh} with a common width.
export function vrCells(tpl, h, { child, num, str, pt }) {
  const tn = child(tpl, 'Texture');
  if (!tn) return { w: 0, cells: [] };
  const tex = str(tn, 'TextureName');
  const T = num(tpl, 'TopHeight'), M = num(tpl, 'RepeatHeight', 1) || 1, B = num(tpl, 'BottomHeight'), w = num(tpl, 'Width');
  const top = pt(tn, 'Top'), mid = pt(tn, 'Repeat'), bot = pt(tn, 'Bottom');
  const cells = [];
  const midH = Math.max(0, h - T - B);
  if (top && T > 0) cells.push({ dy: 0, dh: Math.min(T, h), tex, sx: top.x, sy: top.y, sw: w, sh: T });
  if (mid && midH > 0) cells.push({ dy: T, dh: midH, tex, sx: mid.x, sy: mid.y, sw: w, sh: M });
  if (bot && B > 0 && h >= T + B) cells.push({ dy: h - B, dh: B, tex, sx: bot.x, sy: bot.y, sw: w, sh: B });
  return { w, cells };
}

// --- lists ---

// clampTop keeps the first visible row in range: from 0 to the last position
// that still fills the view.
export function clampTop(top, total, view) {
  return Math.max(0, Math.min(Math.max(0, total - view), Math.round(top)));
}

// scrollThumb is the scrollbar thumb for a list: its height and its offset
// down a track of track pixels, for total rows with view of them visible and
// the first visible row at top. null when everything fits and there is no bar.
export function scrollThumb({ total, view, top, track, min = 16 }) {
  if (view <= 0 || total <= view) return null;
  const h = Math.min(track, Math.max(min, Math.round((track * view) / total)));
  const maxTop = total - view;
  return { h, y: Math.round((track - h) * (clampTop(top, total, view) / maxTop)), maxTop };
}

// topFromThumb is the first visible row for a thumb dragged to y on its track.
export function topFromThumb(y, track, h, total, view) {
  const room = Math.max(1, track - h);
  return clampTop((y / room) * (total - view), total, view);
}

// rowHeight is how tall a list row is: the taller of its text and its icon,
// with the template's line offset between rows.
export function rowHeight(fontH, iconH, lineOffset) {
  return Math.max(fontH, iconH) + lineOffset;
}

// createControls builds the drawers. h holds uiwin's primitives: box, paint,
// px, sprite, hrImage, hrButton, icon and the tree readers child, num, str, pt.
export function createControls(h) {
  const { box, paint, px, sprite, hrImage, hrButton, icon, button, text, frame, rgba, flag, child, num, str, pt } = h;
  const put = (el, r) => {
    el.style.left = px(r.x);
    el.style.top = px(r.y);
    return el;
  };
  const bind = (map, name, item) => {
    if (!name || name === 'none') return;
    if (!map.has(name)) map.set(name, []);
    map.get(name).push(item);
  };
  const D = {};

  // A strip of a horizontal template at the width the control names: the
  // dialog insets that sit behind labels and the dividers between sections.
  D.HorizontalResizeImageDef = (env, c, r) => put(hrImage(env.doc, env.data, str(c, 'TemplateName'), r.w), r);

  // The same turned upright: the gutters of the vault and lists.
  D.VerticalResizeImageDef = (env, c, r) => {
    const tpl = env.data.tpl('VerticalResizeImageTemplate', str(c, 'TemplateName'));
    const e = box(env.doc, r.x, r.y, 0, r.h);
    if (!tpl) return e;
    const { w, cells } = vrCells(tpl, r.h, h);
    e.style.width = px(w);
    for (const cell of cells) {
      const part = box(env.doc, 0, cell.dy, w, cell.dh);
      paint(part, env.data, { ...cell, dx: 0, dw: w });
      e.appendChild(part);
    }
    return e;
  };

  // A button that is a picture on a stretchable button: the chat window's
  // emote and settings buttons. The click goes to the window's onClick.
  D.HorizontalResizeImageButtonDef = (env, c, r) => {
    const id = str(c, 'ControlId');
    const ev = str(c, 'OnClickEvent') || str(c, 'onClickEvent');
    const w = r.w || 20;
    const b = hrButton(env.doc, env.data, str(c, 'HRButtonTemplateName'), '', w, env.opts.onClick && (() => env.opts.onClick(ev || id)));
    const art = env.data.tpl('ImageAreaTemplate', str(c, 'ImageAreaTemplateName'));
    if (art) {
      const size = pt(art, 'Size') || { x: 0, y: 0 };
      const pic = sprite(env.doc, env.data, str(c, 'ImageAreaTemplateName'), Math.max(0, (w - size.x) / 2), 0);
      pic.style.pointerEvents = 'none';
      b.el.appendChild(pic);
    }
    if (ev) env.buttons.set(ev, b);
    return put(b.el, r);
  };

  // A square for a picture that changes: a spec, a spell, an item. Its size is
  // the template's; win.setIcon(adapter, number) fills it from the icon sheets.
  D.IconDef = (env, c, r) => {
    const tpl = env.data.tpl('IconTemplate', str(c, 'TemplateName'));
    const size = (tpl && pt(tpl, 'Size')) || { x: r.w || 32, y: r.h || 32 };
    const holder = box(env.doc, r.x, r.y, size.x, size.y);
    const item = {
      el: holder,
      set(n) {
        holder.textContent = '';
        if (n > 0) holder.appendChild(icon(env.doc, env.data, n, 0, 0));
      },
    };
    const first = num(c, 'IconID', 0);
    if (first > 0) item.set(first);
    bind(env.icons, str(c, 'AdapterName'), item);
    return holder;
  };

  // A small marker that shows a level: the pips beside a group member's
  // endurance and power. It is drawn as its first cell, shown from level 1;
  // win.setStatus(adapter, level) shows or hides it.
  D.StatusIconDef = (env, c, r) => {
    const tpl = env.data.tpl('StatusIconTemplate', str(c, 'TemplateName'));
    const w = tpl ? num(tpl, 'Width') : 0, hgt = tpl ? num(tpl, 'Height') : 0;
    const e = box(env.doc, r.x, r.y, w, hgt);
    const start = tpl && pt(tpl, 'TextureStart');
    if (tpl && start) paint(e, env.data, { tex: str(tpl, 'TextureName'), sx: start.x, sy: start.y, sw: w, sh: hgt, dw: w, dh: hgt });
    e.style.display = 'none';
    bind(env.status, str(c, 'AdapterName'), { el: e, set(level) { e.style.display = level > 0 ? 'block' : 'none'; } });
    return e;
  };

  // A picture the game picks as it runs, such as a map: a rectangle of a texture.
  D.DynamicImageDef = (env, c, r) => {
    const at = pt(c, 'TextureCoords') || { x: 0, y: 0 };
    const dim = pt(c, 'Dimensions') || { x: r.w, y: r.h };
    const e = box(env.doc, r.x, r.y, dim.x, dim.y);
    paint(e, env.data, { tex: str(c, 'TextureName'), sx: at.x, sy: at.y, sw: dim.x, sh: dim.y, dw: dim.x, dh: dim.y });
    return e;
  };

  // vrImage builds a VerticalResizeImageTemplate at a height.
  const vrImage = (env, name, height) => {
    const tpl = env.data.tpl('VerticalResizeImageTemplate', name);
    const e = box(env.doc, 0, 0, 0, height);
    if (!tpl) return e;
    const { w, cells } = vrCells(tpl, height, h);
    e.style.width = px(w);
    for (const cell of cells) {
      const part = box(env.doc, 0, cell.dy, w, cell.dh);
      paint(part, env.data, { ...cell, dx: 0, dw: w });
      e.appendChild(part);
    }
    return e;
  };

  // scrollBar builds a VerticalScrollbarTemplate height pixels tall: the gutter
  // with the up and down buttons on its ends and a thumb between them.
  // set({total, view, top}) moves the thumb; onScroll(top) hears the arrows,
  // the gutter and a dragged thumb. Returns null for a template that is absent.
  const scrollBar = (env, tplName, height, onScroll) => {
    const tpl = env.data.tpl('VerticalScrollbarTemplate', tplName);
    if (!tpl) return null;
    const doc = env.doc;
    const w = num(tpl, 'Width', 15);
    const root = box(doc, 0, 0, w, height);
    root.appendChild(vrImage(env, str(tpl, 'GutterVResizeTemplate'), height));
    const upT = env.data.tpl('ButtonTemplate', str(tpl, 'UpButtonTemplate'));
    const downT = env.data.tpl('ButtonTemplate', str(tpl, 'DownButtonTemplate'));
    const upH = upT && pt(upT, 'Size') ? pt(upT, 'Size').y : 0;
    const downH = downT && pt(downT, 'Size') ? pt(downT, 'Size').y : 0;
    let state = { total: 0, view: 0, top: 0 };
    const step = (d) => () => onScroll(clampTop(state.top + d, state.total, state.view));
    root.appendChild(button(doc, env.data, str(tpl, 'UpButtonTemplate'), 0, 0, step(-1)));
    root.appendChild(button(doc, env.data, str(tpl, 'DownButtonTemplate'), 0, height - downH, step(1)));
    const track = Math.max(1, height - upH - downH);
    const thumbOff = child(tpl, 'ThumbOffset');
    const thumb = box(doc, thumbOff ? num(thumbOff, 'X', 2) : 2, upH, 0, 0);
    thumb.style.cursor = 'pointer';
    root.appendChild(thumb);
    let grab = null;
    thumb.addEventListener('pointerdown', (e) => { grab = { y: e.clientY, top: state.top }; });
    thumb.addEventListener('pointermove', (e) => {
      if (!grab) return;
      const t = scrollThumb({ ...state, track });
      if (t) onScroll(topFromThumb((grab.top / Math.max(1, t.maxTop)) * (track - t.h) + (e.clientY - grab.y), track, t.h, state.total, state.view));
    });
    thumb.addEventListener('pointerup', () => { grab = null; });
    // A click on the gutter above or below the thumb pages that way.
    root.addEventListener('pointerdown', (e) => {
      if (e.target !== root) return;
      const t = scrollThumb({ ...state, track });
      if (t) onScroll(clampTop(state.top + (e.offsetY < upH + t.y ? -state.view : state.view), state.total, state.view));
    });
    return {
      el: root, w,
      set(next) {
        state = { total: next.total, view: next.view, top: next.top };
        const t = scrollThumb({ ...state, track });
        thumb.textContent = '';
        thumb.style.display = t ? 'block' : 'none';
        if (t) {
          thumb.style.top = px(upH + t.y);
          thumb.style.height = px(t.h);
          thumb.appendChild(vrImage(env, str(tpl, 'ThumbVResizeTemplate'), t.h));
        }
      },
    };
  };

  // A list: rows of text with an optional icon, a scrollbar when they overflow,
  // the wheel, a selected row, and rows that can be dragged. Rows are
  // {text, color, icon, title, drag, header, on}: header rows are gold, rows
  // with on false are greyed. win.list(adapter or control id) gives
  //   setRows(rows)  select(i)  selected  onSelect(fn(i, row))  onActivate(fn)
  //   top  scrollTo(i)  rows
  D.ListBoxDef = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('ListBoxTemplate', str(c, 'TemplateName'));
    const root = box(doc, r.x, r.y, r.w, r.h);
    if (tpl) root.appendChild(frame(doc, data, str(tpl, 'BackgroundTemplate'), r.w, r.h));
    const font = tpl && child(tpl, 'TextFont');
    const fontName = (font && str(font, 'Name')) || 'arial9';
    const colors = {
      normal: rgba(child(font, 'ColorNormal')) || 'rgb(225,225,225)',
      high: rgba(child(font, 'ColorHighlit')) || 'rgb(255,192,0)',
      grey: rgba(child(font, 'ColorGreyed')) || 'rgb(127,127,127)',
      head: 'rgb(255,200,100)',
    };
    const bf = data.bitmapFont(fontName);
    const fontH = (bf && bf.h) || 12;
    const iconSize = tpl && child(tpl, 'IconSize') ? pt(tpl, 'IconSize') : null;
    const iconPx = iconSize ? iconSize.x : 16;
    const rowH = rowHeight(fontH, iconPx, num(font, 'LineOffset', 2));
    const pad = 4, leftText = num(font, 'TextLeftOffset', 4);
    const innerW = r.w - pad * 2;
    const inner = box(doc, pad, pad, innerW, r.h - pad * 2);
    inner.style.overflow = 'hidden';
    root.appendChild(inner);
    const view = Math.max(1, Math.floor((r.h - pad * 2) / rowH));
    let rows = [], top = 0, selected = -1, onSelect = null, onActivate = null, bar = null;
    const draw = () => {
      inner.textContent = '';
      rows.slice(top, top + view).forEach((row, k) => {
        const i = top + k;
        const line = box(doc, 0, k * rowH, innerW, rowH);
        if (i === selected) line.style.background = 'rgba(255,192,0,0.2)';
        line.dataset.row = String(i);
        let x = leftText;
        if (row.icon) {
          const ic = icon(doc, data, row.icon, x, Math.max(0, (rowH - iconPx) / 2));
          ic.style.transform = 'scale(' + iconPx / 32 + ')';
          ic.style.transformOrigin = '0 0';
          line.appendChild(ic);
          x += iconPx + 4;
        }
        const color = row.header ? colors.head : row.on === false ? colors.grey : i === selected ? colors.high : (row.color || colors.normal);
        const t = text(doc, data, fontName, row.text, color);
        t.style.left = px(x);
        t.style.top = px(Math.max(0, (rowH - fontH) / 2));
        line.appendChild(t);
        if (row.title) line.title = row.title;
        if (row.drag) {
          line.draggable = true;
          line.addEventListener('dragstart', (e) => { if (e.dataTransfer) e.dataTransfer.setData('text/plain', row.drag); });
        }
        line.style.cursor = row.header ? 'default' : 'pointer';
        line.addEventListener('click', () => { if (!row.header) api.select(i); });
        line.addEventListener('dblclick', () => { if (!row.header && onActivate) onActivate(i, row); });
        inner.appendChild(line);
      });
      if (bar) bar.set({ total: rows.length, view, top });
    };
    const api = {
      el: root,
      get rows() { return rows; },
      get top() { return top; },
      get selected() { return selected; },
      get view() { return view; },
      setRows(next) {
        rows = next || [];
        top = clampTop(top, rows.length, view);
        if (selected >= rows.length) selected = -1;
        draw();
      },
      select(i) {
        selected = i;
        draw();
        if (onSelect && rows[i]) onSelect(i, rows[i]);
      },
      scrollTo(i) { top = clampTop(i, rows.length, view); draw(); },
      onSelect(fn) { onSelect = fn; },
      onActivate(fn) { onActivate = fn; },
    };
    if (tpl && str(tpl, 'DrawScroll') !== 'false') {
      bar = scrollBar(env, str(tpl, 'ScrollbarTemplate'), r.h - 4, (t) => api.scrollTo(t));
      if (bar) {
        bar.el.style.left = px(r.w - bar.w - 3);
        bar.el.style.top = px(2);
        root.appendChild(bar.el);
      }
    }
    root.addEventListener('wheel', (e) => {
      if (rows.length <= view) return;
      if (e.preventDefault) e.preventDefault();
      api.scrollTo(top + (e.deltaY > 0 ? 3 : -3));
    });
    bind(env.lists, str(c, 'AdapterName'), api);
    bind(env.lists, str(c, 'ControlId'), api);
    return root;
  };

  return D;
}
