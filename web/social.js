// The trade window, the friends list and the duel prompt: plain panels, drawn
// from what the server sends (see docs/protocol.md). The pure parts are here
// for node to test; the panels only lay them out.

export const MAX_OFFER = 10;

// coin is a copper amount as the game writes money: 1 silver = 100 copper,
// 1 gold = 10000.
export function coin(copper) {
  const c = Math.max(0, Math.floor(copper || 0));
  const g = Math.floor(c / 10000), s = Math.floor((c % 10000) / 100), cu = c % 100;
  const parts = [];
  if (g) parts.push(g + 'g');
  if (s) parts.push(s + 's');
  if (cu || !parts.length) parts.push(cu + 'c');
  return parts.join(' ');
}

// tradeView is a trade as the window shows it: two columns and a status line.
export function tradeView(m) {
  if (!m) return null;
  const side = (s) => ({
    items: (s.items || []).map((i) => ({ slot: i.slot, text: i.nm + (i.n > 1 ? ' x' + i.n : ''), title: i.d || '' })),
    money: s.mo || 0, moneyText: coin(s.mo), ok: !!s.ok,
  });
  const mine = side(m.mine), theirs = side(m.theirs);
  let status = 'Add items and money, then accept.';
  if (mine.ok && !theirs.ok) status = 'You have accepted. Waiting for ' + m.with + '.';
  else if (!mine.ok && theirs.ok) status = m.with + ' has accepted.';
  return { title: 'Trade with ' + m.with, mine, theirs, status, full: mine.items.length >= MAX_OFFER };
}

// friendsView is the friends list: who is in the game first.
export function friendsView(m) {
  const list = ((m && m.list) || []).map((f) => ({ name: f.nm, on: !!f.on, text: f.nm + (f.on ? '  (in game)' : '') }));
  list.sort((a, b) => (b.on - a.on) || a.name.localeCompare(b.name));
  return { rows: list, online: list.filter((f) => f.on).length };
}

const el = (doc, tag, cls, text) => {
  const e = doc.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
};

// createTrade builds the trade window, hidden.
//   onOp(op, extra)   a request for the server: remove, money, accept, cancel
export function createTrade(doc, { onOp, mover } = {}) {
  const root = el(doc, 'div', 'panel trade');
  Object.assign(root.style, { position: 'fixed', left: '50%', top: '120px', marginLeft: '-190px', width: '380px', zIndex: 16, display: 'none' });
  if (mover && mover.makeMovable) mover.makeMovable(root, 'trade');
  const title = el(doc, 'h2');
  const cols = el(doc, 'div');
  Object.assign(cols.style, { display: 'flex', gap: '10px' });
  const mkCol = (head) => {
    const c = el(doc, 'div');
    c.style.flex = '1';
    const list = el(doc, 'div', 'items');
    list.style.minHeight = '120px';
    const money = el(doc, 'div');
    c.append(el(doc, 'div', 'dim', head), list, money);
    return { c, list, money };
  };
  const mine = mkCol('You offer'), theirs = mkCol('They offer');
  cols.append(mine.c, theirs.c);
  const moneyRow = el(doc, 'div');
  const input = doc.createElement('input');
  input.type = 'number';
  input.min = '0';
  input.placeholder = 'copper';
  input.style.width = '90px';
  const setMoney = el(doc, 'button', '', 'Set money');
  setMoney.addEventListener('click', () => onOp && onOp('money', { cu: Math.max(0, Math.floor(Number(input.value) || 0)) }));
  moneyRow.append(el(doc, 'span', 'dim', 'Your money: '), input, setMoney);
  const status = el(doc, 'div', 'dim');
  const accept = el(doc, 'button', '', 'Accept');
  const cancel = el(doc, 'button', '', 'Cancel');
  accept.addEventListener('click', () => onOp && onOp('accept'));
  cancel.addEventListener('click', () => onOp && onOp('cancel'));
  const btns = el(doc, 'div');
  btns.append(accept, cancel);
  root.append(title, cols, moneyRow, status, btns);

  let open = false;
  return {
    el: root,
    get isOpen() { return open; },
    show(m) {
      const v = tradeView(m);
      title.textContent = v.title;
      const fill = (col, side, own) => {
        col.list.replaceChildren();
        for (const it of side.items) {
          const row = el(doc, 'div', 'row', it.text);
          row.title = own ? 'Click to take it back' : it.title;
          if (own) {
            row.style.cursor = 'pointer';
            row.addEventListener('click', () => onOp && onOp('remove', { slot: it.slot }));
          }
          col.list.appendChild(row);
        }
        col.money.textContent = 'Money: ' + side.moneyText + (side.ok ? '   (accepted)' : '');
      };
      fill(mine, v.mine, true);
      fill(theirs, v.theirs, false);
      status.textContent = v.status;
      accept.disabled = v.mine.ok;
      open = true;
      root.style.display = 'block';
    },
    hide() { open = false; root.style.display = 'none'; },
  };
}

// createFriends builds the friends window, hidden; toggle() shows or hides it.
//   onAdd(name)   add or remove the named friend
export function createFriends(doc, { onAdd, mover } = {}) {
  const root = el(doc, 'div', 'panel friends');
  Object.assign(root.style, { position: 'fixed', right: '12px', top: '160px', width: '200px', zIndex: 14, display: 'none' });
  if (mover && mover.makeMovable) mover.makeMovable(root, 'friends');
  const title = el(doc, 'h2', '', 'Friends');
  const list = el(doc, 'div', 'items');
  const input = doc.createElement('input');
  input.placeholder = 'name';
  input.style.width = '110px';
  const add = el(doc, 'button', '', 'Add / remove');
  add.addEventListener('click', () => {
    if (input.value.trim()) { onAdd && onAdd(input.value.trim()); input.value = ''; }
  });
  root.append(title, list, input, add);
  let state = null, open = false;
  const draw = () => {
    const v = friendsView(state);
    title.textContent = 'Friends (' + v.online + ' in game)';
    list.replaceChildren(...v.rows.map((r) => {
      const d = el(doc, 'div', 'row', r.text);
      d.style.opacity = r.on ? '1' : '.55';
      return d;
    }));
    if (!v.rows.length) list.textContent = 'No friends yet.';
  };
  return {
    el: root,
    get isOpen() { return open; },
    set(m) { state = m; draw(); },
    toggle() { open = !open; root.style.display = open ? 'block' : 'none'; if (open) draw(); },
  };
}

// createDuelPrompt builds the challenge prompt: a name, Accept and Decline.
//   onAnswer(accept)   the player's answer
export function createDuelPrompt(doc, { onAnswer } = {}) {
  const root = el(doc, 'div', 'panel duelask');
  Object.assign(root.style, { position: 'fixed', left: '50%', top: '80px', marginLeft: '-120px', width: '240px', zIndex: 17, display: 'none', textAlign: 'center' });
  const msg = el(doc, 'div');
  const yes = el(doc, 'button', '', 'Accept');
  const no = el(doc, 'button', '', 'Decline');
  root.append(msg, yes, no);
  let timer = 0;
  const hide = () => { root.style.display = 'none'; clearTimeout(timer); };
  yes.addEventListener('click', () => { hide(); onAnswer && onAnswer(true); });
  no.addEventListener('click', () => { hide(); onAnswer && onAnswer(false); });
  return {
    el: root,
    show(name, seconds = 30) {
      msg.textContent = name + ' challenges you to a duel.';
      root.style.display = 'block';
      clearTimeout(timer);
      timer = setTimeout(hide, seconds * 1000);
    },
    hide,
  };
}
