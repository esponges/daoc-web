// The game's panels in network mode: chat, vitals, the style bar, the pack
// and the loot window. Plain DOM over the canvas; the server decides
// everything these show, so they only display state and send requests.

const SLOT_NAMES = { 40: 'Weapon', 41: 'Shield', 42: 'Torso', 43: 'Legs', 44: 'Arms', 45: 'Feet', 46: 'Hands', 47: 'Head' };
const PACK = 40;
const BAR_KEYS = ['Digit2', 'Digit3', 'Digit4', 'Digit5', 'Digit6', 'Digit7', 'Digit8', 'Digit9', 'Digit0'];

import { createStockChat, createStockSummary, createStockBar, createStockRelease } from './hud.js';
import { createStockPack } from './pack.js';
import { createStockLoot } from './loot.js';
import { createStockShop } from './shop.js';
import { createStockMenuBar, createStockCompass, createStockClock } from './chrome.js';
import { sheetView, trainerView, healerText, conColour } from './sheet.js';
import { chatLine, createStockGroup, createPlainGroup, createInvitePrompt } from './group.js';

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

export function createUI(net, { log, onTargetNext, stock, getTarget } = {}) {
  const root = el('div', 'game-ui');
  document.body.appendChild(root);

  // --- chat ---
  // Either the stock chat window, when the converted interface is there, or
  // the plain one below. Both give say / open / close and a typing flag.
  const sendLine = (t) => {
    if (t.startsWith('!')) net.chat('all', t.slice(1).trim());
    else net.chat('say', t);
  };
  let chatApi;
  if (stock) {
    chatApi = createStockChat(document, stock, { onSend: sendLine });
    Object.assign(chatApi.el.style, { position: 'fixed', left: '12px', top: 'auto', bottom: '150px', zIndex: 12 });
    root.appendChild(chatApi.el);
  } else {
    const chat = el('div', 'chat');
    const lines = el('div', 'lines');
    const input = el('input', 'line');
    input.type = 'text'; input.maxLength = 200; input.placeholder = 'Enter to talk · ! to shout · /who';
    input.autocomplete = 'off';
    chat.append(lines, input);
    root.appendChild(chat);
    let isTyping = false;
    const closeChat = () => { isTyping = false; chat.classList.remove('open'); input.blur(); input.value = ''; };
    input.addEventListener('keydown', (e) => {
      e.stopPropagation();
      if (e.key === 'Enter') {
        const t = input.value.trim();
        if (t) sendLine(t);
        closeChat();
      } else if (e.key === 'Escape') closeChat();
    });
    input.addEventListener('keyup', (e) => e.stopPropagation());
    chatApi = {
      say(text, cls = '') {
        lines.appendChild(el('div', cls, text));
        while (lines.childElementCount > 9) lines.firstChild.remove();
      },
      open(prefill = '') { isTyping = true; chat.classList.add('open'); input.value = prefill; input.focus(); },
      close: closeChat,
      get typing() { return isTyping; },
    };
  }
  const say = (text, cls = '') => chatApi.say(text, cls);
  const openChat = (prefill = '') => chatApi.open(prefill);
  net.on.chat = (m) => {
    const l = chatLine(m);
    say(l.text, l.kind);
  };
  net.on.err = (m) => { say(m.text, 'err'); if (!stock) log?.(m.text, 'info'); };

  // --- vitals ---
  // The stock summary window replaces the built-in bars when it is there.
  let summary = null, menuBar = null, compassWin = null, clockWin = null;
  if (stock) {
    summary = createStockSummary(document, stock);
    Object.assign(summary.el.style, { position: 'fixed', right: '12px', left: 'auto', top: '12px', zIndex: 12, pointerEvents: 'none' });
    root.appendChild(summary.el);
    // The menu bar across the top carries the experience and the money, so
    // the separate experience window is not needed.
    menuBar = createStockMenuBar(document, stock, { onAction: (a) => { if (a === 'inventory') togglePack(); else if (a === 'group') toggleGroup(); } });
    Object.assign(menuBar.el.style, { position: 'fixed', left: '50%', top: '0px', marginLeft: '-400px', zIndex: 12 });
    root.appendChild(menuBar.el);
    summary.el.style.top = '34px';
    compassWin = createStockCompass(document, stock);
    Object.assign(compassWin.el.style, { position: 'fixed', left: '50%', top: '34px', marginLeft: '-64px', zIndex: 12, pointerEvents: 'none' });
    root.appendChild(compassWin.el);
    clockWin = createStockClock(document, stock);
    Object.assign(clockWin.el.style, { position: 'fixed', right: '12px', left: 'auto', top: '132px', zIndex: 12, pointerEvents: 'none' });
    root.appendChild(clockWin.el);
    for (const id of ['you-box', 'target-box']) {
      const b = document.getElementById(id);
      if (b) b.style.display = 'none';
    }
  }
  const you = document.getElementById('you-box');
  const endBar = el('div', 'bar thin end'); endBar.appendChild(el('i'));
  const xpBar = el('div', 'bar thin xp'); xpBar.appendChild(el('i'));
  const lvl = el('small', 'lvl');
  you.append(endBar, xpBar);
  you.querySelector('.name').prepend(lvl);

  // --- style bar ---
  // The stock bar when the converted interface is there: a row of icons.
  let stockBar = null;
  const bar = el('div', 'stylebar');
  root.appendChild(bar);
  if (stock) {
    stockBar = createStockBar(document, stock, { onUse: (i) => net.style(i), keys: BAR_KEYS.map((k) => k.replace('Digit', '')) });
    bar.appendChild(stockBar.el);
  }
  let barSlots = [];
  const renderBar = () => {
    if (stockBar) { stockBar.set(barSlots); return; }
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
  // The stock paper doll replaces the plain pack when the interface is there.
  let stockPack = null;
  if (stock) {
    stockPack = createStockPack(document, stock, {
      onMove: (from, to) => net.equip(from, to),
      onUse: (slot) => {
        if (shopOpen()) { net.sell(slot); return; }
        const it = net.inv.items.find((i) => i.slot === slot);
        if (it && (it.k === 'weapon' || it.k === 'shield' || it.k === 'armour')) net.equip(slot, -2);
      },
      onDestroy: (slot) => net.equip(slot, -1),
      onUnequip: (slot) => {
        const free = stockPack.freeSlot();
        if (free >= 0) net.equip(slot, free);
      },
    });
    Object.assign(stockPack.el.style, { position: 'fixed', right: '12px', top: '160px', zIndex: 14 });
    root.appendChild(stockPack.el);
  }
  const renderInv = () => {
    refreshShop();
    if (stockPack) { stockPack.render(net.inv); return; }
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
          if (shopOpen()) { net.sell(i); return; }
          if (it.k === 'weapon' || it.k === 'shield' || it.k === 'armour') net.equip(i, -2);
        });
      }
      bag.appendChild(c);
    }
    cash.textContent = 'Money: ' + money(net.inv.mo || 0);
  };
  net.on.inv = renderInv;
  const togglePack = () => {
    if (stockPack) { stockPack.toggle(); renderInv(); return; }
    pack.classList.toggle('open');
    renderInv();
  };

  // --- loot ---
  const lootWin = el('div', 'panel loot');
  root.appendChild(lootWin);
  // The stock loot window when the interface is there.
  let stockLoot = null;
  if (stock) {
    stockLoot = createStockLoot(document, stock, { onTake: (id, slot) => net.loot(id, slot), formatMoney: money });
    Object.assign(stockLoot.el.style, { position: 'fixed', left: '50%', top: '180px', marginLeft: '-135px', zIndex: 15 });
    root.appendChild(stockLoot.el);
  }
  const renderLoot = (m) => {
    if (stockLoot) { stockLoot.show(m); return; }
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

  // --- shop ---
  // A merchant's window: the stock frame when the interface is there, a plain
  // panel when not. Buying is a click on a row; selling is a click on a pack
  // item while it is open (see the pack handlers above).
  const shopPanel = el('div', 'panel shop');
  root.appendChild(shopPanel);
  let stockShop = null;
  if (stock) {
    stockShop = createStockShop(document, stock, { onBuy: (id) => net.buy(id), onClose: () => hideShop(), formatMoney: money });
    Object.assign(stockShop.el.style, { position: 'fixed', left: '50%', top: '110px', marginLeft: '-150px', zIndex: 15 });
    root.appendChild(stockShop.el);
  }
  const shopOpen = () => (stockShop ? stockShop.isOpen : shopPanel.classList.contains('open'));
  const hideShop = () => {
    if (stockShop) stockShop.hide(); else shopPanel.classList.remove('open');
    net.shop = null;
  };
  const renderPlainShop = (m) => {
    const cash = net.inv.mo || 0;
    shopPanel.replaceChildren(el('h2', '', m.nm));
    for (const it of m.items || []) {
      const r = el('div', 'row', it.nm + (it.d ? ' — ' + it.d : '') + '  ' + money(it.pr || 0));
      if ((it.pr || 0) > cash) r.style.opacity = '.55';
      r.addEventListener('click', () => net.buy(it.id));
      shopPanel.appendChild(r);
    }
    const close = el('button', '', 'Close');
    close.addEventListener('click', hideShop);
    shopPanel.append(close);
    shopPanel.classList.add('open');
  };
  // Prices go red when the character cannot pay, so a change of money redraws.
  const refreshShop = () => {
    if (!net.shop || !shopOpen()) return;
    if (stockShop) stockShop.setMoney(net.inv.mo || 0); else renderPlainShop(net.shop);
  };
  const renderShop = (m) => {
    if (!m) { hideShop(); return; }
    if (stockShop) { stockShop.setMoney(net.inv.mo || 0); stockShop.show(m); } else renderPlainShop(m);
    // Something to sell: bring the pack up beside it.
    if (stockPack && !stockPack.isOpen) togglePack();
  };
  net.on.shop = renderShop;
  net.on.shopclose = hideShop;

  // --- character sheet, trainer, healer ---
  // Plain panels; stock windows for them can come later and fall back to these.
  const sheetPanel = el('div', 'panel sheet');
  const trainPanel = el('div', 'panel trainer');
  const healPanel = el('div', 'panel healer');
  root.append(sheetPanel, trainPanel, healPanel);
  const closeButton = (panel) => {
    const b = el('button', '', 'Close');
    b.addEventListener('click', () => panel.classList.remove('open'));
    return b;
  };
  const renderSheet = () => {
    const v = sheetView(net.stats);
    if (!v) return;
    sheetPanel.replaceChildren(el('h2', '', v.title));
    const grid = el('div', 'stats');
    for (const s of v.stats) {
      const r = el('div', 'stat');
      r.append(el('span', '', s.name), el('b', '', String(s.value)));
      if (s.diff) r.title = 'base ' + s.base + (s.diff > 0 ? ', +' : ', ') + s.diff + ' from levels, items and deaths';
      grid.appendChild(r);
    }
    sheetPanel.append(grid);
    for (const l of v.lines) sheetPanel.append(el('div', 'line', l));
    sheetPanel.append(el('h3', '', 'Specializations (' + v.points + ' points to spend)'));
    for (const s of v.specs) sheetPanel.append(el('div', 'line', s.text));
    sheetPanel.append(el('h3', '', 'Abilities'));
    for (const a of v.abilities) sheetPanel.append(el('div', 'line' + (a.on ? '' : ' off'), a.text + (a.on ? '' : ' (not yet in this game)')));
    sheetPanel.append(closeButton(sheetPanel));
  };
  net.on.stats = () => { if (sheetPanel.classList.contains('open')) renderSheet(); if (trainPanel.classList.contains('open') && net.trainer) renderTrainer(net.trainer); };
  const toggleSheet = () => { sheetPanel.classList.toggle('open'); renderSheet(); };
  const renderTrainer = (m) => {
    const v = trainerView(m);
    trainPanel.replaceChildren(el('h2', '', v.title), el('div', 'line', v.points + ' specialization points to spend'));
    for (const r of v.rows) {
      const row = el('div', 'row' + (r.can ? '' : ' off'));
      row.append(el('span', '', r.text), el('small', '', r.cost + ' pts'));
      row.title = r.hint;
      if (r.can) row.addEventListener('click', () => net.train(r.key, 1));
      trainPanel.appendChild(row);
    }
    trainPanel.append(closeButton(trainPanel));
    trainPanel.classList.add('open');
  };
  net.on.trainer = renderTrainer;
  const renderHealer = (m) => {
    healPanel.replaceChildren(el('h2', '', m.nm), el('div', 'line', healerText(m, money)));
    const ok = el('button', '', 'Accept');
    ok.addEventListener('click', () => { net.restore(m.npc); healPanel.classList.remove('open'); });
    healPanel.append(ok, closeButton(healPanel));
    healPanel.classList.add('open');
  };
  net.on.healer = renderHealer;

  // --- group ---
  // The stock group window and invitation when the interface is there, plain
  // panels when not. INVITE asks for the targeted player, or opens the chat
  // with /invite to type a name.
  const groupHandlers = {
    onInvite: () => {
      const t = getTarget ? getTarget() : null;
      if (t && t.isPlayer) net.groupOp('invite', { id: t.id });
      else openChat('/invite ');
    },
    onDisband: () => net.groupOp('leave'),
  };
  const groupWin = stock ? createStockGroup(document, stock, groupHandlers) : createPlainGroup(document, groupHandlers);
  if (stock) Object.assign(groupWin.el.style, { position: 'fixed', left: '12px', top: '80px', zIndex: 14 });
  root.appendChild(groupWin.el);
  const invite = createInvitePrompt(document, stock, {
    onAccept: () => net.groupOp('accept'),
    onDecline: () => net.groupOp('decline'),
  });
  if (stock) Object.assign(invite.el.style, { position: 'fixed', left: '50%', top: '30%', marginLeft: '-130px', zIndex: 18 });
  root.appendChild(invite.el);
  let grouped = false;
  net.on.group = (members) => {
    groupWin.update(members);
    // Joining a group brings its window up; leaving one puts it away.
    if (!!members.length !== grouped) groupWin.toggle(members.length > 0);
    grouped = members.length > 0;
  };
  net.on.groupinv = (m) => invite.show(m.from, m.secs);
  const toggleGroup = () => { groupWin.update(net.group); groupWin.toggle(); };

  // --- death ---
  const dead = el('div', 'death', 'You have died.');
  root.appendChild(dead);
  // The stock Release window replaces the banner when the interface is there.
  let release = null;
  if (stock) {
    release = createStockRelease(document, stock, { onRelease: () => net.release() });
    Object.assign(release.el.style, { position: 'fixed', left: '50%', top: '22%', marginLeft: '-142px', zIndex: 20 });
    root.appendChild(release.el);
  }

  // --- per frame ---
  const fill = (b, f) => { b.firstElementChild.style.width = (100 * Math.max(0, Math.min(1, f))).toFixed(1) + '%'; };
  function update() {
    const me = net.me;
    if (!me) return;
    if (menuBar) menuBar.update(me, net.inv.mo || 0);
    if (clockWin) clockWin.update();
    if (summary) {
      const t = getTarget ? getTarget() : null;
      summary.update(me, t, t ? (t.isPlayer ? '#9fe3ff' : t.friendly ? '#c8d0da' : conColour(t.e?.con) || conColor(t.level, me.level)) : undefined);
    }
    fill(endBar, me.end / me.maxEnd);
    fill(xpBar, me.xpNext ? me.xp / me.xpNext : 0);
    lvl.textContent = 'L' + me.level;
    if (release) release.update(me.dead, performance.now() / 1000);
    else dead.classList.toggle('on', me.dead);
    if (stockBar) stockBar.update(me);
    bar.querySelectorAll('.style').forEach((b) => {
      const i = Number(b.dataset.slot);
      b.classList.toggle('ready', me.avail.includes(i));
      b.classList.toggle('poor', barSlots[i] && me.end < barSlots[i].cost);
    });
    // A corpse window closes when you walk away from it.
    const lootOpen = stockLoot ? stockLoot.isOpen : lootWin.classList.contains('open');
    if (lootOpen && net.corpse) {
      const v = net.ents.get(net.corpse.id);
      if (!v || Math.hypot(v.x - me.x, v.y - me.y) > 600) {
        if (stockLoot) stockLoot.hide(); else lootWin.classList.remove('open');
        net.corpse = null;
      }
    }
  }

  // Keys the UI owns; returns true when it used the event.
  function key(e) {
    if (chatApi.typing) return true;
    if (e.code === 'Enter') { openChat(); e.preventDefault(); return true; }
    if (e.code === 'Slash') { openChat('/'); e.preventDefault(); return true; }
    if (e.code === 'KeyI' && !e.repeat) { togglePack(); return true; }
    if (e.code === 'KeyC' && !e.repeat && !e.shiftKey) { toggleSheet(); return true; } // shift-C is still the free camera
    const i = BAR_KEYS.indexOf(e.code);
    if (i >= 0 && !e.repeat) { net.style(i); return true; }
    return false;
  }

  // The bar, pack and corpse may have arrived before this UI existed.
  if (net.bar.length) net.on.bar(net.bar);
  if (net.inv.items.length) renderInv();
  if (net.corpse) renderLoot(net.corpse);
  if (net.stats) net.on.stats(net.stats);

  // The compass follows the camera; the viewer gives its heading each frame.
  const setHeading = (deg) => { if (compassWin) compassWin.setHeading(deg); };
  return { update, setHeading, key, say, get typing() { return chatApi.typing; }, renderLoot, togglePack, toggleSheet, toggleGroup };
}
