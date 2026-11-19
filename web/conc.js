// Concentration: the buffs a caster holds, and the pool they draw on.
//
// The server sends `conc` whenever it changes: the pool (`max`), what the
// buffs hold (`used`) and each buff (`id`, `nm`, `tg`, `cost`, `sp`). The
// window only shows it and asks to cancel one (`conc-cancel`): the server
// decides. The converted interface has no concentration window, so the stock
// version is the stock dialog frame with the list inside; the fallback is a
// plain panel.

import { frame } from './uiwin.js';

// concRows turns the conc message into what to draw.
export function concRows(m) {
  const list = (m && m.list) || [];
  return {
    title: 'Concentration ' + ((m && m.used) || 0) + ' / ' + ((m && m.max) || 0),
    rows: list.map((b) => ({ id: b.id, text: b.nm + ' on ' + b.tg, cost: b.cost, spell: b.sp || 0 })),
    empty: list.length === 0 ? 'You hold no buffs.' : '',
  };
}

// heldSpells is the set of spell ids a caster holds buffs of, for marking the
// spell bar and the spell list.
export function heldSpells(m) {
  return new Set(((m && m.list) || []).map((b) => b.sp).filter(Boolean));
}

// createConcWindow builds the list: on the stock dialog frame when the
// interface is there (data), a plain panel when not. update(m) redraws it,
// toggle() opens or closes it. onCancel(id) asks the server to end a buff.
export function createConcWindow(doc, data, { onCancel } = {}) {
  const root = doc.createElement('div');
  root.className = data ? '' : 'panel conc';
  const W = 280;
  let open = false, last = null;
  const show = () => {
    if (data) root.style.display = open ? 'block' : 'none';
    else if (open) root.classList.add('open'); else root.classList.remove('open');
  };
  const line = (text, cls) => {
    const d = doc.createElement('div');
    if (cls) d.className = cls;
    d.textContent = text;
    return d;
  };
  const render = () => {
    const v = concRows(last);
    const body = doc.createElement('div');
    body.append(line(v.title, 'conc-title'));
    if (v.empty) body.append(line(v.empty, 'conc-empty'));
    for (const r of v.rows) {
      const row = doc.createElement('div');
      row.className = 'conc-row';
      const label = line(r.text + '  (' + r.cost + ')');
      const x = doc.createElement('button');
      x.textContent = 'Cancel';
      x.addEventListener('click', () => onCancel?.(r.id));
      row.append(label, x);
      body.append(row);
    }
    root.replaceChildren();
    if (data) {
      const H = 44 + 24 * Math.max(1, v.rows.length);
      Object.assign(root.style, { position: 'fixed', width: W + 'px', height: H + 'px' });
      Object.assign(body.style, { position: 'absolute', left: '12px', top: '10px', width: W - 24 + 'px', font: '12px sans-serif', color: '#fff', textShadow: '0 1px 1px #000' });
      root.append(frame(doc, data, 'dlg_sm_title_noresize', W, H));
    }
    root.append(body);
    show();
  };
  render();
  return {
    el: root,
    get isOpen() { return open; },
    update(m) { last = m; render(); },
    toggle(on = !open) { open = on; show(); },
  };
}
