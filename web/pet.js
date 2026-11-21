// The pet window: the owner's pet, its health, stance and orders.
//
// The server sends `pet` whenever the pet or its state changes, and with
// `cl` true when it is gone. The window only shows it and sends `pet-cmd`:
// the server decides what the pet does. The converted interface has no pet
// window, so the stock version is the stock dialog frame with the controls
// inside; the fallback is a plain panel.

import { frame } from './uiwin.js';
import { effectRows } from './effects.js';

// The most effects the window shows, as the stock pet window has room for.
export const MAX_PET_EFFECTS = 6;

// petEffects picks what to show of the pet's lasting effects at a time: the
// six with the least time left, an effect with no timer after those.
export function petEffects(m, now) {
  const list = (m && m.fx) || [];
  const rows = effectRows({ at: (m && m.at) || 0, list }, now).map((r) => ({ ...r, timed: r.of > 0, text: r.of > 0 ? r.text : '' }));
  rows.sort((a, b) => (a.timed === b.timed ? a.left - b.left : a.timed ? -1 : 1));
  return rows.slice(0, MAX_PET_EFFECTS);
}

export const STANCES = ['passive', 'defensive', 'aggressive'];
export const ORDERS = [['attack', 'Attack'], ['follow', 'Follow'], ['stay', 'Stay'], ['release', 'Release']];

// petView turns the pet message into what to draw; null when there is no pet.
export function petView(m, now = 0) {
  if (!m || m.cl || !m.id) return null;
  const pct = Math.max(0, Math.min(1, m.hp ?? 0));
  return {
    title: m.nm + '  (level ' + (m.lv || 0) + ')',
    hp: pct,
    hpText: Math.round(pct * 100) + '%',
    stance: m.st || 'defensive',
    mode: m.md || 'follow',
    effects: petEffects(m, now),
    minions: (m.mn || []).map((x) => ({ id: x.id, text: x.nm + ' (' + (x.lv || 0) + ')  ' + Math.round(Math.max(0, Math.min(1, x.hp)) * 100) + '%' })),
  };
}

// petOrderText is the chat line that gives an order, for /pet and the buttons.
export function petOrderText(cmd) { return '/pet ' + cmd; }

// ownerLabel is the line drawn under a pet's name: whose it is.
export function ownerLabel(ent) { return ent && ent.ow ? ent.ow + "'s pet" : ''; }

// createPetWindow builds the window: on the stock dialog frame when the
// interface is there (data), a plain panel when not. update(m) redraws it;
// it opens when there is a pet and closes when there is none.
// onOrder(cmd) sends a pet-cmd.
export function createPetWindow(doc, data, { onOrder } = {}) {
  const root = doc.createElement('div');
  root.className = data ? '' : 'panel pet';
  const W = 260;
  let view = null, last = null;
  const btn = (text, cmd, on) => {
    const b = doc.createElement('button');
    b.textContent = text;
    if (on) b.className = 'on';
    b.addEventListener('click', () => onOrder?.(cmd));
    return b;
  };
  const line = (text, cls) => {
    const d = doc.createElement('div');
    if (cls) d.className = cls;
    d.textContent = text;
    return d;
  };
  const render = () => {
    root.replaceChildren();
    if (!view) { root.style.display = 'none'; return; }
    root.style.display = data ? 'block' : '';
    if (!data) root.classList.add('open');
    const body = doc.createElement('div');
    body.append(line(view.title, 'pet-title'));
    const bar = doc.createElement('div');
    bar.className = 'pet-hp';
    const fill = doc.createElement('div');
    fill.className = 'fill';
    fill.style.width = (view.hp * 100).toFixed(0) + '%';
    bar.append(fill, line(view.hpText, 'nm'));
    body.append(bar);
    const orders = doc.createElement('div');
    orders.className = 'pet-orders';
    for (const [cmd, label] of ORDERS) orders.append(btn(label, cmd, cmd === view.mode));
    body.append(orders);
    const stances = doc.createElement('div');
    stances.className = 'pet-stances';
    for (const s of STANCES) stances.append(btn(s[0].toUpperCase() + s.slice(1), s, s === view.stance));
    body.append(stances);
    if (view.effects.length) {
      const fx = doc.createElement('div');
      fx.className = 'pet-fx';
      for (const r of view.effects) {
        const b = doc.createElement('div');
        b.className = 'fx';
        b.title = r.title;
        b.style.background = r.colour;
        b.textContent = r.label;
        const t = doc.createElement('span');
        t.textContent = r.text;
        b.append(t);
        fx.append(b);
      }
      body.append(fx);
    }
    for (const m of view.minions) body.append(line(m.text, 'pet-minion'));
    if (data) {
      const H = 120 + 16 * view.minions.length + (view.effects.length ? 36 : 0);
      Object.assign(root.style, { position: 'fixed', width: W + 'px', height: H + 'px' });
      Object.assign(body.style, { position: 'absolute', left: '12px', top: '10px', width: W - 24 + 'px', font: '12px sans-serif', color: '#fff', textShadow: '0 1px 1px #000' });
      root.append(frame(doc, data, 'dlg_sm_title_noresize', W, H));
    }
    root.append(body);
  };
  render();
  return {
    el: root,
    get isOpen() { return !!view; },
    update(m, now = 0) { last = m; view = petView(m, now); render(); },
    // refresh redraws with the time left counted down; call it as the seconds pass.
    refresh(now) { if (last && view) { view = petView(last, now); render(); } },
  };
}
