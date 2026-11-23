// The guild window: who is in your guild, their ranks, and the buttons for
// what your rank may do. A plain panel drawn from the server's `guild`
// message (docs/protocol.md); the pure parts are here for node to test.

// The permission bits the server sends in `pm`, as player-guilds defines them.
export const PERM = {
  guildSpeak: 1, guildHear: 2, officerSpeak: 4, officerHear: 8, invite: 16, promote: 32, remove: 64, emblem: 128, motd: 256,
};

// can says whether a rank's permission bits include one.
export const can = (pm, bit) => (pm & bit) !== 0;

// guildView is the window's content from a `guild` message, or null when the
// player is in no guild. Members sort by rank then name, or by name or level;
// offline members are left out unless asked for.
export function guildView(m, { sort = 'rank', showOffline = true } = {}) {
  if (!m || !m.nm) return null;
  let rows = (m.mem || []).map((x) => ({
    name: x.nm, rank: x.rk, note: x.nt || '', on: !!x.on, level: x.lv || 0, cls: x.cl || '',
    text: x.nm + '  (rank ' + x.rk + (x.on ? ', L' + x.lv + ' ' + x.cl : ', offline') + ')' + (x.nt ? '  - ' + x.nt : ''),
  }));
  if (!showOffline) rows = rows.filter((r) => r.on);
  const by = {
    rank: (a, b) => a.rank - b.rank || a.name.localeCompare(b.name),
    name: (a, b) => a.name.localeCompare(b.name),
    level: (a, b) => b.level - a.level || a.name.localeCompare(b.name),
  }[sort] || (() => 0);
  rows.sort(by);
  return {
    title: '<' + m.nm + '>', sub: (m.ti || 'Member') + ', rank ' + m.rk,
    motd: m.motd || '', omotd: m.omotd || '', emblem: m.em || 0,
    alliance: m.al ? 'Alliance: ' + m.al : '', rows, online: (m.mem || []).filter((x) => x.on).length, total: (m.mem || []).length,
    mayInvite: can(m.pm, PERM.invite), mayPromote: can(m.pm, PERM.promote), mayRemove: can(m.pm, PERM.remove),
    mayMotd: can(m.pm, PERM.motd), mayEmblem: can(m.pm, PERM.emblem),
  };
}

const el = (doc, tag, cls, text) => {
  const e = doc.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
};

// createGuild builds the window, hidden.
//   onCommand(line)   a chat command for the server, such as "/gc promote Bjorn"
export function createGuild(doc, { onCommand, mover } = {}) {
  const root = el(doc, 'div', 'panel guild');
  Object.assign(root.style, { position: 'fixed', left: '50%', top: '90px', marginLeft: '-190px', width: '380px', zIndex: 15, display: 'none' });
  if (mover && mover.makeMovable) mover.makeMovable(root, 'guild');
  const title = el(doc, 'h2');
  const sub = el(doc, 'div', 'dim');
  const alliance = el(doc, 'div', 'dim');
  const motd = el(doc, 'div');
  const list = el(doc, 'div', 'items');
  Object.assign(list.style, { maxHeight: '200px', overflowY: 'auto' });
  const bar = el(doc, 'div');
  const sortSel = doc.createElement('select');
  for (const [v, t] of [['rank', 'By rank'], ['name', 'By name'], ['level', 'By level']]) {
    const o = doc.createElement('option');
    o.value = v;
    o.textContent = t;
    sortSel.appendChild(o);
  }
  const offline = doc.createElement('input');
  offline.type = 'checkbox';
  offline.checked = true;
  const offlineLabel = el(doc, 'label', 'dim', ' show offline');
  offlineLabel.prepend(offline);
  bar.append(sortSel, offlineLabel);
  const input = doc.createElement('input');
  input.placeholder = 'name or message';
  input.style.width = '150px';
  const btn = (label, fn) => {
    const b = el(doc, 'button', '', label);
    b.addEventListener('click', fn);
    return b;
  };
  let selected = '';
  const actions = el(doc, 'div');
  const send = (line) => onCommand && onCommand(line);
  const invite = btn('Invite', () => input.value.trim() && send('/gc invite ' + input.value.trim()));
  const promote = btn('Promote', () => selected && send('/gc promote ' + selected));
  const demote = btn('Demote', () => selected && send('/gc demote ' + selected));
  const remove = btn('Remove', () => selected && send('/gc remove ' + selected));
  const setMotd = btn('Set message', () => send('/gc motd ' + input.value.trim()));
  const leave = btn('Leave guild', () => send('/gc leave'));
  actions.append(input, invite, promote, demote, remove, setMotd, leave);
  root.append(title, sub, alliance, motd, bar, list, actions);

  let state = null, open = false;
  const draw = () => {
    const v = guildView(state, { sort: sortSel.value, showOffline: offline.checked });
    if (!v) { title.textContent = 'Guild'; sub.textContent = 'You are not in a guild.'; list.replaceChildren(); alliance.textContent = motd.textContent = ''; return; }
    title.textContent = v.title + (v.emblem ? '  (emblem ' + v.emblem + ')' : '');
    sub.textContent = v.sub + '  -  ' + v.online + ' of ' + v.total + ' in game';
    alliance.textContent = v.alliance;
    motd.textContent = v.motd ? 'Message: ' + v.motd : '';
    list.replaceChildren(...v.rows.map((r) => {
      const d = el(doc, 'div', 'row', r.text);
      d.style.cursor = 'pointer';
      d.style.opacity = r.on ? '1' : '.55';
      if (r.name === selected) d.style.background = 'rgba(255,255,255,.12)';
      d.addEventListener('click', () => { selected = r.name; input.value = r.name; draw(); });
      return d;
    }));
    invite.disabled = !v.mayInvite;
    promote.disabled = demote.disabled = !v.mayPromote;
    remove.disabled = !v.mayRemove;
    setMotd.disabled = !v.mayMotd;
  };
  sortSel.addEventListener('change', draw);
  offline.addEventListener('change', draw);
  return {
    el: root,
    get isOpen() { return open; },
    set(m) { state = m; draw(); },
    toggle() { open = !open; root.style.display = open ? 'block' : 'none'; if (open) draw(); },
  };
}
