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

// createControls builds the drawers. h holds uiwin's primitives: box, paint,
// px, sprite, hrImage, hrButton, icon and the tree readers child, num, str, pt.
export function createControls(h) {
  const { box, paint, px, sprite, hrImage, hrButton, icon, child, num, str, pt } = h;
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

  return D;
}
