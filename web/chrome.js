// The windows around the game view, from the game's own layouts: the menu bar
// across the top, the compass and the clock.

import { buildWindow } from './uiwin.js';
import { coins } from './pack.js';

// What the menu bar's buttons do. The inventory and the group have windows
// to open; the rest are drawn greyed.
const MENU_EVENTS = { ToggleStatsInventory: 'inventory', ToggleStatsGroup: 'group' };

// createStockMenuBar builds the stock menu bar: the experience bubbles and
// bar, the money, and the row of window buttons.
//   onAction(name)  a button that has a window was pressed ("inventory", "group")
// update(me, money) takes the server's vitals (xp, xpNext) and the money in copper.
export function createStockMenuBar(doc, data, { onAction } = {}) {
  const win = buildWindow(doc, data, 'menu_bar', {
    onClick: (ev) => {
      if (MENU_EVENTS[ev]) onAction?.(MENU_EVENTS[ev]);
    },
    disabled: (ev) => !MENU_EVENTS[ev] && ev !== 'MenuBarToggleExp',
  });
  return {
    el: win.el,
    update(me, money) {
      const f = me && me.xpNext > 0 ? Math.max(0, Math.min(1, me.xp / me.xpNext)) : 0;
      win.setBar('summary_exp_coarse', Math.floor(f * 10) / 10);
      win.setBar('summary_exp_fine', (f * 10) % 1);
      const [, platinum, gold, silver, copper] = coins(money || 0);
      win.setText('merchant_platinum', platinum || '');
      win.setText('merchant_gold', gold || platinum ? gold : '');
      win.setText('merchant_silver', silver || gold || platinum ? silver : '');
      win.setText('merchant_copper', copper);
    },
  };
}

// createStockCompass builds the stock compass: the strip of points, scrolled
// to a heading in degrees clockwise from north.
export function createStockCompass(doc, data) {
  const win = buildWindow(doc, data, 'Compass');
  // The window is 100 square but its control is 128 by 32; the window is the control's size.
  win.resize(128, 32);
  return {
    el: win.el,
    setHeading(deg) {
      win.setHeading(deg);
    },
  };
}

// clockText is a 12-hour time such as "7:05 PM".
export function clockText(date) {
  const h = date.getHours(), m = date.getMinutes();
  return `${h % 12 === 0 ? 12 : h % 12}:${String(m).padStart(2, '0')} ${h < 12 ? 'AM' : 'PM'}`;
}

// createStockClock builds the stock clock. update(date) takes a Date or
// anything with getHours and getMinutes (the game's time of day when a server
// is sending it); with none, it shows the player's own time.
export function createStockClock(doc, data) {
  const win = buildWindow(doc, data, 'Clock');
  let last = '';
  return {
    el: win.el,
    update(date = new Date()) {
      const t = clockText(date);
      if (t !== last) {
        last = t;
        win.setText('time_of_day', t);
      }
    },
  };
}
