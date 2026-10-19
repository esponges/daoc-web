// The group: who is in it and how they fare, the invitation, and the chat
// lines of tells and the group channel.
//
// The server sends the group (`group`: members, leader first, each with
// health and endurance in percent, zone and whether in range) and an
// invitation (`groupinv`). These windows only show them and ask: inviting,
// accepting and leaving are requests the server decides.

import { buildWindow, frame, button } from './uiwin.js';

export const GROUP_MAX = 8;

// groupRows turns the group message into rows to draw: a fraction for each
// bar, and a dim row for a member out of range, dead or out of the game.
export function groupRows(members) {
  return (members || []).slice(0, GROUP_MAX).map((m, i) => {
    let where = '';
    if (m.gone) where = 'not in the game';
    else if (m.dead) where = 'dead';
    else if (!m.inRange) where = m.zoneName ? 'in ' + m.zoneName : 'out of range';
    return {
      name: m.name, level: m.level || '', cls: m.class || '', leader: i === 0, you: !!m.you,
      hp: m.gone ? 0 : (m.hp || 0) / 100, end: m.gone ? 0 : (m.end || 0) / 100,
      where, dim: !!where,
      title: m.name + (i === 0 ? ' (leader)' : '') + (where ? ' — ' + where : ''),
    };
  });
}

// chatLine is how a chat message reads in the log, and its kind for colour
// and tab: what a player says, shouts, tells or says to the group, or the
// server's own lines.
export function chatLine(m) {
  switch (m.ch) {
    case 'sys': return { text: m.text, kind: 'sys' };
    case 'all': return { text: '[shout] ' + m.from + ': ' + m.text, kind: 'all' };
    case 'tell':
      return m.to ? { text: 'You send, "' + m.text + '" to ' + m.to + '.', kind: 'tell' }
        : { text: m.from + ' sends, "' + m.text + '"', kind: 'tell' };
    case 'group': return { text: '[Group] ' + m.from + ': "' + m.text + '"', kind: 'group' };
  }
  return { text: m.from + ': ' + m.text, kind: '' };
}

const DIM = 'rgb(140,140,140)';

// createStockGroup builds the stock group window (StatsGroup), and its small
// form (MiniGroup) for the MINI button. update(members) fills them.
//   onInvite()   INVITE was pressed
//   onDisband()  DISBAND was pressed
export function createStockGroup(doc, data, { onInvite, onDisband } = {}) {
  const big = buildWindow(doc, data, 'StatsGroup', {
    labels: true,
    onClick: (id, ev) => {
      if (ev === 'GroupInvite') onInvite?.();
      else if (ev === 'GroupDisband') onDisband?.();
      else if (id === '1002') setMini(true);
    },
    // Only INVITE, DISBAND and MINI have something behind them; FIND, ASSIST,
    // FOLLOW and the rest are greyed.
    disabled: (ev, id) => ev !== 'GroupInvite' && ev !== 'GroupDisband' && id !== '1002',
  });
  const mini = data.windows.has('MiniGroup') ? buildWindow(doc, data, 'MiniGroup') : null;
  const root = doc.createElement('div');
  root.append(big.el);
  if (mini) {
    mini.el.style.display = 'none';
    mini.el.style.cursor = 'pointer';
    mini.el.title = 'Click for the full group window';
    mini.el.addEventListener('click', () => setMini(false));
    root.append(mini.el);
  }
  root.style.display = 'none';
  let small = false;
  const setMini = (on) => {
    small = on && !!mini;
    big.el.style.display = small ? 'none' : 'block';
    if (mini) mini.el.style.display = small ? 'block' : 'none';
  };
  setMini(false);
  return {
    el: root,
    get isOpen() { return root.style.display !== 'none'; },
    toggle(on = !this.isOpen) { root.style.display = on ? 'block' : 'none'; },
    update(members) {
      const rows = groupRows(members);
      for (let i = 0; i < GROUP_MAX; i++) {
        const r = rows[i];
        const colour = r && r.dim ? DIM : undefined;
        big.setText('group_name' + i, r ? (r.leader ? '*' : '') + r.name : '', colour);
        big.setText('group_class' + i, r ? (r.where || r.cls) : '', colour);
        big.setText('group_level' + i, r ? r.level : '', colour);
        // Each member has a red health bar, then endurance and power.
        big.setBarAt(i * 3, r ? r.hp : 0);
        big.setBarAt(i * 3 + 1, r ? r.end : 0);
        big.setBarAt(i * 3 + 2, 0);
        if (mini) {
          mini.setText('group_name' + i, r ? r.name : '', colour);
          mini.setBarAt(i, r ? r.hp : 0);
        }
      }
      if (mini) mini.setText('none', rows.length ? '' : 'No group');
    },
  };
}

// createPlainGroup is the group panel without the converted interface.
export function createPlainGroup(doc, { onInvite, onDisband } = {}) {
  const root = doc.createElement('div');
  root.className = 'panel group';
  const list = doc.createElement('div');
  const buttons = doc.createElement('div');
  const btn = (label, fn) => {
    const b = doc.createElement('button');
    b.textContent = label;
    b.addEventListener('click', () => fn?.());
    return b;
  };
  buttons.append(btn('Invite', onInvite), btn('Disband', onDisband), btn('Close', () => root.classList.remove('open')));
  const h = doc.createElement('h2');
  h.textContent = 'Group';
  root.append(h, list, buttons);
  return {
    el: root,
    get isOpen() { return root.classList.contains('open'); },
    toggle(on = !this.isOpen) { root.classList.toggle('open', on); },
    update(members) {
      const rows = groupRows(members);
      list.replaceChildren();
      if (!rows.length) {
        const r = doc.createElement('div');
        r.className = 'line';
        r.textContent = 'You are not in a group.';
        list.append(r);
      }
      for (const m of rows) {
        const r = doc.createElement('div');
        r.className = 'member' + (m.dim ? ' off' : '');
        r.title = m.title;
        const name = doc.createElement('span');
        name.textContent = (m.leader ? '★ ' : '') + m.name + ' ' + m.level + ' ' + m.cls + (m.where ? ' (' + m.where + ')' : '');
        const bar = doc.createElement('div');
        bar.className = 'bar thin';
        const fillEl = doc.createElement('i');
        fillEl.style.width = (100 * m.hp).toFixed(0) + '%';
        bar.append(fillEl);
        r.append(name, bar);
        list.append(r);
      }
    },
  };
}

// createInvitePrompt asks whether to join a group: on the stock dialog frame
// when the interface is there (data), a plain panel when not. show(from)
// opens it; it closes itself on an answer or after secs.
export function createInvitePrompt(doc, data, { onAccept, onDecline } = {}) {
  const root = doc.createElement('div');
  root.className = data ? '' : 'panel invite';
  const W = 260, H = 76;
  const body = doc.createElement('div');
  if (data) {
    Object.assign(root.style, { position: 'relative', width: W + 'px', height: H + 'px', display: 'none' });
    root.append(frame(doc, data, 'dlg_sm_title_noresize', W, H));
    Object.assign(body.style, { position: 'absolute', left: '12px', top: '10px', width: W - 24 + 'px', font: '12px sans-serif', color: '#fff', textShadow: '0 1px 1px #000' });
  }
  const line = doc.createElement('div');
  body.append(line);
  root.append(body);
  if (data) {
    // The stock buttons, under the line.
    const answer = (fn) => () => { hide(); fn?.(); };
    root.append(button(doc, data, 'button_large', W / 2 - 66, H - 28, answer(onAccept), 'ACCEPT'),
      button(doc, data, 'button_large', W / 2 + 6, H - 28, answer(onDecline), 'DECLINE'));
  } else {
    const buttons = doc.createElement('div');
    buttons.style.marginTop = '10px';
    const btn = (label, fn) => {
      const b = doc.createElement('button');
      b.textContent = label;
      b.addEventListener('click', () => { hide(); fn?.(); });
      return b;
    };
    buttons.append(btn('Accept', onAccept), btn('Decline', onDecline));
    body.append(buttons);
  }
  let timer = null;
  const hide = () => {
    if (data) root.style.display = 'none'; else root.classList.remove('open');
    clearTimeout(timer);
  };
  return {
    el: root,
    get isOpen() { return data ? root.style.display !== 'none' : root.classList.contains('open'); },
    show(from, secs = 60) {
      line.textContent = from + ' has invited you to join a group.';
      if (data) root.style.display = 'block'; else root.classList.add('open');
      clearTimeout(timer);
      timer = setTimeout(hide, secs * 1000);
    },
    hide,
  };
}
