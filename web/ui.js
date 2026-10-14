// The game's panels in network mode: chat, vitals, the style bar, the pack
// and the loot window. Plain DOM over the canvas; the server decides
// everything these show, so they only display state and send requests.

const SLOT_NAMES = { 40: 'Weapon', 41: 'Shield', 42: 'Torso', 43: 'Legs', 44: 'Arms', 45: 'Feet', 46: 'Hands', 47: 'Head' };
const PACK = 40;
const BAR_KEYS = ['Digit2', 'Digit3', 'Digit4', 'Digit5', 'Digit6', 'Digit7', 'Digit8', 'Digit9', 'Digit0'];

const el = (tag, cls, text) => {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
};

export const money = (c) => {
  const g = Math.floor(c / 10000), s = Math.floor((c % 10000) / 100), cu = c % 100;
  return [g && g + 'g', (g || s) && s + 's', cu + 'c'].filter(Boolean).join(' ');
};

// conColor is how a creature of a level looks to a character of another, as
// the game colours names: grey, green, blue, yellow, orange, red.
export function conColor(mob, me) {
  const d = mob - me;
  if (d <= -9) return '#9aa3ad';
  if (d <= -4) return '#7bd47b';
  if (d <= -1) return '#7fb4ff';
  if (d <= 1) return '#ffe066';
  if (d <= 3) return '#ffa24a';
  return '#ff5a5a';
}

export function createUI(net, { log, onTargetNext } = {}) {
  const root = el('div', 'game-ui');
  document.body.appendChild(root);

  // --- chat ---
  const chat = el('div', 'chat');
  const lines = el('div', 'lines');
  const input = el('input', 'line');
  input.type = 'text'; input.maxLength = 200; input.placeholder = 'Enter to talk · ! to shout · /who';
  input.autocomplete = 'off';
  chat.append(lines, input);
  root.appendChild(chat);
  let typing = false;
  const say = (text, cls = '') => {
    const l = el('div', cls, text);
    lines.appendChild(l);
    while (lines.childElementCount > 9) lines.firstChild.remove();
  };
  net.on.chat = (m) => {
    if (m.ch === 'sys') say(m.text, 'sys');
    else say((m.ch === 'all' ? '[shout] ' : '') + m.from + ': ' + m.text, m.ch === 'all' ? 'all' : '');
  };
  net.on.err = (m) => { say(m.text, 'err'); log?.(m.text, 'info'); };
  const openChat = (prefill = '') => {
    typing = true; chat.classList.add('open'); input.value = prefill; input.focus();
  };
  const closeChat = () => { typing = false; chat.classList.remove('open'); input.blur(); input.value = ''; };
  input.addEventListener('keydown', (e) => {
    e.stopPropagation();
    if (e.key === 'Enter') {
      let t = input.value.trim();
      if (t) {
        if (t.startsWith('!')) net.chat('all', t.slice(1).trim());
        else net.chat('say', t);
      }
      closeChat();
    } else if (e.key === 'Escape') closeChat();
  });
  input.addEventListener('keyup', (e) => e.stopPropagation());

  // --- vitals ---
  const you = document.getElementById('you-box');
  const endBar = el('div', 'bar thin end'); endBar.appendChild(el('i'));
  const xpBar = el('div', 'bar thin xp'); xpBar.appendChild(el('i'));
  const lvl = el('small', 'lvl');
  you.append(endBar, xpBar);
  you.querySelector('.name').prepend(lvl);

  // --- style bar ---
  const bar = el('div', 'stylebar');
  root.appendChild(bar);
  let barSlots = [];
  const renderBar = () => {
    bar.replaceChildren();
    barSlots.forEach((s, i) => {
      const b = el('button', 'style');
      b.dataset.slot = i;
      b.append(el('span', 'key', i < BAR_KEYS.length ? BAR_KEYS[i].replace('Digit', '') : ''), el('span', 'nm', s.nm),
        el('span', 'cost', s.cost + ' end'));
      b.title = s.nm + ' (' + s.opens + ', level ' + s.lv + ')';
      b.addEventListener('click', () => net.style(i));
      bar.appendChild(b);
    });
  };
  net.on.bar = (slots) => { barSlots = slots; renderBar(); };

  // --- pack ---
  const pack = el('div', 'panel pack');
  pack.append(el('h2', '', 'Pack'));
  const worn = el('div', 'worn'), bag = el('div', 'bag'), cash = el('div', 'cash');
  pack.append(worn, bag, cash);
  root.appendChild(pack);
  const renderInv = () => {
    const bySlot = new Map(net.inv.items.map((i) => [i.slot, i]));
    worn.replaceChildren(); bag.replaceChildren();
    for (const s of Object.keys(SLOT_NAMES).map(Number)) {
      const it = bySlot.get(s);
      const c = el('div', 'cell worn' + (it ? ' full' : ''));
      c.append(el('small', '', SLOT_NAMES[s]), el('span', '', it ? it.nm : '—'));
      if (it) {
        c.title = it.nm + (it.d ? ' — ' + it.d : '');
        c.addEventListener('click', () => {
          const free = [...Array(PACK).keys()].find((k) => !bySlot.has(k));
          if (free !== undefined) net.equip(s, free);
        });
      }
      worn.appendChild(c);
    }
    for (let i = 0; i < PACK; i++) {
      const it = bySlot.get(i);
      const c = el('div', 'cell' + (it ? ' full ' + (it.k || '') : ''));
      if (it) {
        c.textContent = it.nm + (it.n > 1 ? ' ×' + it.n : '');
        c.title = it.nm + (it.d ? ' — ' + it.d : '') + '\nclick: equip · shift-click: destroy';
        c.addEventListener('click', (e) => {
          if (e.shiftKey) { net.equip(i, -1); return; }
          if (it.k === 'weapon' || it.k === 'shield' || it.k === 'armour') net.equip(i, -2);
        });
      }
      bag.appendChild(c);
    }
    cash.textContent = 'Money: ' + money(net.inv.mo || 0);
  };
  net.on.inv = renderInv;
  const togglePack = () => { pack.classList.toggle('open'); renderInv(); };

  // --- loot ---
  const lootWin = el('div', 'panel loot');
  root.appendChild(lootWin);
  const renderLoot = (m) => {
    if (!m || (!m.items?.length && !m.mo)) { lootWin.classList.remove('open'); return; }
    lootWin.replaceChildren(el('h2', '', m.nm));
    if (m.mo) {
      const r = el('div', 'row', money(m.mo)); r.addEventListener('click', () => net.loot(m.id, -1)); lootWin.appendChild(r);
    }
    (m.items || []).forEach((it) => {
      const r = el('div', 'row', it.nm + (it.d ? ' — ' + it.d : ''));
      r.addEventListener('click', () => net.loot(m.id, it.slot));
      lootWin.appendChild(r);
    });
    const all = el('button', '', 'Take all');
    all.addEventListener('click', () => {
      if (m.mo) net.loot(m.id, -1);
      (m.items || []).slice().reverse().forEach((it) => net.loot(m.id, it.slot));
    });
    lootWin.append(all);
    lootWin.classList.add('open');
  };
  net.on.corpse = renderLoot;

  // --- death ---
  const dead = el('div', 'death', 'You have died.');
  root.appendChild(dead);

  // --- per frame ---
  const fill = (b, f) => { b.firstElementChild.style.width = (100 * Math.max(0, Math.min(1, f))).toFixed(1) + '%'; };
  function update() {
    const me = net.me;
    if (!me) return;
    fill(endBar, me.end / me.maxEnd);
    fill(xpBar, me.xpNext ? me.xp / me.xpNext : 0);
    lvl.textContent = 'L' + me.level;
    dead.classList.toggle('on', me.dead);
    bar.querySelectorAll('.style').forEach((b) => {
      const i = Number(b.dataset.slot);
      b.classList.toggle('ready', me.avail.includes(i));
      b.classList.toggle('poor', barSlots[i] && me.end < barSlots[i].cost);
    });
    // A corpse window closes when you walk away from it.
    if (lootWin.classList.contains('open') && net.corpse) {
      const v = net.ents.get(net.corpse.id);
      if (!v || Math.hypot(v.x - me.x, v.y - me.y) > 600) { lootWin.classList.remove('open'); net.corpse = null; }
    }
  }

  // Keys the UI owns; returns true when it used the event.
  function key(e) {
    if (typing) return true;
    if (e.code === 'Enter') { openChat(); e.preventDefault(); return true; }
    if (e.code === 'Slash') { openChat('/'); e.preventDefault(); return true; }
    if (e.code === 'KeyI' && !e.repeat) { togglePack(); return true; }
    const i = BAR_KEYS.indexOf(e.code);
    if (i >= 0 && !e.repeat) { net.style(i); return true; }
    return false;
  }

  // The bar, pack and corpse may have arrived before this UI existed.
  if (net.bar.length) net.on.bar(net.bar);
  if (net.inv.items.length) renderInv();
  if (net.corpse) renderLoot(net.corpse);

  return { update, key, say, get typing() { return typing; }, renderLoot, togglePack };
}
