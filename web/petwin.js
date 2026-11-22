// The pet window from the stock MiniPet window: the pet's name, its health bar,
// the stance and movement check boxes, ATTACK and RELEASE, and the pet's
// lasting effects as icons. The server decides what the pet does; the window
// shows it and sends the order.

import { buildWindow } from './uiwin.js';
import { SPELL_ICON } from './charwin.js';

// The check boxes of the window, by control id: which order each one gives and
// which it shows. Go Target and Here have no order on the server.
export const PET_CHECKS = {
  1000: { order: 'aggressive', stance: 'aggressive' },
  1001: { order: 'defensive', stance: 'defensive' },
  1002: { order: 'passive', stance: 'passive' },
  1003: { order: 'follow', mode: 'follow' },
  1004: { order: 'stay', mode: 'stay' },
  1005: { off: true },
  1006: { off: true },
};

// petChecked says whether check box id is ticked for a pet view.
export function petChecked(id, view) {
  const c = PET_CHECKS[id];
  if (!c || !view) return false;
  return (c.stance && view.stance === c.stance) || (c.mode && view.mode === c.mode) || false;
}

// The window shows six effects, as many as it has icon slots.
export const PET_SLOTS = 6;

// petIcons is the icon number for each effect shown, by what it does: the
// spell sheet's picture for a buff or a debuff (icon numbers in the imported
// data do not match the install's sheets, so effects are drawn by kind).
export function petIcons(effects) {
  return effects.slice(0, PET_SLOTS).map((r) => (r.harmful ? SPELL_ICON.debuff : SPELL_ICON.buff));
}

// createStockPet builds the window, hidden until there is a pet.
//   onOrder(cmd)   an order or stance for the pet
export function createStockPet(doc, data, { onOrder, mover } = {}) {
  let view = null, closed = false, last = null;
  const win = buildWindow(doc, data, 'mini_pet', {
    captions: true, labels: true,
    onClose: () => { closed = true; show(); },
    onClick: (id, ev, label) => {
      if (ev === 'PetAttack' || label === 'ATTACK') onOrder?.('attack');
      else if (ev === 'PetRelease' || label === 'RELEASE') onOrder?.('release');
    },
  });
  Object.assign(win.el.style, { position: 'fixed', left: '12px', top: '200px', zIndex: 14, display: 'none' });
  mover?.makeMovable(win.el, 'pet');
  // What a Bonedancer's commander has with it has no place in the stock window: lines under it.
  const minions = doc.createElement('div');
  Object.assign(minions.style, { position: 'absolute', left: '0px', top: '128px', width: '135px', font: '11px sans-serif', color: '#fff', textShadow: '0 1px 1px #000', pointerEvents: 'none' });
  win.el.appendChild(minions);

  const checks = Object.keys(PET_CHECKS).map((id) => ({ id, box: win.check(id) }));
  for (const { id, box } of checks) {
    if (!box) continue;
    if (PET_CHECKS[id].off) { box.disable(true); continue; }
    // A click asks the server; the box keeps to what the server says.
    box.onChange(() => { onOrder?.(PET_CHECKS[id].order); sync(); });
  }
  function sync() {
    for (const { id, box } of checks) if (box) box.set(petChecked(id, view));
  }
  function show() {
    win.el.style.display = view && !closed ? 'block' : 'none';
  }
  const api = {
    el: win.el,
    get isOpen() { return !!view && !closed; },
    update(m, now = 0, v = null) {
      last = m;
      const wasId = view && view.id;
      view = v;
      if (!view) { show(); return; }
      if (wasId !== view.id) { closed = false; mover?.restore('pet'); }
      win.setText('mini_pet_title', view.name);
      win.setBar('mini_pet_life', view.hp);
      sync();
      const icons = petIcons(view.effects);
      for (let i = 0; i < PET_SLOTS; i++) win.setIcon('mini_pet_effect' + i, icons[i] || 0);
      minions.textContent = view.minions.map((x) => x.text).join('\n');
      minions.style.whiteSpace = 'pre';
      show();
    },
    get last() { return last; },
  };
  return api;
}
