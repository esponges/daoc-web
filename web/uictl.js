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

// flattenGroups turns the groups of a tree into the rows of its list: a heading
// for each group, folded with a box that says open or closed, then the group's
// own rows, indented, when it is open. A heading remembers its group.
export function flattenGroups(groups) {
  const rows = [];
  groups.forEach((g, i) => {
    rows.push({ text: g.title, header: true, check: g.open ? 'open' : 'closed', group: i, title: g.titleTip });
    if (g.open) for (const row of g.rows || []) rows.push({ indent: 12, ...row });
  });
  return rows;
}

// wrapText breaks text into lines no wider than width, where measure(string)
// gives a string's width in pixels. Lines break at spaces and at newlines;
// a word wider than the line is broken wherever it fits.
export function wrapText(measure, text, width) {
  const out = [];
  for (const para of String(text ?? '').split(/\r?\n/)) {
    let line = '';
    for (const word of para.split(' ')) {
      const trial = line ? line + ' ' + word : word;
      if (measure(trial) <= width || !line && !word) { line = trial; continue; }
      if (line) { out.push(line); line = ''; }
      let w = word;
      while (measure(w) > width && w.length > 1) {
        let k = w.length - 1;
        while (k > 1 && measure(w.slice(0, k)) > width) k--;
        out.push(w.slice(0, k));
        w = w.slice(k);
      }
      line = w;
    }
    out.push(line);
  }
  return out;
}

// sliderValue is the value a slider holds when the pointer is x pixels along
// its track: ticks values, 0 at the left, ticks - 1 at the right.
export function sliderValue(x, track, ticks) {
  if (ticks <= 1 || track <= 0) return 0;
  return Math.max(0, Math.min(ticks - 1, Math.round((x / track) * (ticks - 1))));
}

// sliderX is where the indicator sits for a value, in pixels along the track.
export function sliderX(value, track, ticks) {
  if (ticks <= 1) return 0;
  return Math.round((Math.max(0, Math.min(ticks - 1, value)) / (ticks - 1)) * track);
}

// createControls builds the drawers. h holds uiwin's primitives: box, paint,
// px, sprite, hrImage, hrButton, icon and the tree readers child, num, str, pt.
export function createControls(h) {
  const { box, paint, px, sprite, hrImage, hrButton, icon, button, text, layoutText, frame, rgba, flag, child, num, str, pt } = h;
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

  // buildList builds the rows, scrollbar and wheel of a list. cfg says how it
  // looks: {bg, font: {name, normal, high, grey, head}, lineOffset, iconPx,
  // textLeft, scroll: scrollbar template name or '', check: art for a heading's
  // box, onHeader(i, row)}. Rows are {text, color, icon, title, drag, header,
  // on, indent, check} where check is 'open' or 'closed' for a heading that
  // folds. It returns {root, api}.
  const buildList = (env, r, cfg) => {
    const { doc, data } = env;
    const root = box(doc, r.x, r.y, r.w, r.h);
    if (cfg.bg) root.appendChild(frame(doc, data, cfg.bg, r.w, r.h));
    const fontName = cfg.font.name;
    const bf = data.bitmapFont(fontName);
    const fontH = (bf && bf.h) || 12;
    const iconPx = cfg.iconPx;
    const rowH = rowHeight(fontH, iconPx, cfg.lineOffset);
    const pad = 4;
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
        let x = cfg.textLeft + (row.indent || 0);
        if (row.check && cfg.check) {
          const cb = box(doc, x, Math.max(0, (rowH - cfg.check.size) / 2), cfg.check.size, cfg.check.size);
          const at = row.check === 'open' ? cfg.check.pressed : cfg.check.normal;
          if (at) paint(cb, data, { tex: cfg.check.tex, sx: at.x, sy: at.y, sw: cfg.check.size, sh: cfg.check.size, dw: cfg.check.size, dh: cfg.check.size });
          line.appendChild(cb);
          x += cfg.check.size + 5;
        }
        if (row.icon) {
          const ic = icon(doc, data, row.icon, x, Math.max(0, (rowH - iconPx) / 2));
          ic.style.transform = 'scale(' + iconPx / 32 + ')';
          ic.style.transformOrigin = '0 0';
          line.appendChild(ic);
          x += iconPx + 4;
        }
        const color = row.header ? cfg.font.head : row.on === false ? cfg.font.grey : i === selected ? cfg.font.high : (row.color || cfg.font.normal);
        const t = text(doc, data, fontName, row.text, color);
        t.style.left = px(x);
        t.style.top = px(Math.max(0, (rowH - fontH) / 2));
        line.appendChild(t);
        if (row.title) line.title = row.title;
        if (row.drag) {
          line.draggable = true;
          line.addEventListener('dragstart', (e) => { if (e.dataTransfer) e.dataTransfer.setData('text/plain', row.drag); });
        }
        line.style.cursor = row.header && !row.check ? 'default' : 'pointer';
        line.addEventListener('click', () => {
          if (row.check && cfg.onHeader) { cfg.onHeader(i, row); return; }
          if (!row.header) api.select(i);
        });
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
    if (cfg.scroll) {
      bar = scrollBar(env, cfg.scroll, r.h - 4, (t) => api.scrollTo(t));
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
    return { root, api };
  };

  // A list: rows of text with an optional icon, a scrollbar when they overflow,
  // the wheel, a selected row, and rows that can be dragged. win.list(adapter
  // or control id) gives
  //   setRows(rows)  select(i)  selected  onSelect(fn(i, row))  onActivate(fn)
  //   top  scrollTo(i)  rows
  D.ListBoxDef = (env, c, r) => {
    const { data } = env;
    const tpl = data.tpl('ListBoxTemplate', str(c, 'TemplateName'));
    const font = tpl && child(tpl, 'TextFont');
    const iconSize = tpl && child(tpl, 'IconSize') ? pt(tpl, 'IconSize') : null;
    const { root, api } = buildList(env, r, {
      bg: tpl ? str(tpl, 'BackgroundTemplate') : '',
      font: {
        name: (font && str(font, 'Name')) || 'arial9',
        normal: rgba(child(font, 'ColorNormal')) || 'rgb(225,225,225)',
        high: rgba(child(font, 'ColorHighlit')) || 'rgb(255,192,0)',
        grey: rgba(child(font, 'ColorGreyed')) || 'rgb(127,127,127)',
        head: 'rgb(255,200,100)',
      },
      lineOffset: num(font, 'LineOffset', 2),
      iconPx: iconSize ? iconSize.x : 16,
      textLeft: num(font, 'TextLeftOffset', 4),
      scroll: tpl && str(tpl, 'DrawScroll') !== 'false' ? str(tpl, 'ScrollbarTemplate') : '',
    });
    bind(env.lists, str(c, 'AdapterName'), api);
    bind(env.lists, str(c, 'ControlId'), api);
    return root;
  };

  // A tree: headings that fold with a small box, and rows under them. It is a
  // list of the open groups' rows; win.tree(adapter or control id) gives the
  // list's calls and
  //   setGroups([{title, open, rows}])  toggle(group)  isOpen(group)
  // where a group's rows are list rows and open says whether they show.
  D.TreeControlDef = (env, c, r) => {
    const { data } = env;
    const tpl = data.tpl('TreeControlTemplate', str(c, 'TemplateName'));
    const checkTpl = tpl && data.tpl('CheckBoxTemplate', str(tpl, 'CheckBoxTemplate'));
    const tex = checkTpl && child(checkTpl, 'Texture');
    const checkSize = checkTpl && pt(checkTpl, 'Size') ? pt(checkTpl, 'Size').x : 12;
    const colorOf = tpl && child(tpl, 'FontColor');
    const iconTpl = tpl && data.tpl('IconTemplate', str(tpl, 'IconTemplate'));
    let groups = [];
    let list = null;
    const refresh = () => list.setRows(flattenGroups(groups));
    const made = buildList(env, r, {
      bg: tpl && str(tpl, 'BackgroundTemplate') !== 'none' ? str(tpl, 'BackgroundTemplate') : '',
      font: {
        name: (tpl && str(tpl, 'FontName')) || 'arial9',
        normal: rgba(colorOf) || 'rgb(192,192,192)',
        high: 'rgb(255,192,0)',
        grey: 'rgb(127,127,127)',
        head: 'rgb(230,200,120)',
      },
      lineOffset: tpl ? num(tpl, 'LinePadding', 1) + 1 : 2,
      iconPx: iconTpl && pt(iconTpl, 'Size') ? pt(iconTpl, 'Size').x : 16,
      textLeft: tpl ? num(tpl, 'OffsetX', 0) + 2 : 2,
      scroll: tpl ? str(tpl, 'ScrollbarTemplate') : '',
      check: tex ? { tex: str(tex, 'TextureName'), normal: pt(tex, 'Normal'), pressed: pt(tex, 'Pressed'), size: checkSize } : null,
      onHeader: (i, row) => { groups[row.group].open = !groups[row.group].open; refresh(); },
    });
    list = made.api;
    const api = Object.assign(Object.create(list), {
      setGroups(next) { groups = (next || []).map((g) => ({ ...g })); refresh(); },
      toggle(g) { if (groups[g]) { groups[g].open = !groups[g].open; refresh(); } },
      isOpen(g) { return !!(groups[g] && groups[g].open); },
      get groups() { return groups; },
    });
    bind(env.lists, str(c, 'AdapterName'), api);
    bind(env.lists, str(c, 'ControlId'), api);
    return made.root;
  };

  // Some layouts spell the adapter name Adaptername.
  const adapterOf = (c) => str(c, 'AdapterName') || str(c, 'Adaptername');
  const bindBoth = (map, c, item) => {
    bind(map, adapterOf(c), item);
    bind(map, str(c, 'ControlId'), item);
  };
  const fontColors = (font) => ({
    normal: rgba(child(font, 'ColorNormal')) || 'rgb(192,192,192)',
    pressed: rgba(child(font, 'ColorPressed')) || 'rgb(255,192,0)',
    high: rgba(child(font, 'ColorHighlit')) || 'rgb(255,255,255)',
    off: rgba(child(font, 'ColorDisabled')) || 'rgb(128,128,128)',
  });

  // A check box: the skin's little box and a caption in the template's font.
  // win.check(adapter or control id) gives set(on), get(), onChange(fn(on)),
  // setLabel(text) and disable(on). A click toggles it and reports the new state.
  D.CheckBoxDef = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('CheckBoxTemplate', str(c, 'TemplateName'));
    const size = tpl && pt(tpl, 'Size') ? pt(tpl, 'Size').x : 12;
    const font = tpl && child(tpl, 'Font');
    const colors = fontColors(font);
    const fontName = (font && str(font, 'Name')) || 'arial9';
    const textAt = font && child(font, 'TextOffset') ? pt(font, 'TextOffset').x : size + 5;
    const tex = tpl && child(tpl, 'Texture');
    const labelW = num(c, 'LabelWidth', 60);
    const root = box(doc, r.x, r.y, size + 5 + labelW, size);
    root.style.cursor = 'pointer';
    let on = false, over = false, off = false, label = str(c, 'Data'), listener = null;
    const art = box(doc, 0, 0, size, size);
    const cap = box(doc, textAt, 0, labelW, size);
    root.append(art, cap);
    const draw = () => {
      const state = off ? 'Disabled' : on ? (over ? 'PressedHighlit' : 'Pressed') : over ? 'NormalHighlit' : 'Normal';
      const at = tex && (pt(tex, state) || pt(tex, 'Normal'));
      if (at) paint(art, data, { tex: str(tex, 'TextureName'), sx: at.x, sy: at.y, sw: size, sh: size, dw: size, dh: size });
      cap.textContent = '';
      if (label) {
        const t = text(doc, data, fontName, label, off ? colors.off : on ? colors.pressed : over ? colors.high : colors.normal);
        const bf = data.bitmapFont(fontName);
        t.style.top = px(Math.max(0, (size - ((bf && bf.h) || 11)) / 2));
        cap.appendChild(t);
      }
    };
    const api = {
      el: root,
      set(v) { on = !!v; draw(); },
      get() { return on; },
      onChange(fn) { listener = fn; },
      setLabel(t) { label = t; draw(); },
      disable(v = true) { off = v; root.style.cursor = v ? 'default' : 'pointer'; draw(); },
    };
    root.addEventListener('pointerenter', () => { over = true; draw(); });
    root.addEventListener('pointerleave', () => { over = false; draw(); });
    root.addEventListener('click', () => {
      if (off) return;
      on = !on;
      draw();
      if (listener) listener(on);
    });
    draw();
    bindBoth(env.checks, c, api);
    return root;
  };

  // A block of wrapped text with a scrollbar: what a trainer or healer says,
  // a quest's story. win.area(adapter or control id).setText(text) fills it.
  D.TextAreaDef = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('TextAreaTemplate', str(c, 'TemplateName'));
    const font = tpl && child(tpl, 'Font');
    const colors = { normal: rgba(child(font, 'ColorNormal')) || 'rgb(192,192,192)' };
    const fontName = (font && str(font, 'Name')) || 'arial11';
    const bf = data.bitmapFont(fontName);
    const fontH = (bf && bf.h) || 12;
    const lineH = fontH + 2 + (tpl ? num(tpl, 'LinePadding', 0) : 0);
    const off = tpl && child(tpl, 'TextOffset') ? pt(tpl, 'TextOffset') : { x: 8, y: 4 };
    const root = box(doc, r.x, r.y, r.w, r.h);
    if (tpl && str(tpl, 'BackgroundTemplate')) root.appendChild(frame(doc, data, str(tpl, 'BackgroundTemplate'), r.w, r.h));
    const view = Math.max(1, Math.floor((r.h - off.y * 2) / lineH));
    const inner = box(doc, off.x, off.y, r.w - off.x * 2 - 14, r.h - off.y * 2);
    inner.style.overflow = 'hidden';
    root.appendChild(inner);
    const measure = (t) => (bf ? layoutText(bf, t).width : t.length * 6);
    let lines = [], top = 0, bar = null;
    const draw = () => {
      inner.textContent = '';
      lines.slice(top, top + view).forEach((ln, k) => {
        if (!ln) return;
        const t = text(doc, data, fontName, ln, colors.normal);
        t.style.top = px(k * lineH);
        inner.appendChild(t);
      });
      if (bar) bar.set({ total: lines.length, view, top });
    };
    const api = {
      el: root,
      get lines() { return lines; },
      get top() { return top; },
      get view() { return view; },
      setText(t) {
        lines = wrapText(measure, t, r.w - off.x * 2 - 14);
        top = 0;
        draw();
      },
      scrollTo(i) { top = clampTop(i, lines.length, view); draw(); },
    };
    if (tpl && str(tpl, 'ScrollbarTemplate')) {
      bar = scrollBar(env, str(tpl, 'ScrollbarTemplate'), r.h - 4, (t) => api.scrollTo(t));
      if (bar) {
        bar.el.style.left = px(r.w - bar.w - 3);
        bar.el.style.top = px(2);
        root.appendChild(bar.el);
      }
    }
    root.addEventListener('wheel', (e) => {
      if (lines.length <= view) return;
      if (e.preventDefault) e.preventDefault();
      api.scrollTo(top + (e.deltaY > 0 ? 3 : -3));
    });
    bindBoth(env.areas, c, api);
    return root;
  };

  // Tabs: a row of tab buttons over a framed area, and the controls each tab
  // shows (the layout's TabControl entries by ControlId). win.tabs(id) gives
  // select(tab id), selected and onChange(fn(id)).
  D.TabsDef = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('TabsTemplate', str(c, 'TemplateName'));
    const btnTpl = tpl && data.tpl('ButtonTemplate', str(tpl, 'TabButtonTemplate'));
    const bw = btnTpl && pt(btnTpl, 'Size') ? pt(btnTpl, 'Size').x : 100;
    const bh = btnTpl && pt(btnTpl, 'Size') ? pt(btnTpl, 'Size').y : 16;
    const root = box(doc, r.x, r.y, r.w, r.h);
    if (tpl && str(tpl, 'ControlAreaFullResizeImageTemplate')) {
      const area = frame(doc, data, str(tpl, 'ControlAreaFullResizeImageTemplate'), r.w, r.h - bh);
      area.style.top = px(bh);
      root.appendChild(area);
    }
    const tabs = (c.c || []).filter((x) => x.t === 'Tab').map((x) => ({ id: str(x, 'Id'), name: str(x, 'Name') }));
    const owned = (c.c || []).filter((x) => x.t === 'TabControl').map((x) => ({ tab: str(x, 'TabId'), id: str(x, 'ControlId') }));
    let selected = null, listener = null;
    const labels = new Map();
    const api = {
      el: root,
      get tabs() { return tabs; },
      get selected() { return selected; },
      select(id) {
        selected = id;
        for (const [tid, b] of labels) b.look(tid === id ? 'Pressed' : 'Normal');
        for (const o of owned) {
          const el = env.win.ctl[o.id];
          if (el) el.style.display = o.tab === id ? '' : 'none';
        }
        if (listener) listener(id);
      },
      onChange(fn) { listener = fn; },
    };
    tabs.forEach((t, i) => {
      const b = button(doc, data, str(tpl || c, 'TabButtonTemplate'), (tpl ? num(tpl, 'TabXOffset', 5) : 5) + i * (bw + 2), 0, () => api.select(t.id), t.name);
      labels.set(t.id, b);
      root.appendChild(b);
    });
    // Once every control of the window is drawn, show the first tab's.
    env.after.push(() => { if (tabs.length) api.select(tabs[0].id); });
    bindBoth(env.tabsets, c, api);
    return root;
  };

  // A slider: the track, an indicator, and ticks values. win.slider(adapter or
  // control id) gives set(v), get(), onChange(fn(v)). Dragging along the track
  // or clicking on it sets the value.
  D.HorizontalSliderDef = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('HorizontalSliderTemplate', str(c, 'TemplateName'));
    const w = r.w || num(c, 'Width', 100);
    const trackH = tpl ? num(tpl, 'Height', 13) : 13;
    const root = box(doc, r.x, r.y, w, trackH);
    root.style.cursor = 'pointer';
    if (tpl) root.appendChild(hrImage(doc, data, str(tpl, 'BackgroundHResizeTemplate'), w));
    const art = tpl && data.tpl('ImageAreaTemplate', str(tpl, 'ImageAreaTemplate'));
    const iw = art && pt(art, 'Size') ? pt(art, 'Size').x : 13;
    const ih = art && pt(art, 'Size') ? pt(art, 'Size').y : 7;
    const ticks = num(c, 'Numticks', num(c, 'NumTicks', 2));
    const track = Math.max(1, w - iw);
    const mark = tpl ? sprite(doc, data, str(tpl, 'ImageAreaTemplate'), 0, Math.max(0, (trackH - ih) / 2)) : box(doc, 0, 0, iw, ih);
    root.appendChild(mark);
    let value = 0, listener = null, dragging = false;
    const draw = () => { mark.style.left = px(sliderX(value, track, ticks)); };
    const setFrom = (e) => {
      const rect = root.getBoundingClientRect ? root.getBoundingClientRect() : { left: 0 };
      const v = sliderValue((e.clientX - rect.left) - iw / 2, track, ticks);
      if (v !== value) { value = v; draw(); if (listener) listener(value); }
    };
    root.addEventListener('pointerdown', (e) => { dragging = true; setFrom(e); });
    root.addEventListener('pointermove', (e) => { if (dragging) setFrom(e); });
    root.addEventListener('pointerup', () => { dragging = false; });
    const api = {
      el: root, ticks,
      set(v) { value = Math.max(0, Math.min(ticks - 1, v)); draw(); },
      get() { return value; },
      onChange(fn) { listener = fn; },
    };
    draw();
    bindBoth(env.sliders, c, api);
    return root;
  };

  // A combo box: the chosen option on a stock button, and the down arrow.
  // Clicking steps to the next option. win.combo(id) gives setOptions(list),
  // select(i), selected, value and onChange(fn(i, option)).
  D.ComboBoxDef = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('ComboBoxTemplate', str(c, 'Templatename') || str(c, 'TemplateName'));
    const root = box(doc, r.x, r.y, 120, 18);
    let options = [], at = -1, listener = null, face = null;
    const down = tpl ? button(doc, data, str(tpl, 'DownButtonTemplate'), 104, 1, () => api.select((at + 1) % Math.max(1, options.length))) : null;
    const draw = () => {
      if (face) face.remove ? face.remove() : null;
      root.textContent = '';
      face = box(doc, 0, 0, 104, 18);
      face.style.cursor = 'pointer';
      face.addEventListener('click', () => api.select((at + 1) % Math.max(1, options.length)));
      if (at >= 0 && options[at] !== undefined) face.appendChild(text(doc, data, 'arial9', String(options[at]), 'rgb(220,220,220)'));
      root.appendChild(face);
      if (down) root.appendChild(down);
    };
    const api = {
      el: root,
      get selected() { return at; },
      get value() { return options[at]; },
      setOptions(list) { options = list || []; at = options.length ? Math.min(Math.max(at, 0), options.length - 1) : -1; draw(); },
      select(i) { if (!options.length) return; at = i; draw(); if (listener) listener(i, options[i]); },
      onChange(fn) { listener = fn; },
    };
    draw();
    bindBoth(env.combos, c, api);
    return root;
  };

  // A line to type in: a text box on the entry bar's art. win.edit(id) gives
  // value, set(text), focus() and onEnter(fn(text)).
  const edit = (env, c, r) => {
    const { doc, data } = env;
    const tpl = data.tpl('EditBoxTemplate', str(c, 'TemplateName'));
    const w = r.w || num(c, 'Width', 150), hgt = r.h || num(c, 'Height', 16);
    const root = box(doc, r.x, r.y, w, hgt);
    const bg = tpl && str(tpl, 'BackgroundTemplate');
    if (bg && bg !== 'none') root.appendChild(frame(doc, data, bg, w, hgt));
    const input = doc.createElement('input');
    input.type = 'text';
    input.autocomplete = 'off';
    Object.assign(input.style, { position: 'absolute', left: '6px', top: '0px', width: px(w - 12), height: px(hgt), border: '0', outline: '0', background: 'transparent', color: '#fff', font: '12px sans-serif' });
    root.appendChild(input);
    let onEnter = null;
    input.addEventListener('keydown', (e) => { if (e.key === 'Enter' && onEnter) onEnter(input.value); });
    const api = {
      el: root, input,
      get value() { return input.value; },
      set(t) { input.value = t; },
      focus() { if (input.focus) input.focus(); },
      onEnter(fn) { onEnter = fn; },
    };
    bindBoth(env.edits, c, api);
    return root;
  };
  D.EditBoxDef = edit;
  D.ClickableEditBoxDef = edit;

  return D;
}
