// The Command window: the stock window of buttons for things you do rather
// than windows you open -- ATTACK, SPRINT, SIT, PRAY, RELEASE -- and the ones
// this client does not build yet, greyed with the reason.
//
// Without the converted interface a pair of plain buttons (Sit and Sprint)
// stands in, so the postures can still be asked for with the mouse.

import { buildWindow } from './uiwin.js';

// What each stock button asks for; any button not here is greyed.
export const COMMAND_ACTIONS = {
  ToggleAttackMode: 'attack',
  Sprint: 'sprint',
  ToggleSitting: 'sit',
  Pray: 'pray',
  Release: 'release',
  Follow: 'follow',
  Face: 'face',
};

// Why the rest are greyed.
export const COMMAND_WHY = {
  Stick: 'Sticking to a target is not built yet.',
  Bonuses: 'The bonuses list is not built yet.',
  RespecRealm: 'There are no realm abilities to respec.',
};

// createCommandWindow builds the window, stock or plain. onAction(name) gets
// attack, sprint, sit, pray, release, follow or face. toggle() shows or hides it.
export function createCommandWindow(doc, data, { onAction, mover } = {}) {
  if (data && data.window && data.window('command')) return createStockCommand(doc, data, { onAction, mover });
  const root = doc.createElement('div');
  root.className = 'panel command';
  let open = true;
  for (const [label, action] of [['Sit', 'sit'], ['Sprint', 'sprint'], ['Follow', 'follow'], ['Face', 'face']]) {
    const b = doc.createElement('button');
    b.textContent = label;
    b.title = { Sit: 'Sit or stand (X)', Sprint: 'Sprint on or off (Z)', Follow: 'Follow your targeted player', Face: 'Turn to face your target' }[label];
    b.addEventListener('click', () => onAction?.(action));
    root.appendChild(b);
  }
  return {
    el: root,
    get isOpen() { return open; },
    toggle(on = !open) { open = on; root.style.display = open ? '' : 'none'; },
  };
}

function createStockCommand(doc, data, { onAction, mover }) {
  let open = false;
  const win = buildWindow(doc, data, 'command', {
    captions: true, labels: true,
    onClick: (id, ev) => { if (COMMAND_ACTIONS[ev]) onAction?.(COMMAND_ACTIONS[ev]); },
    disabled: (ev) => !COMMAND_ACTIONS[ev],
    tooltip: (ev) => COMMAND_WHY[ev] || '',
  });
  Object.assign(win.el.style, { position: 'fixed', left: '12px', top: '200px', zIndex: 14, display: 'none' });
  mover?.makeMovable(win.el, 'command');
  return {
    el: win.el,
    get isOpen() { return open; },
    toggle(on = !open) { open = on; win.el.style.display = open ? 'block' : 'none'; if (open) mover?.restore('command'); },
  };
}
